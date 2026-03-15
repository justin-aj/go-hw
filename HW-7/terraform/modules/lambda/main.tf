# Lambda module — Part III
#
# Replaces the SQS+ECS processor with a serverless function.
# The Lambda subscribes DIRECTLY to the SNS topic — no SQS queue needed.
#
# Trade-offs vs ECS worker:
#   + Zero operational overhead (no queue depth monitoring, no worker tuning)
#   + Auto-scales to any concurrency (AWS manages it)
#   + Pay per invocation (FREE within free tier for most startups)
#   - No retry control (SNS retries twice, then drops the message)
#   - Cold starts (~73ms on first invocation or after ~5min idle)
#   - 15-min max execution time (fine for 3s payment processing)

resource "aws_lambda_function" "processor" {
  function_name = var.function_name
  role          = var.execution_role_arn

  # The zip is built from order-lambda/ and uploaded by Terraform.
  filename         = var.zip_path
  source_code_hash = filebase64sha256(var.zip_path)

  # Go on Lambda uses "provided.al2" runtime with a bootstrap binary.
  runtime = "provided.al2"
  handler = "bootstrap"

  memory_size = 512
  timeout     = 30 # >3s processing + buffer

  environment {
    variables = {
      FUNCTION_ENV = "production"
    }
  }
}

# Allow SNS to invoke the Lambda.
resource "aws_lambda_permission" "sns" {
  statement_id  = "AllowSNSInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.processor.function_name
  principal     = "sns.amazonaws.com"
  source_arn    = var.sns_topic_arn
}

# Subscribe the Lambda directly to the SNS topic.
resource "aws_sns_topic_subscription" "lambda_sub" {
  topic_arn = var.sns_topic_arn
  protocol  = "lambda"
  endpoint  = aws_lambda_function.processor.arn
}
