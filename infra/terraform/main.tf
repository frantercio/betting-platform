terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

variable "region" {
  default = "sa-east-1" # São Paulo — obrigatório para SPA
}

provider "aws" {
  region = var.region
}

# EKS cluster (sa-east-1)
resource "aws_eks_cluster" "bets" {
  name     = "bets-prod"
  version  = "1.30"
  role_arn = aws_iam_role.eks.arn

  vpc_config {
    endpoint_public_access = true
    subnet_ids             = aws_subnet.private[*].id
  }

  depends_on = [
    aws_iam_role_policy_attachment.eks_cluster_policy,
    aws_iam_role_policy_attachment.eks_service_policy
  ]
}

# Variables to be added: vpc, subnets, rds Postgres (sa-east-1),
# secrets in Secrets Manager, ACM certificates, ALB in Brazil.

output "cluster_name" {
  value = aws_eks_cluster.bets.name
}
