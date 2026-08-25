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
terraform apply \
  -var="docdb_password=<strong-password>" \
  -var="redis_auth_token=<>=16-char-token>" \
  -var="gateway_api_key=<key>" -var="evaluator_api_key=<key>" \
  -var="anthropic_api_key=<key>" -var="image_tag=<sha>" \
  -var="acm_certificate_arn=<optional ACM arn for HTTPS>"
```

Security posture:
- **Secrets** (Mongo URI, Redis URL, gateway/evaluator/Anthropic keys) live in
  **Secrets Manager**, injected via the ECS `secrets` block — never plaintext env.
- **Encryption:** DocumentDB at rest; ElastiCache Redis at rest + in transit with
  an AUTH token (`rediss://`).
- **TLS:** set `acm_certificate_arn` to add a 443 listener and redirect 80→443.
- **Auth fail-closed:** unset `gateway_api_key`/`evaluator_api_key` means those
  services refuse traffic (no `ALLOW_OPEN_API` in prod), so set them.
- ECS services run container health checks + a deployment circuit breaker
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
