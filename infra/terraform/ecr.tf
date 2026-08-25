# Container registries for the service images (mirror your CI push target here if
# not using GHCR).

locals {
  images = ["gateway", "evaluator", "ai", "mcp", "console"]
}

resource "aws_ecr_repository" "svc" {
  for_each             = toset(local.images)
  name                 = "${var.project}/${each.key}"
  image_tag_mutability = "MUTABLE"
  image_scanning_configuration {
    scan_on_push = true
  }
}
