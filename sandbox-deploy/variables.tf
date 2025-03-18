variable "cf_org_manager" {
  type        = string
  description = "The OrgManager developer email that is running the sandbox deploy"
}

variable "cf_user" {
  type        = string
  description = "A regular space developer to log into the community provider"
}
variable "cf_password" {
  type        = string
  sensitive   = true
  description = "The password associated with cf_user to log into the community provider"
}
variable "cf_space_prefix" {
  type = string
}
variable "ci_server_token" {
  type      = string
  sensitive = true
}
variable "docker_hub_user" {
  type = string
}
variable "docker_hub_token" {
  type      = string
  sensitive = true
}
variable "developer_emails" {
  type = list(string)
}
variable "worker_disk_size" {
  type = string
}
variable "program_technologies" {
  type = set(string)
}
variable "worker_egress_allowlist" {
  type = set(string)
}
variable "allow_ssh" {
  type = bool
}
