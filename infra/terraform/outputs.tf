output "gateway_url" {
  description = "Public URL of the gateway / console (behind the ALB)."
  value       = "http://${aws_lb.main.dns_name}"
}

output "ecr_repositories" {
  description = "ECR repository URLs to push service images to."
  value       = { for k, r in aws_ecr_repository.svc : k => r.repository_url }
}

output "docdb_endpoint" {
  value = aws_docdb_cluster.main.endpoint
}

output "redis_endpoint" {
  value = aws_elasticache_cluster.redis.cache_nodes[0].address
}
