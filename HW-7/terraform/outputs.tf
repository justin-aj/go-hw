output "alb_dns_name" {
  description = "Point Locust at this URL"
  value       = "http://${module.alb.dns_name}"
}

output "sns_topic_arn" {
  value = module.sns_sqs.topic_arn
}

output "sqs_queue_url" {
  value = module.sns_sqs.queue_url
}

output "lambda_function_name" {
  value = module.lambda.function_name
}
