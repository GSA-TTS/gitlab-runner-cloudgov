variable "cf_password" {
  sensitive = true
}
variable "cf_user" {}
variable "developer_emails" {
  type        = list(string)
  description = "cloud.gov accounts to grant SpaceDeveloper access to the runner space and runner egress space"
  default     = []
}

variable "cf_org_name" {
  type        = string
  default     = ""
  description = "Cloud Foundry Organization"
}

variable "cf_space_prefix" {
  type        = string
  default     = ""
  description = "Prefix name for the 3 created spaces"
}

variable "ci_server_token" {
  type        = string
  default     = ""
  sensitive   = true
  description = "Gitlab CI Server Token"
}

variable "ci_server_url" {
  type        = string
  default     = "https://gitlab.com/"
  description = "Gitlab URL"
}

variable "default_job_image" {
  type        = string
  default     = "ubuntu:24.04"
  description = "Default Job Image"
}

# Two executors are supported:
#  custom - Runs jobs in new application instances, deleted after the run.
#  shell - Runs jobs directly on the Runner manager.
variable "runner_executor" {
  type        = string
  default     = "custom"
  description = "Runner Executer"
}

variable "manager_memory" {
  type        = number
  default     = 512
  description = "Manager Runner Memory, given as number of megabytes"
}

variable "worker_memory" {
  type        = string
  default     = "512M"
  description = "Worker Memory - Unit required (e.g. 512M or 2G)"
}

variable "worker_disk_size" {
  type        = string
  default     = "1G"
  description = "Worker Disk Size - Unit required (e.g. 512M or 2G)"
}

variable "service_account_instance" {
  type        = string
  default     = ""
  description = "Service Account Instance"
}

variable "object_store_instance" {
  type        = string
  default     = ""
  description = "S3 Bucket for Gitlab Runner"
}

#Todo: dynamic service bindings
variable "runner_service_bindings" {
  type        = list(object({ service_instance = string }))
  description = "A list of service instances that should be bound to the gitlab runner app"
  default     = []
}

variable "runner_manager_app_name" {
  type        = string
  default     = "devtools-runner-manager"
  description = "Cloud Foundry App Name for the Runner Manager"
}

variable "docker_hub_user" {
  type        = string
  default     = ""
  description = "Docker Hub User"
}

variable "docker_hub_token" {
  type        = string
  default     = ""
  description = "Docker Hub Token"
}

variable "worker_egress_allowlist" {
  type        = set(string)
  default     = []
  description = "A list of external domain names that runner workers must be able to connect to"
}
