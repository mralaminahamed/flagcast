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

variable "redis_auth_token" {
  description = "ElastiCache Redis AUTH token (>=16 chars; required for transit encryption)."
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

variable "evaluator_api_key" {
  type      = string
  sensitive = true
  default   = ""
}

variable "acm_certificate_arn" {
  description = "ACM cert ARN to enable HTTPS on the ALB. Empty = HTTP only."
  type        = string
  default     = ""
}
