data "cloudfoundry_service_plans" "cg_service_account" {
  name                  = "space-deployer"
  service_offering_name = "cloud-gov-service-account"
}

data "cloudfoundry_service_credential_binding" "runner-service-account-key" {
  name             = local.runner_sa_key_name
  service_instance = cloudfoundry_service_instance.runner_service_account.id
  depends_on       = [cloudfoundry_service_credential_binding.runner-service-account-key]
}

# Archive a single file.
data "archive_file" "src" {
  type        = "zip"
  source_dir  = "../../runner-manager"
  output_path = "${path.module}/files/src.zip"
}
