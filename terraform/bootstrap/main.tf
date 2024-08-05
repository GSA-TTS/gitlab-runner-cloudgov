locals {
  s3_service_name = "notify-terraform-state"
}

module "s3" {
  source = "github.com/GSA-TTS/terraform-cloudgov//s3?ref=v1.0.0"

  cf_org_name   = "gsa-tts-devtools-prototyping"
  cf_space_name = "anna.mercer"
  name          = local.s3_service_name
}

resource "cloudfoundry_service_key" "bucket_creds" {
  name             = "${local.s3_service_name}-access"
  service_instance = module.s3.bucket_id

  lifecycle {
    prevent_destroy = true
  }
}
