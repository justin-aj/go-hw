variable "service_name" {
  description = "Base name used for all resources"
  type        = string
}

variable "container_port" {
  description = "Port the ECS container listens on"
  type        = number
  default     = 8080
}
