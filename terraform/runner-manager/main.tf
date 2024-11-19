locals {

}

module "object_store_instance" {
  source = "github.com/GSA-TTS/terraform-cloudgov//s3?ref=v1.0.0"

  cf_org_name   = var.cf_org_name
  cf_space_name = var.cf_space_name
  name          = var.object_store_instance
  s3_plan_name  = "basic-sandbox"
}

resource "cloudfoundry_app" "gitlab-runner-manager" {
  name              = var.runner_manager_app_name
  space             = data.cloudfoundry_space.space.id
  path              = "${path.module}/files/src.zip"
  buildpacks        = ["https://github.com/cloudfoundry/apt-buildpack", "binary_buildpack"]
  instances         = 1
  command           = "gitlab-runner run"
  memory            = var.runner_memory
  health_check_type = "process"

  environment = {
    # These are used by .profile
    DEFAULT_JOB_IMAGE = var.default_job_image
    # Following vars are used directly by gitlab-runner register
    # See gitlab-runner register --help for available vars
    CI_SERVER_TOKEN = var.ci_server_token
    CI_SERVER_URL   = var.ci_server_url
    RUNNER_EXECUTOR = var.runner_executor
    RUNNER_NAME     = var.runner_name
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
    service_instance = data.cloudfoundry_service_instance.runner_service_account.id
  }
  service_binding {
    service_instance = data.cloudfoundry_service_instance.runner_object_store_instance.id
  }




}
