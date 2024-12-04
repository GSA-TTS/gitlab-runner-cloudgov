locals {
  # the list of egress hosts to allow for runner-manager and always needed by runner workers
  devtools_egress_allowlist = [
    "*.fr.cloud.gov",                      # cf-cli calls from manager
    "gsa-0.gitlab-dedicated.us",           # connections from both manager and runners
    "deb.debian.org",                      # debian runner dependencies install
    "*.ubuntu.com",                        # ubuntu runner dependencies install
    "dl-cdn.alpinelinux.org",              # alpine runner dependencies install
    "*.fedoraproject.org",                 # fedora runner dependencies install
    "s3.dualstack.us-east-1.amazonaws.com" # gitlab-runner-helper source for runners
  ]
}

# the `depends_on` lines for each resource or module is needed to properly sequence initial creation
# they can be safely removed after initial creation if needed to avoid too many cascading changes,
# but should be put back in place before attempting `terraform destroy` to ensure the destroy
# happens in the proper order as well

# manager_space: cloud.gov space for running the manager app
module "manager_space" {
  source = "github.com/GSA-TTS/terraform-cloudgov//cg_space?ref=migrate-provider"

  cf_org_name   = var.cf_org_name
  cf_space_name = "${var.cf_space_prefix}-manager"
  allow_ssh     = var.allow_ssh
  deployers     = [var.cf_user]
  developers    = var.developer_emails
}

# worker_space: cloud.gov space for running runner workers and runner services
module "worker_space" {
  source = "github.com/GSA-TTS/terraform-cloudgov//cg_space?ref=migrate-provider"

  cf_org_name   = var.cf_org_name
  cf_space_name = "${var.cf_space_prefix}-workers"
  allow_ssh     = true # manager must be able to cf ssh into workers
  deployers     = [var.cf_user]
  developers    = var.developer_emails
}

# temporary method for setting egress rules until terraform provider supports it and cg_space module is updated
data "external" "set-worker-egress" {
  program     = ["/bin/sh", "set_space_egress.sh", "-t", "-s", module.worker_space.space_name]
  working_dir = path.module
  depends_on  = [module.worker_space]
}

# object_store_instance: s3 bucket for caching build dependencies
module "object_store_instance" {
  source = "github.com/GSA-TTS/terraform-cloudgov//s3?ref=migrate-provider"

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
resource "cloudfoundry_service_key" "runner-service-account-key" {
  provider         = cloudfoundry-community
  name             = "runner-manager-cfapi-key"
  service_instance = cloudfoundry_service_instance.runner_service_account.id
}

# gitlab-runner-manager: the actual runner manager app
resource "cloudfoundry_app" "gitlab-runner-manager" {
  provider          = cloudfoundry-community
  name              = var.runner_manager_app_name
  space             = module.manager_space.space_id
  path              = "${path.module}/files/src.zip"
  source_code_hash  = filesha256("${path.module}/files/src.zip")
  buildpacks        = ["https://github.com/cloudfoundry/apt-buildpack", "binary_buildpack"]
  instances         = 1
  command           = "gitlab-runner run"
  memory            = var.manager_memory
  health_check_type = "process"

  environment = {
    # These are used by .profile
    DEFAULT_JOB_IMAGE = var.default_job_image
    # Following vars are used directly by gitlab-runner register
    # See gitlab-runner register --help for available vars
    CI_SERVER_TOKEN = var.ci_server_token
    CI_SERVER_URL   = var.ci_server_url
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
    #
    # DANGER: Do not set RUNNER_DEBUG to true without reading
    # https://docs.gitlab.com/runner/faq/#enable-debug-logging-mode
    # and ensuring job logs are removed to avoid leaking secrets.
    RUNNER_DEBUG              = "false"
    OBJECT_STORE_INSTANCE     = var.object_store_instance
    PROXY_CREDENTIAL_INSTANCE = cloudfoundry_service_instance.egress-proxy-credentials.name
    PROXY_APP_NAME            = var.egress_app_name
    PROXY_SPACE               = module.egress_space.space_name
    CF_USERNAME               = cloudfoundry_service_key.runner-service-account-key.credentials.username
    CF_PASSWORD               = cloudfoundry_service_key.runner-service-account-key.credentials.password
    DOCKER_HUB_USER           = var.docker_hub_user
    DOCKER_HUB_TOKEN          = var.docker_hub_token
  }
  service_binding {
    service_instance = module.object_store_instance.bucket_id
  }
  service_binding {
    service_instance = cloudfoundry_service_instance.egress-proxy-credentials.id
  }
}

# egress_space: cloud.gov space for running the egress proxy
module "egress_space" {
  source = "github.com/GSA-TTS/terraform-cloudgov//cg_space?ref=migrate-provider"

  cf_org_name   = var.cf_org_name
  cf_space_name = "${var.cf_space_prefix}-egress"
  allow_ssh     = var.allow_ssh
  deployers     = [var.cf_user]
  developers    = var.developer_emails
}

# service-account-egress-role: grant the service account user space_developer in the egress space to
# allow it to set up network policies between runner workers and the egress proxy
resource "cloudfoundry_space_role" "service-account-egress-role" {
  username = cloudfoundry_service_key.runner-service-account-key.credentials.username
  space    = module.egress_space.space_id
  type     = "space_developer"
}

# temporary method for setting egress rules until terraform provider supports it and cg_space module is updated
data "external" "set-proxy-egress" {
  program     = ["/bin/sh", "set_space_egress.sh", "-p", "-s", module.egress_space.space_name]
  working_dir = path.module
  depends_on  = [module.egress_space]
}

# egress_proxy: set up the egress proxy app
module "egress_proxy" {
  source = "github.com/GSA-TTS/terraform-cloudgov//egress_proxy?ref=migrate-provider"

  cf_org_name     = var.cf_org_name
  cf_egress_space = module.egress_space.space
  name            = var.egress_app_name
  allowports      = [80, 443, 2222]
  allowlist       = setunion(local.devtools_egress_allowlist, var.worker_egress_allowlist)
  # see egress_proxy/variables.tf for full list of optional arguments
  depends_on = [module.egress_space]
}

# egress_routing: open up access to the egress proxy from the runner-manager
resource "cloudfoundry_network_policy" "egress_routing" {
  provider = cloudfoundry-community
  policy {
    source_app      = cloudfoundry_app.gitlab-runner-manager.id
    destination_app = module.egress_proxy.app_id
    port            = "61443"
  }

  policy {
    source_app      = cloudfoundry_app.gitlab-runner-manager.id
    destination_app = module.egress_proxy.app_id
    port            = "8080"
  }
}

# egress-proxy-credentials: store the egress proxy credentials in a UPSI for the manager
# to retrieve, use, and pass on to runner workers
resource "cloudfoundry_service_instance" "egress-proxy-credentials" {
  name  = "${var.egress_app_name}-credentials"
  space = module.manager_space.space_id
  type  = "user-provided"
  credentials = jsonencode({
    "https_uri"   = module.egress_proxy.https_proxy
    "http_uri"    = module.egress_proxy.http_proxy
    "cred_string" = "${module.egress_proxy.username}:${module.egress_proxy.password}"
    "domain"      = module.egress_proxy.domain
    "http_port"   = module.egress_proxy.http_port
  })
}
