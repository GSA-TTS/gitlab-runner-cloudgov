cf_org_name       = ""
cf_space_name     = ""
ci_server_token   = ""
ci_server_url     = "https://gitlab.com/"
default_job_image = "ubuntu:jammy"
# Two executors are supported:
#  custom - Runs jobs in new application instances, deleted after the run.
#  shell - Runs jobs directly on the Runner manager.
runner_executor          = "custom"
runner_name              = ""
runner_memory            = 512
worker_memory            = "512M"
worker_disk_size         = "1G"
service_account_instance = ""
object_store_instance    = ""
