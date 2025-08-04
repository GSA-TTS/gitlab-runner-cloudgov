variable "cf_org_managers" {
  type        = set(string)
  description = "The cloud.gov usernames that may be deploying the root terraform module, must be OrgManagers"
}

variable "developer_emails" {
  type        = set(string)
  description = "cloud.gov accounts to grant SpaceDeveloper access to the manager, worker, and egress spaces"
  default     = []
}

variable "auditor_emails" {
  type        = set(string)
  description = "cloud.gov accounts to grant SpaceAuditor access to the manager, worker, and egress spaces"
  default     = []
}

variable "cf_org_name" {
  type        = string
  default     = "gsa-tts-devtools-prototyping"
  description = "Cloud Foundry Organization"
}

variable "cf_space_prefix" {
  type        = string
  description = "Prefix name for the 3 created spaces"
}

variable "ci_server_token" {
  type        = string
  sensitive   = true
  description = "Gitlab CI Server Token"
}

variable "ci_server_url" {
  type        = string
  default     = "gsa.gitlab-dedicated.us"
  description = "Gitlab Dedicated for Government URL"
}

variable "cf_api_base" {
  type        = string
  default     = "fr.cloud.gov"
  description = "Hostname under which all api and ssh endpoints fall"
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

variable "manager_instances" {
  type        = number
  default     = 2
  description = "Number of manager instances to run"
}

variable "runner_concurrency" {
  type        = number
  default     = 10
  description = "The number of parallel jobs a single manager instance will support"
}

variable "manager_memory" {
  type        = string
  default     = "256M"
  description = "Manager Runner Memory - Unit required (e.g. 512M or 2G)"
}

variable "worker_memory" {
  type        = string
  default     = "768M"
  description = "Worker Memory - Unit required (e.g. 512M or 2G)"
}

variable "worker_disk_size" {
  type        = string
  default     = "3G"
  description = "Worker Disk Size - Unit required (e.g. 512M or 2G)"
}

variable "service_account_instance" {
  type        = string
  default     = "glr-orchestration-bot"
  description = "Service Account Instance"
}

variable "object_store_instance" {
  type        = string
  default     = "glr-dependency-cache"
  description = "S3 Bucket for Gitlab Runner"
}

variable "runner_service_account_key_name" {
  type        = string
  default     = "runner-manager-cfapi-access-key"
  description = "Name of the service account credentials"
}

variable "runner_manager_app_name" {
  type        = string
  default     = "devtools-runner-manager"
  description = "Cloud Foundry App Name for the Runner Manager"
}

variable "egress_app_name" {
  type        = string
  default     = "glr-egress-proxy"
  description = "Cloud Foundry App Name for the Egress Proxy"
}

variable "docker_hub_user" {
  type        = string
  default     = ""
  description = "Docker Hub User"
}

variable "docker_hub_token" {
  type        = string
  default     = ""
  sensitive   = true
  description = "Docker Hub Token"
}

variable "program_technologies" {
  type        = set(string)
  default     = []
  description = "A list of technologies in use by program repositories, to enable known egress endpoints"

  validation {
    condition     = alltrue([for t in var.program_technologies : contains(keys(local.allowlist_map), t)])
    error_message = "program_technologies must be a subset of ${join(", ", keys(local.allowlist_map))}"
  }
}

variable "worker_egress_allowlist" {
  type        = set(string)
  default     = []
  description = "A list of external domain names that runner workers must be able to connect to"
}

variable "worker_egress_ports" {
  type        = list(number)
  default     = [443]
  description = "List of ports that the egress proxy will forward traffic to from the runner workers"
}

variable "worker_egress_https_mode" {
  type        = string
  default     = "https"
  description = "Which egress proxy protocol to send traffic over. Must be http, https, or both"

  validation {
    condition     = contains(["http", "https", "both"], var.worker_egress_https_mode)
    error_message = "worker_egress_https_mode must be one of 'http', 'https', or 'both'"
  }
}

variable "allow_ssh" {
  type        = bool
  default     = false
  description = "Flag for whether ssh access should be allowed to the manager and egress spaces. Should be false for production"
}
