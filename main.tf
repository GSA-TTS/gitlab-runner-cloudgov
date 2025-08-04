locals {
  # the list of egress hosts to allow for runner-manager and always needed by runner workers
  manager_egress_allowlist = toset([
    var.cg_api_wildcard, # cf-cli calls from manager
    var.ci_server_url
  ])
  technology_allowlist    = flatten([for t in var.program_technologies : local.allowlist_map[t]])
  worker_egress_allowlist = setunion([var.ci_server_url], local.technology_allowlist, var.worker_egress_allowlist)
  cg_space_users          = setunion(var.cf_org_managers, var.developer_emails)
}

# the `depends_on` lines for each resource or module is needed to properly sequence initial creation
# they can be safely removed after initial creation if needed to avoid too many cascading changes,
# but should be put back in place before attempting `terraform destroy` to ensure the destroy
# happens in the proper order as well

# manager_space: cloud.gov space for running the manager app
module "manager_space" {
  source = "github.com/GSA-TTS/terraform-cloudgov//cg_space?ref=b28e89af19aa88d3a6132f6a1d7697bf29accf5f" #v2.3.0

  cf_org_name   = var.cf_org_name
  cf_space_name = "${var.cf_space_prefix}-manager"
  allow_ssh     = var.allow_ssh
  developers    = local.cg_space_users
  auditors      = var.auditor_emails
}

# worker_space: cloud.gov space for running runner workers and runner services
module "worker_space" {
  source = "github.com/GSA-TTS/terraform-cloudgov//cg_space?ref=b28e89af19aa88d3a6132f6a1d7697bf29accf5f" #v2.3.0

  cf_org_name          = var.cf_org_name
  cf_space_name        = "${var.cf_space_prefix}-workers"
  allow_ssh            = true # manager must be able to cf ssh into workers
  developers           = local.cg_space_users
  auditors             = var.auditor_emails
  security_group_names = ["trusted_local_networks_egress"]
}

# object_store_instance: s3 bucket for caching build dependencies
module "object_store_instance" {
  source = "github.com/GSA-TTS/terraform-cloudgov//s3?ref=b28e89af19aa88d3a6132f6a1d7697bf29accf5f" #v2.3.0

  cf_space_id  = module.manager_space.space_id
  name         = var.object_store_instance
  s3_plan_name = "basic-sandbox"
  depends_on   = [module.manager_space]
}

# runner_service_account: service account with permissions for provisioning runner workers and services
resource "cloudfoundry_service_instance" "runner_service_account" {
  name         = var.service_account_instance
  type         = "managed"
  space        = module.worker_space.space_id
  service_plan = data.cloudfoundry_service_plans.cg_service_account.service_plans.0.id
  depends_on   = [module.worker_space]
}

# runner-service-account-key: the actual username & password for the service account user
# needed to pass into the manager and to assign space_developer in the egress space
resource "cloudfoundry_service_credential_binding" "runner-service-account-key" {
  name             = var.runner_service_account_key_name
  service_instance = cloudfoundry_service_instance.runner_service_account.id
  type             = "key"
}
locals {
  sa_bot_credentials = jsondecode(data.cloudfoundry_service_credential_binding.runner-service-account-key.credential_bindings.0.credential_binding).credentials
  sa_cf_username     = nonsensitive(local.sa_bot_credentials.username)
  sa_cf_password     = sensitive(local.sa_bot_credentials.password)
}

