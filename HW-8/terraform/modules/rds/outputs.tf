output "endpoint" {
  description = "RDS endpoint (host:port) — inject into ECS as DB_HOST"
  value       = aws_db_instance.mysql.endpoint
}

output "db_name" {
  value = aws_db_instance.mysql.db_name
}
