# Wire together modules: network, ecr, logging, alb, ecs, autoscaling

module "network" {
  source         = "./modules/network"
  service_name   = var.service_name
  container_port = var.container_port
}

module "ecr" {
  source          = "./modules/ecr"
  repository_name = var.ecr_repository_name
}

module "logging" {
  source            = "./modules/logging"
  service_name      = var.service_name
  retention_in_days = var.log_retention_days
}

# Reuse existing LabRole for ECS tasks
data "aws_iam_role" "lab_role" {
  name = "LabRole"
}

# -- ALB (always created, used as Locust target) --
module "alb" {
  source         = "./modules/alb"
  service_name   = var.service_name
  vpc_id         = module.network.vpc_id
  subnet_ids     = module.network.subnet_ids
  container_port = var.container_port
}

# -- ECS --
module "ecs" {
  source             = "./modules/ecs"
  service_name       = var.service_name
  image              = "${module.ecr.repository_url}:latest"
  container_port     = var.container_port
  subnet_ids         = module.network.subnet_ids
  security_group_ids = [module.network.security_group_id]
  execution_role_arn = data.aws_iam_role.lab_role.arn
  task_role_arn      = data.aws_iam_role.lab_role.arn
  log_group_name     = module.logging.log_group_name
  ecs_count          = var.ecs_count
  region             = var.aws_region
  target_group_arn   = module.alb.target_group_arn
}

# -- Auto Scaling --
module "autoscaling" {
  source       = "./modules/autoscaling"
  service_name = module.ecs.service_name
  cluster_name = module.ecs.cluster_name
  min_capacity = var.autoscaling_min
  max_capacity = var.autoscaling_max
  cpu_target   = var.autoscaling_cpu_target
}

# -- Docker build & push --
resource "docker_image" "app" {
  name = "${module.ecr.repository_url}:latest"

  build {
    context    = "../"
    dockerfile = "../Dockerfile"
    builder    = "default"
  }
}

resource "docker_registry_image" "app" {
  name = docker_image.app.name
}
