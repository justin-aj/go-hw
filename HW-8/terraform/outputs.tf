output "alb_url" {
  description = "Shopping cart service URL — point Locust at this"
  value       = "http://${module.alb.dns_name}"
}

output "ecr_repository_url" {
  description = "Push your Docker image here before terraform apply"
  value       = module.ecr.repository_url
}

output "rds_endpoint" {
  description = "MySQL endpoint (host:port)"
  value       = module.rds.endpoint
}

output "dynamodb_table_name" {
  value = module.dynamodb.table_name
}
