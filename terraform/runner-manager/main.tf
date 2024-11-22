locals {
  # single flag to turn on and off ssh access to the manager and egress spaces
  allow_ssh = true
}

# the `depends_on` lines for each resource or module is needed to properly sequence initial creation
# they can be safely removed after initial creation to avoid too many cascading changes, but should be
# put back in place before attempting `terraform destroy` to ensure the destroy happens in the proper order
# as well

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
  space        = module.manager_space.space_id
  service_plan = data.cloudfoundry_service_plans.cg_service_account.service_plans.0.id
  tags         = ["gitlab-service-account"]
  depends_on   = [module.manager_space]
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
    PROXY_CREDENTIAL_INSTANCE = module.egress_proxy.credential_service_name
    DOCKER_HUB_USER           = var.docker_hub_user
    DOCKER_HUB_TOKEN          = var.docker_hub_token
  }
  service_binding {
    service_instance = module.object_store_instance.bucket_id
  }
  service_binding {
    service_instance = cloudfoundry_service_instance.runner_service_account.id
  }
  service_binding {
    service_instance = module.egress_proxy.credential_service_ids[module.manager_space.space_name]
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

# temporary method for setting egress rules until terraform provider supports it and cg_space module is updated
data "external" "set-proxy-egress" {
  program     = ["/bin/sh", "set_space_egress.sh", "-p", "-s", module.egress_space.space_name]
  working_dir = path.module
  depends_on  = [module.egress_space]
}

module "egress_proxy" {
  source = "github.com/GSA-TTS/terraform-cloudgov//egress_proxy?ref=migrate-provider"

  cf_org_name      = var.cf_org_name
  cf_egress_space  = module.egress_space.space
  cf_client_spaces = { (module.manager_space.space_name) = module.manager_space.space_id }
  name             = "glr-egress-proxy"
  allowports       = [443, 2222]
  allowlist = {
    (var.runner_manager_app_name) = [
      "*.fr.cloud.gov",
      "gsa-0.gitlab-dedicated.us"
    ]
  }
  # see egress_proxy/variables.tf for full list of optional arguments
  depends_on = [module.egress_space, module.manager_space]
}

resource "cloudfoundry_network_policy" "egress_routing" {
  provider = cloudfoundry-community
  policy {
    source_app      = cloudfoundry_app.gitlab-runner-manager.id
    destination_app = module.egress_proxy.app_id
    port            = "61443"
  }
}
