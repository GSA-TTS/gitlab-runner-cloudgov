locals {
  # single flag to turn on and off ssh access to the manager and egress spaces
  allow_ssh       = true
  egress_app_name = "glr-egress-proxy"
  # the list of egress hosts to allow for runner-manager and always needed by runner workers
  devtools_egress_allowlist = [
    "*.fr.cloud.gov",                      # cf-cli calls from manager
    "gsa-0.gitlab-dedicated.us",           # connections from both manager and runners
    "deb.debian.org",                      # runner dependencies install
    "s3.dualstack.us-east-1.amazonaws.com" # gitlab-runner-helper source for runners
  ]
}

# the `depends_on` lines for each resource or module is needed to properly sequence initial creation
# they can be safely removed after initial creation if needed to avoid too many cascading changes,
# but should be put back in place before attempting `terraform destroy` to ensure the destroy
# happens in the proper order as well

module "manager_space" {
  source = "github.com/GSA-TTS/terraform-cloudgov//cg_space?ref=migrate-provider"

  cf_org_name   = var.cf_org_name
  cf_space_name = var.cf_space_name
  allow_ssh     = local.allow_ssh
  deployers     = [var.cf_user]
  developers    = var.developer_emails
}

# temporary method for setting egress rules until terraform provider supports it and cg_space module is updated
data "external" "set-manager-egress" {
  program     = ["/bin/sh", "set_space_egress.sh", "-p", "-t", "-s", module.manager_space.space_name]
  working_dir = path.module
  depends_on  = [module.manager_space]
}

module "worker_space" {
  source = "github.com/GSA-TTS/terraform-cloudgov//cg_space?ref=migrate-provider"

  cf_org_name   = var.cf_org_name
  cf_space_name = "${var.cf_space_name}-workers"
  allow_ssh     = true # manager must be able to cf ssh into workers
  deployers     = [var.cf_user]
  developers    = var.developer_emails
}

module "object_store_instance" {
  source = "github.com/GSA-TTS/terraform-cloudgov//s3?ref=migrate-provider"

  cf_space_id  = module.manager_space.space_id
  name         = var.object_store_instance
  s3_plan_name = "basic-sandbox"
  depends_on   = [module.manager_space]
}

resource "cloudfoundry_service_instance" "runner_service_account" {
  name         = var.service_account_instance
  type         = "managed"
  space        = module.worker_space.space_id
  service_plan = data.cloudfoundry_service_plans.cg_service_account.service_plans.0.id
  tags         = ["gitlab-service-account"]
  depends_on   = [module.worker_space]
}

resource "cloudfoundry_service_key" "runner-service-account-key" {
  provider         = cloudfoundry-community
  name             = "runner-manager-cfapi-key"
  service_instance = cloudfoundry_service_instance.runner_service_account.id
}

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
    PROXY_APP_NAME            = local.egress_app_name
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

module "egress_space" {
  source = "github.com/GSA-TTS/terraform-cloudgov//cg_space?ref=migrate-provider"

  cf_org_name   = var.cf_org_name
  cf_space_name = "${var.cf_space_name}-egress"
  allow_ssh     = local.allow_ssh
  deployers     = [var.cf_user]
  developers    = var.developer_emails
}

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

module "egress_proxy" {
  source = "github.com/GSA-TTS/terraform-cloudgov//egress_proxy?ref=migrate-provider"

  cf_org_name     = var.cf_org_name
  cf_egress_space = module.egress_space.space
  name            = local.egress_app_name
  allowports      = [80, 443, 2222]
  allowlist       = setunion(local.devtools_egress_allowlist, var.worker_egress_allowlist)
  # see egress_proxy/variables.tf for full list of optional arguments
  depends_on = [module.egress_space]
}

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

resource "cloudfoundry_service_instance" "egress-proxy-credentials" {
  name  = "${local.egress_app_name}-credentials"
  space = module.manager_space.space_id
  type  = "user-provided"
  credentials = jsonencode({
    "uri"      = module.egress_proxy.https_proxy
    "http_uri" = module.egress_proxy.http_proxy
  })
}
