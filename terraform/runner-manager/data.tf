data "cloudfoundry_org" "org" {
  name = var.cf_org_name
}

data "cloudfoundry_service_plans" "cg_service_account" {
  name                  = "space-deployer"
  service_offering_name = "cloud-gov-service-account"
}

# Archive a single file.
data "archive_file" "src" {
  type        = "zip"
  source_dir  = "../../runner-manager"
  output_path = "${path.module}/files/src.zip"
}
