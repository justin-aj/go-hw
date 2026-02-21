variable "service_name" {
  type = string
}

variable "cluster_name" {
  type = string
}

variable "min_capacity" {
  type    = number
  default = 1
}

variable "max_capacity" {
  type    = number
  default = 1
}

variable "cpu_target" {
  type    = number
  default = 70
}

variable "scale_in_cooldown" {
  type    = number
  default = 300
}

variable "scale_out_cooldown" {
  type    = number
  default = 300
}
