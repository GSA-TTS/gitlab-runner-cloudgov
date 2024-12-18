output "manager_app_id" {
  value = cloudfoundry_app.gitlab-runner-manager.id
}

output "manager_space_id" {
  value = module.manager_space.space_id
}

output "worker_space_id" {
  value = module.worker_space.space_id
}

output "egress_space_id" {
  value = module.egress_space.space_id
}

output "object_cache_service_id" {
  value = module.object_store_instance.bucket_id
}

output "service_account_username" {
  value = local.sa_cf_username
}

output "egress_app_id" {
  value = module.egress_proxy.app_id
}
