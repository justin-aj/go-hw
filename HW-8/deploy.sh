#!/usr/bin/env bash
# deploy.sh — build, push, and apply Terraform for HW-8
#
# Usage:
#   ./deploy.sh mysql    # deploy with MySQL backend
#   ./deploy.sh dynamo   # deploy with DynamoDB backend
#
# Prerequisites:
#   - AWS CLI configured (or AWS Academy credentials active)
#   - Terraform installed
#   - Docker running
#   - TF_VAR_db_password set in environment

set -euo pipefail

BACKEND="${1:-mysql}"
REGION="us-east-1"
SERVICE_DIR="$(dirname "$0")/shopping-cart-service"
TF_DIR="$(dirname "$0")/terraform"

echo "==> Backend: $BACKEND"

# ── 1. Terraform init + apply ─────────────────────────────────────────────────
echo "==> terraform init"
cd "$TF_DIR"
terraform init -input=false

echo "==> terraform apply (db_backend=$BACKEND)"
terraform apply -auto-approve \
  -var "db_backend=$BACKEND"

# ── 2. Grab outputs ───────────────────────────────────────────────────────────
ECR_URL=$(terraform output -raw ecr_repository_url)
ALB_URL=$(terraform output -raw alb_url)
echo "==> ECR: $ECR_URL"
echo "==> ALB: $ALB_URL"

# ── 3. Docker login to ECR ────────────────────────────────────────────────────
echo "==> ECR docker login"
aws ecr get-login-password --region "$REGION" \
  | docker login --username AWS --password-stdin "${ECR_URL%%/*}"

# ── 4. Build + push image ─────────────────────────────────────────────────────
echo "==> docker build"
cd "$SERVICE_DIR"
docker build --platform linux/amd64 -t cart-service:latest .

echo "==> docker tag + push"
docker tag cart-service:latest "${ECR_URL}:latest"
docker push "${ECR_URL}:latest"

# ── 5. Force ECS to pull new image ────────────────────────────────────────────
CLUSTER=$(terraform -chdir="$TF_DIR" output -raw ecr_repository_url | cut -d/ -f1 | cut -d. -f1 || true)
# Simpler: just force a new deployment
SERVICE_NAME="hw8-cart-cart"
CLUSTER_NAME="hw8-cart-cart-cluster"
echo "==> force ECS redeploy"
aws ecs update-service \
  --region "$REGION" \
  --cluster "$CLUSTER_NAME" \
  --service "$SERVICE_NAME" \
  --force-new-deployment \
  --query 'service.serviceName' \
  --output text

echo ""
echo "==> Deploy complete!"
echo "    ALB URL : $ALB_URL"
echo "    Backend : $BACKEND"
echo ""
echo "    Wait ~60s for ECS tasks to become healthy, then run:"
echo "    python test/run_test.py $ALB_URL $BACKEND"
