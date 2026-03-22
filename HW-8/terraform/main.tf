# HW-8: Shopping Cart Service — Terraform root module
#
# Architecture:
#   network     → VPC, subnets, security groups (ECS + RDS)
#   ecr         → ECR repository for the shopping-cart image
#   alb         → Application Load Balancer (public subnets)
#   rds         → MySQL 8.0 on db.t3.micro (private subnets)
#   dynamodb    → DynamoDB carts table (PAY_PER_REQUEST)
#   ecs         → Shopping Cart ECS Fargate service
#
# Switch backends by setting var.db_backend = "mysql" or "dynamo".

# ── IAM ──────────────────────────────────────────────────────────────────────
# Create an ECS task execution + task role since this is an AWS Sandbox
# (no pre-existing LabRole).

data "aws_iam_policy_document" "ecs_assume" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["ecs-tasks.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "ecs_role" {
  name               = "${var.service_name}-ecs-role"
  assume_role_policy = data.aws_iam_policy_document.ecs_assume.json
}

# Execution role: pull images from ECR, write logs to CloudWatch
resource "aws_iam_role_policy_attachment" "ecs_execution" {
  role       = aws_iam_role.ecs_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

# Task role: allow the running container to call DynamoDB + RDS
resource "aws_iam_role_policy" "ecs_task_policy" {
  name = "${var.service_name}-task-policy"
  role = aws_iam_role.ecs_role.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow"
        Action   = ["dynamodb:PutItem", "dynamodb:GetItem", "dynamodb:UpdateItem"]
        Resource = module.dynamodb.table_arn
      }
    ]
  })
}

# ── Network ──────────────────────────────────────────────────────────────────

module "network" {
  source         = "./modules/network"
  service_name   = var.service_name
  container_port = 8080
}

# ── ECR ──────────────────────────────────────────────────────────────────────

module "ecr" {
  source          = "./modules/ecr"
  repository_name = "${var.service_name}-cart"
}

# ── ALB ──────────────────────────────────────────────────────────────────────

module "alb" {
  source         = "./modules/alb"
  service_name   = var.service_name
  vpc_id         = module.network.vpc_id
  subnet_ids     = module.network.public_subnet_ids
  container_port = 8080
}

# ── RDS (MySQL) ───────────────────────────────────────────────────────────────

module "rds" {
  source            = "./modules/rds"
  service_name      = var.service_name
  subnet_ids        = module.network.private_subnet_ids
  security_group_id = module.network.rds_security_group_id
  db_password       = var.db_password
}

# ── DynamoDB ──────────────────────────────────────────────────────────────────

module "dynamodb" {
  source       = "./modules/dynamodb"
  service_name = var.service_name
}

# ── ECS: Shopping Cart Service ────────────────────────────────────────────────
# DB_BACKEND selects between mysql and dynamo at runtime.
# Both connection strings are always injected; only one is used.

module "ecs_cart" {
  source             = "./modules/ecs"
  service_name       = "${var.service_name}-cart"
  image              = "${module.ecr.repository_url}:latest"
  container_port     = 8080
  subnet_ids         = module.network.private_subnet_ids
  security_group_ids = [module.network.ecs_security_group_id]
  execution_role_arn = aws_iam_role.ecs_role.arn
  task_role_arn      = aws_iam_role.ecs_role.arn
  region             = var.aws_region
  target_group_arn   = module.alb.target_group_arn
  task_count         = 1

  environment_vars = [
    { name = "PORT",          value = "8080" },
    { name = "DB_BACKEND",    value = var.db_backend },
    { name = "DB_HOST",       value = split(":", module.rds.endpoint)[0] },
    { name = "DB_PORT",       value = "3306" },
    { name = "DB_NAME",       value = module.rds.db_name },
    { name = "DB_USER",       value = "admin" },
    { name = "DB_PASSWORD",   value = var.db_password },
    { name = "DYNAMO_TABLE",  value = module.dynamodb.table_name },
    { name = "AWS_REGION",    value = var.aws_region },
  ]
}
