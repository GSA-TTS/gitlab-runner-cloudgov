variable "cf_user" {
  type = string
}
variable "cf_password" {
  type = string
  sensitive = true
}
variable "cf_space_prefix" {
  type = string
}
variable "ci_server_token" {
  type = string
  sensitive = true
}
variable "docker_hub_user" {
  type = string
}
variable "docker_hub_token" {
  type = string
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
