variable "cf_user" {}
variable "cf_password" {
  sensitive = true
}
variable "cf_space_prefix" {}
variable "ci_server_token" {
  sensitive = true
}
variable "docker_hub_user" {}
variable "docker_hub_token" {
  sensitive = true
}
variable "developer_emails" {
  type = list(string)
}
variable "worker_egress_allowlist" {
  type = set(string)
}
variable "allow_ssh" {
  type = bool
}
