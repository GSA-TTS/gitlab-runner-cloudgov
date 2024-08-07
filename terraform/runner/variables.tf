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

variable "ci_server_token" {
  type = string
  default = ""
  sensitive = true
  description = "Gitlab CI Server Token"
}

variable "ci_server_url" {
  type = string
  default = "https://gitlab.com/"
  description = "Gitlab URL"
}

variable "default_job_image" {
  type = string
  default = "ubuntu:jammy"
  description = "Default Job Image"
}

# Two executors are supported:
#  custom - Runs jobs in new application instances, deleted after the run.
#  shell - Runs jobs directly on the Runner manager.
variable "runner_executor" {
  type = string
  default = "custom"
  description = "Runner Executer"
}

variable "runner_name" {
  type = string
  default = "gitlab-runner"
  description = "Cloud Foundry Organization"
}

variable "runner_memory" {
  type = number
  default = 512
  description = "Cloud Foundry Organization"
}

variable "worker_memory" {
  type = string
  default = ""
  description = "Cloud Foundry Organization"
}

variable "service_account_instance" {
  type = string
  default = ""
  description = "Service Account Instance"
}

variable "object_store_instance" {
  type = string
  default = ""
  description = "Cloud Foundry Organization"
}

#Todo: dynamic service bindings
variable "runner_service_bindings" {
  type        = list(object({ service_instance = string }))
  description = "A list of service instances that should be bound to the thanos app"
  default     = []
}
