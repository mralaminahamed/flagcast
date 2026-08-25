terraform {
  required_version = ">= 1.6"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.60"
    }
  }
  # Remote state. Partial config — supply bucket/key/region/dynamodb_table at
  # init: `terraform init -backend-config=backend.hcl` (see README). Without a
  # remote backend, CI runs would start from empty local state each time.
  backend "s3" {}
}

provider "aws" {
  region = var.region
  default_tags {
    tags = {
      Project   = var.project
      ManagedBy = "terraform"
    }
  }
}
