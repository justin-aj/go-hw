output "table_name" {
  description = "DynamoDB table name — inject into ECS as DYNAMO_TABLE"
  value       = aws_dynamodb_table.carts.name
}

output "table_arn" {
  value = aws_dynamodb_table.carts.arn
}
