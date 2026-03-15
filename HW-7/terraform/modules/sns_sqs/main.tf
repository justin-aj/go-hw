# SNS + SQS module
#
# SNS Topic:  order-processing-events
#   - The Order Receiver publishes here (one Publish call per async order).
#   - Both the SQS queue AND the Lambda (Part III) subscribe to this topic.
#
# SQS Queue:  order-processing-queue
#   - Standard queue (at-least-once delivery; good enough for a lab).
#   - Subscribed to the SNS topic via aws_sns_topic_subscription.
#   - The Order Processor polls this queue.
#
# Key settings (from the HW spec):
#   visibility_timeout_seconds = 30   — message hidden for 30s while a worker
#                                        processes it; if the worker crashes,
#                                        SQS re-delivers after 30s.
#   message_retention_seconds  = 345600 — 4 days
#   receive_wait_time_seconds  = 20   — long polling (processor uses this too)

# ── SNS Topic ────────────────────────────────────────────────────────────────

resource "aws_sns_topic" "orders" {
  name = "order-processing-events"
}

# ── SQS Queue ────────────────────────────────────────────────────────────────

resource "aws_sqs_queue" "orders" {
  name                       = "order-processing-queue"
  visibility_timeout_seconds = 30
  message_retention_seconds  = 345600 # 4 days
  receive_wait_time_seconds  = 20     # long polling default

  tags = { Name = "order-processing-queue" }
}

# Allow SNS to send messages to SQS.
resource "aws_sqs_queue_policy" "sns_to_sqs" {
  queue_url = aws_sqs_queue.orders.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { Service = "sns.amazonaws.com" }
      Action    = "sqs:SendMessage"
      Resource  = aws_sqs_queue.orders.arn
      Condition = {
        ArnEquals = { "aws:SourceArn" = aws_sns_topic.orders.arn }
      }
    }]
  })
}

# Subscribe SQS to SNS.
# raw_message_delivery = false: SNS wraps messages in a JSON envelope.
# The Order Processor unwraps this (the snsWrapper struct in main.go).
resource "aws_sns_topic_subscription" "sqs_sub" {
  topic_arn            = aws_sns_topic.orders.arn
  protocol             = "sqs"
  endpoint             = aws_sqs_queue.orders.arn
  raw_message_delivery = false
}
