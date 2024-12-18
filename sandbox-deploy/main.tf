terraform {
  required_version = "~> 1.5"
  required_providers {
    cloudfoundry = {
      source  = "cloudfoundry/cloudfoundry"
      version = "1.1.0"
    }
    cloudfoundry-community = {
      source  = "cloudfoundry-community/cloudfoundry"
      version = "0.53.1"
    }
  }
}
provider "cloudfoundry" {
  api_url  = "https://api.fr.cloud.gov"
  user     = var.cf_user
  password = var.cf_password
}
provider "cloudfoundry-community" {
  api_url  = "https://api.fr.cloud.gov"
  user     = var.cf_user
  password = var.cf_password
}

module "sandbox-runner" {
  source = "../"

  cf_user                 = var.cf_user
  cf_space_prefix         = var.cf_space_prefix
  ci_server_token         = var.ci_server_token
  docker_hub_user         = var.docker_hub_user
  docker_hub_token        = var.docker_hub_token
  developer_emails        = var.developer_emails
  worker_egress_allowlist = var.worker_egress_allowlist
  allow_ssh               = var.allow_ssh
}
