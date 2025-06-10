terraform {
  required_version = "~> 1.5"
  required_providers {
    cloudfoundry = {
      source  = "cloudfoundry/cloudfoundry"
      version = "1.6.0"
    }
  }
}
provider "cloudfoundry" {}

module "sandbox-runner" {
  source = "../"

  cg_ssh_host             = "ssh.fr-stage.cloud.gov"
  cf_org_name             = "cloud-gov-devtools-development"
  cf_org_managers         = [var.cf_org_manager]
  cf_space_prefix         = var.cf_space_prefix
  ci_server_token         = var.ci_server_token
  docker_hub_user         = var.docker_hub_user
  docker_hub_token        = var.docker_hub_token
  manager_instances       = 1
  runner_concurrency      = 10
  developer_emails        = var.developer_emails
  worker_disk_size        = var.worker_disk_size
  program_technologies    = var.program_technologies
  worker_egress_allowlist = setunion(["*.fr-stage.cloud.gov"], var.worker_egress_allowlist)
  allow_ssh               = var.allow_ssh
}

locals {
  s3_key   = "developer-local-access-key"
  s3_creds = jsondecode(data.cloudfoundry_service_credential_binding.cache-bucket-key.credential_bindings.0.credential_binding).credentials
}
resource "cloudfoundry_service_credential_binding" "cache-bucket-key" {
  service_instance = module.sandbox-runner.object_cache_service_id
  name             = local.s3_key
  type             = "key"
}
data "cloudfoundry_service_credential_binding" "cache-bucket-key" {
  service_instance = module.sandbox-runner.object_cache_service_id
  name             = local.s3_key
  depends_on       = [cloudfoundry_service_credential_binding.cache-bucket-key]
}
resource "local_sensitive_file" "bucket-creds" {
  filename        = "${path.module}/.shadowenv.d/500_bucket_creds.lisp"
  file_permission = "0600"
  content         = <<EOT
(provide "sandbox-s3-cache-access")
(env/set "AWS_ACCESS_KEY_ID" "${local.s3_creds.access_key_id}")
(env/set "AWS_SECRET_ACCESS_KEY" "${local.s3_creds.secret_access_key}")
(env/set "S3_BUCKET_NAME" "${local.s3_creds.bucket}")
(env/set "AWS_DEFAULT_REGION" "${local.s3_creds.region}")
EOT
}
