# Sensitive values live in Secrets Manager and are injected into tasks via the
# ECS `secrets` block (valueFrom), never as plaintext task-def environment.

locals {
  secret_values = {
    mongo-uri         = local.mongo_uri
    redis-url         = local.redis_url
    gateway-api-key   = var.gateway_api_key
    evaluator-api-key = var.evaluator_api_key
    anthropic-api-key = var.anthropic_api_key
  }
}

resource "aws_secretsmanager_secret" "app" {
  for_each                = local.secret_values
  name                    = "${var.project}/${each.key}"
  recovery_window_in_days = 0
}

resource "aws_secretsmanager_secret_version" "app" {
  for_each      = local.secret_values
  secret_id     = aws_secretsmanager_secret.app[each.key].id
  secret_string = each.value
}
