locals {

}

module "runner_space" {
  source = "github.com/GSA-TTS/terraform-cloudgov//cg_space?ref=migrate-provider"

  cf_org_name   = var.cf_org_name
  cf_space_name = var.cf_space_name
  allow_ssh     = true
  deployers     = [var.cf_user]
  developers    = var.developer_emails
}

# temporary method for setting egress rules until terraform provider supports it and cg_space module is updated
data "external" "set-trusted-egress" {
  program     = ["/bin/sh", "set_space_egress.sh", "-t", "-s", module.runner_space.space_name]
  working_dir = path.module
  depends_on  = [module.runner_space]
}

module "object_store_instance" {
  source = "github.com/GSA-TTS/terraform-cloudgov//s3?ref=migrate-provider"

  cf_space_id  = module.runner_space.space_id
  name         = var.object_store_instance
  s3_plan_name = "basic-sandbox"
  depends_on   = [module.runner_space]
}

resource "cloudfoundry_service_instance" "runner_service_account" {
  name         = var.service_account_instance
  type         = "managed"
  space        = module.runner_space.space_id
  service_plan = data.cloudfoundry_service_plans.cg_service_account.service_plans.0.id
  tags         = ["gitlab-service-account"]
  depends_on   = [module.runner_space]
}

resource "cloudfoundry_app" "gitlab-runner-manager" {
  provider          = cloudfoundry-community
  name              = var.runner_manager_app_name
  space             = module.runner_space.space_id
  path              = "${path.module}/files/src.zip"
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
    RUNNER_DEBUG          = "false"
    OBJECT_STORE_INSTANCE = var.object_store_instance
    DOCKER_HUB_USER       = var.docker_hub_user
    DOCKER_HUB_TOKEN      = var.docker_hub_token
  }
  service_binding {
    service_instance = module.object_store_instance.bucket_id
  }
  service_binding {
    service_instance = cloudfoundry_service_instance.runner_service_account.id
  }
}

module "egress_space" {
  source = "github.com/GSA-TTS/terraform-cloudgov//cg_space?ref=migrate-provider"

  cf_org_name   = var.cf_org_name
  cf_space_name = "${var.cf_space_name}-egress"
  allow_ssh     = true
  deployers     = [var.cf_user]
  developers    = var.developer_emails
}

# temporary method for setting egress rules until terraform provider supports it and cg_space module is updated
data "external" "set-public-egress" {
  program     = ["/bin/sh", "set_space_egress.sh", "-p", "-s", module.egress_space.space_name]
  working_dir = path.module
  depends_on  = [module.egress_space]
}

module "egress_proxy" {
  source = "github.com/GSA-TTS/terraform-cloudgov//egress_proxy?ref=migrate-provider"

  cf_org_name     = var.cf_org_name
  cf_egress_space = module.egress_space.space
  cf_client_space = module.runner_space.space
  name            = "glr-egress-proxy"
  allowlist = {
    (cloudfoundry_app.gitlab-runner-manager.name) = ["api.fr.cloud.gov"]
  }
  # see egress_proxy/variables.tf for full list of optional arguments
  depends_on = [ module.egress_space, module.runner_space, cloudfoundry_app.gitlab-runner-manager ]
}
