data "cloudfoundry_org" "org" {
    name = var.cf_org_name    
}

data "cloudfoundry_space" "space"{
  name = var.cf_space_name
  org = data.cloudfoundry_org.org.id
  }

data "cloudfoundry_service_instance" "runner_service_account" {
  name_or_id = var.service_account_instance
  space = data.cloudfoundry_space.space.id
}

data "cloudfoundry_service_instance" "runner_object_store_instance" {
  name_or_id = var.object_store_instance
  space = data.cloudfoundry_space.space.id
  depends_on = [ module.object_store_instance ]
} 
# Archive a single file.

data "archive_file" "src" {
  type        = "zip"
  source_dir = "../../runner"
  output_path = "${path.module}/files/src.zip"
}