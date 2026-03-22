variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "service_name" {
  description = "Base name prefix for all resources"
  type        = string
  default     = "hw8-cart"
}

variable "db_password" {
  description = "MySQL root password — set via TF_VAR_db_password env var, never commit"
  type        = string
  sensitive   = true
}

variable "db_backend" {
  description = "Which storage backend the service uses: 'mysql' or 'dynamo'"
  type        = string
  default     = "mysql"

  validation {
    condition     = contains(["mysql", "dynamo"], var.db_backend)
    error_message = "db_backend must be 'mysql' or 'dynamo'."
  }
}
