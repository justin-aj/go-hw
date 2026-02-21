# Region to deploy into
variable "aws_region" {
  type    = string
  default = "us-east-1"
}

# ECR & ECS settings
variable "ecr_repository_name" {
  type    = string
  default = "product-search"
}

variable "service_name" {
  type    = string
  default = "product-search"
}

variable "container_port" {
  type    = number
  default = 8080
}

variable "ecs_count" {
  type    = number
  default = 1
}

# How long to keep logs
variable "log_retention_days" {
  type    = number
  default = 7
}

# -- ALB (Part III) --
variable "enable_alb" {
  description = "Set to true to create ALB + target group"
  type        = bool
  default     = true
}

# -- Auto Scaling (Part III) --
variable "autoscaling_min" {
  description = "Min ECS tasks (1 for Part II, 2 for Part III)"
  type        = number
  default     = 1
}

variable "autoscaling_max" {
  description = "Max ECS tasks (1 for Part II, 4 for Part III)"
  type        = number
  default     = 1
}

variable "autoscaling_cpu_target" {
  description = "Target CPU % for auto scaling"
  type        = number
  default     = 70
}
