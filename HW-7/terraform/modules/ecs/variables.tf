variable "service_name" {
  type = string
}

variable "image" {
  type = string
}

variable "container_port" {
  type    = number
  default = 0
}

variable "subnet_ids" {
  type = list(string)
}

variable "security_group_ids" {
  type = list(string)
}

variable "execution_role_arn" {
  type = string
}

variable "task_role_arn" {
  type = string
}

variable "region" {
  type = string
}

variable "target_group_arn" {
  type    = string
  default = ""
}

variable "task_count" {
  type    = number
  default = 1
}

variable "environment_vars" {
  description = "List of {name, value} environment variable maps for the container"
  type        = list(object({ name = string, value = string }))
  default     = []
}
