variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "service_name" {
  description = "Base name prefix for all resources"
  type        = string
  default     = "hw7-orders"
}

variable "num_workers" {
  description = "Number of concurrent worker goroutines in the Order Processor"
  type        = number
  default     = 100
}
