# flagcast on AWS (Terraform)

Provisions flagcast on **AWS ECS Fargate**: a VPC, an ALB fronting the gateway,
Fargate services for gateway/evaluator/ai/nats (internal DNS via Cloud Map),
**DocumentDB** (Mongo-compatible) and **ElastiCache Redis**, ECR repositories,
and CloudWatch logs.

State is stored in S3 (partial backend config). Provide it at init via a
`backend.hcl` (git-ignored) or `-backend-config` flags:

```hcl
# backend.hcl
bucket         = "your-tf-state-bucket"
key            = "flagcast/terraform.tfstate"
region         = "us-east-1"
dynamodb_table = "your-tf-lock-table"
```

```bash
cd infra/terraform
terraform init -backend-config=backend.hcl
terraform apply -var="docdb_password=<strong-password>" \
  -var="image_tag=<sha>" -var="anthropic_api_key=<key>"
```

ECS services run with container health checks and a deployment circuit breaker
(auto-rollback on a failed rollout).

Outputs the public gateway URL and the ECR repository URLs to push images to.

Notes:
- Images default to `ghcr.io/mralaminahamed/flagcast/*`; set `image_registry` to
  the ECR URLs (from `terraform output`) if pulling from ECR instead.
- The console is served by the gateway/its own static host; the MCP server is a
  stdio process run by an agent, not a Fargate service.
- Follow-ups for production: HTTPS on the ALB (ACM cert + 443 listener), a NAT
  gateway if services move to private subnets, and an NLB for exposing the
  evaluator gRPC port to external SDK clients.
