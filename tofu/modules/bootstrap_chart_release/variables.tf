variable "name" {
  type = string
}

variable "namespace" {
  type    = string
  default = null
}

variable "chart_path" {
  type    = string
  default = null
}

variable "create_namespace" {
  type    = bool
  default = true
}
