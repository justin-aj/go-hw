output "ecr_repository_url" {
  description = "ECR repo URL for docker push"
  value       = module.ecr.repository_url
}

output "ecs_cluster_name" {
  description = "ECS cluster name"
  value       = module.ecs.cluster_name
}

output "ecs_service_name" {
  description = "ECS service name"
  value       = module.ecs.service_name
}

output "alb_dns_name" {
  description = "ALB DNS - use as Locust host"
  value       = "http://${module.alb.alb_dns_name}"
}
