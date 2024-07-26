variable "cf_password" {
  sensitive = true
}
variable "cf_user" {}

variable "cf_org_name" {
  type = string
  default = ""
  description = "Cloud Foundry Organization"
}

variable "cf_space_name" {
  type = string
  default = ""
  description = "Cloud Foundry Space"
}