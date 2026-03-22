# RDS module — MySQL 8.0 in private subnets.
#
# Uses db.t3.micro (AWS Academy eligible).
# Multi-AZ disabled for cost; single AZ is fine for a lab.
# Password is passed in via variable — set via TF_VAR_db_password env var.

resource "aws_db_subnet_group" "main" {
  name       = "${var.service_name}-rds-subnet-group"
  subnet_ids = var.subnet_ids

  tags = { Name = "${var.service_name}-rds-subnet-group" }
}

resource "aws_db_instance" "mysql" {
  identifier             = "${var.service_name}-mysql"
  engine                 = "mysql"
  engine_version         = "8.0"
  instance_class         = "db.t3.micro"
  allocated_storage      = 20
  storage_type           = "gp2"
  db_name                = var.db_name
  username               = var.db_username
  password               = var.db_password
  db_subnet_group_name   = aws_db_subnet_group.main.name
  vpc_security_group_ids = [var.security_group_id]
  skip_final_snapshot    = true
  publicly_accessible    = false
  multi_az               = false

  tags = { Name = "${var.service_name}-mysql" }
}
