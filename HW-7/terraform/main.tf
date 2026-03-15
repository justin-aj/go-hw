# HW-7: Order Processing — Terraform root module
#
# Instantiation order (Terraform resolves dependencies automatically):
#   1. network     — VPC, subnets, security groups
#   2. ecr_*       — ECR repositories for each service image
#   3. alb         — Application Load Balancer (public subnets)
#   4. sns_sqs     — SNS topic + SQS queue + subscription
#   5. docker_*    — Build & push images to ECR
#   6. ecs_receiver  — Order Receiver ECS service (connected to ALB)
#   7. ecs_processor — Order Processor ECS service (polls SQS)
#   8. lambda      — Part III serverless processor (SNS trigger)

# ── IAM ──────────────────────────────────────────────────────────────────────
# Reuse the AWS Academy LabRole (same as HW-6).

data "aws_iam_role" "lab_role" {
  name = "LabRole"
}

# ── Network ──────────────────────────────────────────────────────────────────

module "network" {
  source         = "./modules/network"
  service_name   = var.service_name
  container_port = 8080
}

# ── ECR Repositories ─────────────────────────────────────────────────────────

module "ecr_receiver" {
  source          = "./modules/ecr"
  repository_name = "${var.service_name}-receiver"
}

module "ecr_processor" {
  source          = "./modules/ecr"
  repository_name = "${var.service_name}-processor"
}

# ── ALB ──────────────────────────────────────────────────────────────────────

module "alb" {
  source         = "./modules/alb"
  service_name   = var.service_name
  vpc_id         = module.network.vpc_id
  subnet_ids     = module.network.public_subnet_ids
  container_port = 8080
}

# ── SNS + SQS ────────────────────────────────────────────────────────────────

module "sns_sqs" {
  source = "./modules/sns_sqs"
}

# Images are built and pushed manually via docker build/push before terraform apply.

# ── ECS: Order Receiver ───────────────────────────────────────────────────────
# Handles HTTP; connected to ALB.

module "ecs_receiver" {
  source             = "./modules/ecs"
  service_name       = "${var.service_name}-receiver"
  image              = "${module.ecr_receiver.repository_url}:latest"
  container_port     = 8080
  subnet_ids         = module.network.private_subnet_ids
  security_group_ids = [module.network.ecs_security_group_id]
  execution_role_arn = data.aws_iam_role.lab_role.arn
  task_role_arn      = data.aws_iam_role.lab_role.arn
  region             = var.aws_region
  target_group_arn   = module.alb.target_group_arn
  task_count         = 1

  environment_vars = [
    { name = "PORT",          value = "8080" },
    { name = "SNS_TOPIC_ARN", value = module.sns_sqs.topic_arn },
  ]

}

# ── ECS: Order Processor ──────────────────────────────────────────────────────
# Background worker; polls SQS.  No ALB attachment.
# Change num_workers variable (or override here) for Phase 5 experiments.

module "ecs_processor" {
  source             = "./modules/ecs"
  service_name       = "${var.service_name}-processor"
  image              = "${module.ecr_processor.repository_url}:latest"
  container_port     = 0 # no HTTP port — processor doesn't serve traffic
  subnet_ids         = module.network.private_subnet_ids
  security_group_ids = [module.network.ecs_security_group_id]
  execution_role_arn = data.aws_iam_role.lab_role.arn
  task_role_arn      = data.aws_iam_role.lab_role.arn
  region             = var.aws_region
  task_count         = 1

  environment_vars = [
    { name = "SQS_QUEUE_URL", value = module.sns_sqs.queue_url },
    { name = "NUM_WORKERS",   value = tostring(var.num_workers) },
  ]

}

# ── Lambda: Order Processor (Part III) ───────────────────────────────────────
# Build the Lambda zip before running terraform apply.
# See order-lambda/build.sh

module "lambda" {
  source             = "./modules/lambda"
  function_name      = "${var.service_name}-lambda-processor"
  execution_role_arn = data.aws_iam_role.lab_role.arn
  zip_path           = "../order-lambda/processor.zip"
  sns_topic_arn      = module.sns_sqs.topic_arn
}