# gitlab-runner-manager: the actual runner manager app
resource "cloudfoundry_app" "gitlab-runner-manager" {
  name              = var.runner_manager_app_name
  space_name        = module.manager_space.space_name
  org_name          = var.cf_org_name
  path              = data.archive_file.src.output_path
  source_code_hash  = data.archive_file.src.output_base64sha256
  buildpacks        = ["https://github.com/cloudfoundry/apt-buildpack", "binary_buildpack"]
  instances         = var.manager_instances
  strategy          = "rolling"
  command           = "gitlab-runner run"
  no_route          = true
  memory            = var.manager_memory
  health_check_type = "process"
  enable_ssh        = var.allow_ssh

  environment = {
    # These are used by .profile
    DEFAULT_JOB_IMAGE = var.default_job_image
    # Following vars are used directly by gitlab-runner register
    # See gitlab-runner register --help for available vars
    CI_SERVER_TOKEN = var.ci_server_token
    CI_SERVER_URL   = "https://${var.ci_server_url}"
    RUNNER_EXECUTOR = var.runner_executor
    RUNNER_NAME     = var.runner_manager_app_name
    # Following vars are for tuning worker defaults
    WORKER_MEMORY    = var.worker_memory
    WORKER_DISK_SIZE = var.worker_disk_size
    WORKER_SPACE     = module.worker_space.space_name
    # Remaining runner configuration is generally static. In order to surface
    # the entire configuration input, we are using envvars for all of it.
    RUNNER_BUILDS_DIR        = "/tmp/build"
    RUNNER_CACHE_DIR         = "/tmp/cache"
    CUSTOM_CLEANUP_EXEC      = "/home/vcap/app/cf-driver/cleanup.sh"
    CUSTOM_PREPARE_EXEC      = "/home/vcap/app/cf-driver/prepare.sh"
    CUSTOM_RUN_EXEC          = "/home/vcap/app/cf-driver/run.sh"
    REGISTER_NON_INTERACTIVE = true
    # TODO - Add timeouts like CUSTOM_CLEANUP_EXEC_TIMEOUT
    RUNNER_CONCURRENCY    = var.runner_concurrency
    OBJECT_STORE_INSTANCE = var.object_store_instance
    PROXY_APP_NAME        = var.egress_app_name
    PROXY_SPACE           = module.egress_space.space_name
    CF_USERNAME           = local.sa_cf_username
    CF_PASSWORD           = local.sa_cf_password
    CG_SSH_HOST           = var.cg_ssh_host
    DOCKER_HUB_USER       = var.docker_hub_user
    DOCKER_HUB_TOKEN      = var.docker_hub_token
    # DANGER: Do not set RUNNER_DEBUG to true without reading
    # https://docs.gitlab.com/runner/faq/#enable-debug-logging-mode
    # and ensuring job logs are removed to avoid leaking secrets.
    RUNNER_DEBUG                 = "false"
    RUNNER_DEBUG_USERS           = join(" ", var.developer_emails)
    CUSTOM_ENV_PRESERVE_WORKER   = "false"
    CUSTOM_ENV_PRESERVE_SERVICES = "false"
  }
  service_bindings = [
    { service_instance = var.object_store_instance },
    { service_instance = cloudfoundry_service_instance.manager-egress-credentials.name }
  ]
  depends_on = [
    module.object_store_instance,
    cloudfoundry_service_instance.manager-egress-credentials
  ]
  lifecycle {
    replace_triggered_by = [cloudfoundry_service_instance.manager-egress-credentials.credentials]
  }
}

# egress_space: cloud.gov space for running the egress proxy
module "egress_space" {
  source = "github.com/GSA-TTS/terraform-cloudgov//cg_space?ref=b28e89af19aa88d3a6132f6a1d7697bf29accf5f" #v2.3.0

  cf_org_name          = var.cf_org_name
  cf_space_name        = "${var.cf_space_prefix}-egress"
  allow_ssh            = var.allow_ssh
  developers           = local.cg_space_users
  auditors             = var.auditor_emails
  security_group_names = ["public_networks_egress"]
}

# service-account-egress-role: grant the service account user space_developer in the egress space to
# allow it to set up network policies between runner workers and the egress proxy
resource "cloudfoundry_space_role" "service-account-egress-role" {
  username = local.sa_cf_username
  space    = module.egress_space.space_id
  type     = "space_developer"
}

# egress_proxy: set up the egress proxy app
module "egress_proxy" {
  source = "github.com/gsa-tts/cg-egress-proxy?ref=f0856cbf84af96b1ce6088a49c71cd8aa414b9b3"

  cf_org_name     = var.cf_org_name
  cf_egress_space = module.egress_space.space
  name            = var.egress_app_name
  client_configuration = {
    "wsr-manager" = {
      ports     = [443, 2222]
      allowlist = local.manager_egress_allowlist
    }
    "wsr-worker" = {
      ports     = [80, 443]
      allowlist = local.worker_egress_allowlist
    }
  }
  # see egress_proxy/variables.tf for full list of optional arguments
  depends_on = [module.egress_space]
}

# egress_routing: open up access to the egress proxy from the runner-manager
resource "cloudfoundry_network_policy" "egress_routing" {
  policies = [
    {
      source_app      = cloudfoundry_app.gitlab-runner-manager.id
      destination_app = module.egress_proxy.app_id
      port            = module.egress_proxy.https_port
    },
    {
      source_app      = cloudfoundry_app.gitlab-runner-manager.id
      destination_app = module.egress_proxy.app_id
      port            = module.egress_proxy.http_port
    }
  ]
}

# manager-egress-credentials: store the egress proxy credentials in a UPSI for the manager
resource "cloudfoundry_service_instance" "manager-egress-credentials" {
  name  = "manager-egress-credentials"
  space = module.manager_space.space_id
  type  = "user-provided"
  credentials = jsonencode({
    https_uri   = module.egress_proxy.https_proxy["wsr-manager"]
    http_uri    = module.egress_proxy.https_proxy["wsr-manager"]
    cred_string = "${module.egress_proxy.username["wsr-manager"]}:${module.egress_proxy.password["wsr-manager"]}"
    domain      = module.egress_proxy.domain
    http_port   = module.egress_proxy.http_port
  })
  depends_on = [module.manager_space]
}
moved {
  from = cloudfoundry_service_instance.egress-proxy-credentials
  to   = cloudfoundry_service_instance.manager-egress-credentials
}


# worker-egress-credentials: store the egress proxy credentials in a UPSI for the workers
resource "cloudfoundry_service_instance" "worker-egress-credentials" {
  name  = "worker-egress-credentials"
  space = module.worker_space.space_id
  type  = "user-provided"
  credentials = jsonencode({
    http_uri = module.egress_proxy.http_proxy["wsr-worker"]
  })
  depends_on = [module.worker_space]
}
