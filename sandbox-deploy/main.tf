terraform {
  required_version = "~> 1.5"
  required_providers {
    cloudfoundry = {
      source  = "cloudfoundry/cloudfoundry"
      version = "1.5.0"
    }
  }
}
provider "cloudfoundry" {}

module "sandbox-runner" {
  source = "../"

  cf_org_name             = "cloud-gov-devtools-development"
  cf_org_manager          = var.cf_org_manager
  cf_space_prefix         = var.cf_space_prefix
  ci_server_token         = var.ci_server_token
  docker_hub_user         = var.docker_hub_user
  docker_hub_token        = var.docker_hub_token
  manager_instances       = 1
  runner_concurrency      = 10
  developer_emails        = var.developer_emails
  worker_disk_size        = var.worker_disk_size
  program_technologies    = var.program_technologies
  worker_egress_allowlist = var.worker_egress_allowlist
  allow_ssh               = var.allow_ssh
}
