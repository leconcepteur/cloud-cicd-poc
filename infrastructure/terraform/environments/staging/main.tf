terraform {
  required_version = ">= 1.5.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.0"
    }
  }

  backend "s3" {
    bucket         = "tictactoe-terraform-state"
    key            = "staging/terraform.tfstate"
    region         = "us-east-1"
    dynamodb_table = "tictactoe-terraform-locks"
    encrypt        = true
  }
}

provider "aws" {
  region = "us-east-1"

  default_tags {
    tags = {
      Project     = "tictactoe"
      Environment = "staging"
      ManagedBy   = "terraform"
    }
  }
}

locals {
  name        = "tictactoe-staging"
  environment = "staging"
}

module "vpc" {
  source = "../../modules/vpc"

  name        = local.name
  environment = local.environment
}

module "ecr" {
  source = "../../modules/ecr"

  name        = "tictactoe"
  environment = local.environment
}

module "eks" {
  source = "../../modules/eks"

  name               = local.name
  environment        = local.environment
  vpc_id             = module.vpc.vpc_id
  subnet_ids         = module.vpc.private_subnet_ids
  node_count         = 1 # Minimal for staging
  node_instance_type = "t3.medium"
}

module "rds" {
  source = "../../modules/rds"

  name                       = local.name
  environment                = local.environment
  vpc_id                     = module.vpc.vpc_id
  subnet_ids                 = module.vpc.private_subnet_ids
  allowed_security_group_ids = [] # Will be populated with EKS node security group
  instance_class             = "db.t3.micro"
}

module "elasticache" {
  source = "../../modules/elasticache"

  name                       = local.name
  environment                = local.environment
  vpc_id                     = module.vpc.vpc_id
  subnet_ids                 = module.vpc.private_subnet_ids
  allowed_security_group_ids = [] # Will be populated with EKS node security group
  node_type                  = "cache.t3.micro"
}

output "eks_cluster_endpoint" {
  value = module.eks.cluster_endpoint
}

output "eks_cluster_name" {
  value = module.eks.cluster_name
}

output "rds_endpoint" {
  value = module.rds.endpoint
}

output "redis_endpoint" {
  value = module.elasticache.endpoint
}

output "backend_ecr_url" {
  value = module.ecr.backend_repository_url
}

output "frontend_ecr_url" {
  value = module.ecr.frontend_repository_url
}
