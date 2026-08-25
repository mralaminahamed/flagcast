variable "region" {
  type    = string
  default = "us-east-1"
}

variable "project" {
  type    = string
  default = "flagcast"
}

variable "image_tag" {
  type    = string
  default = "latest"
}

variable "image_registry" {
  description = "Registry hosting the service images (e.g. ghcr.io/mralaminahamed/flagcast)."
  type        = string
  default     = "ghcr.io/mralaminahamed/flagcast"
}

variable "desired_count" {
  type    = number
  default = 1
}

variable "docdb_password" {
  description = "Master password for the DocumentDB (Mongo-compatible) cluster."
  type        = string
  sensitive   = true
}

variable "anthropic_api_key" {
  type      = string
  sensitive = true
  default   = ""
}

variable "gateway_api_key" {
  type      = string
  sensitive = true
  default   = ""
}
