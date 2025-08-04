variable "cf_org_manager" {
  type        = string
  description = "The OrgManager developer email that is running the sandbox deploy"
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
variable "worker_proxy_https_mode" {
  type    = string
  default = "http"
}
variable "worker_proxy_ports" {
  type    = list(number)
  default = [443, 80]
}
variable "allow_ssh" {
  type = bool
}
