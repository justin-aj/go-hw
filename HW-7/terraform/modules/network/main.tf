# Network module — VPC, subnets, IGW, NAT, route tables
#
# Public subnets (10.0.1/2.0/24):  ALB lives here (internet-facing)
# Private subnets (10.0.10/11.0/24): ECS tasks live here (no direct inbound)
#
# NAT Gateway lets private-subnet ECS tasks reach the internet (ECR, SNS, SQS)
# without being reachable from the internet themselves.

data "aws_availability_zones" "available" {
  state = "available"
}

# ── VPC ──────────────────────────────────────────────────────────────────────

resource "aws_vpc" "main" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = { Name = "${var.service_name}-vpc" }
}

# ── Internet Gateway (public subnets → internet) ─────────────────────────────

resource "aws_internet_gateway" "igw" {
  vpc_id = aws_vpc.main.id
  tags   = { Name = "${var.service_name}-igw" }
}

# ── Public Subnets ───────────────────────────────────────────────────────────

resource "aws_subnet" "public" {
  count             = 2
  vpc_id            = aws_vpc.main.id
  cidr_block        = ["10.0.1.0/24", "10.0.2.0/24"][count.index]
  availability_zone = data.aws_availability_zones.available.names[count.index]

  map_public_ip_on_launch = true
  tags                    = { Name = "${var.service_name}-public-${count.index + 1}" }
}

# ── Private Subnets ──────────────────────────────────────────────────────────

resource "aws_subnet" "private" {
  count             = 2
  vpc_id            = aws_vpc.main.id
  cidr_block        = ["10.0.10.0/24", "10.0.11.0/24"][count.index]
  availability_zone = data.aws_availability_zones.available.names[count.index]

  tags = { Name = "${var.service_name}-private-${count.index + 1}" }
}

# ── NAT Gateway (private subnets → internet via public subnet) ───────────────
# One NAT in the first public subnet is enough for a lab.

resource "aws_eip" "nat" {
  domain = "vpc"
  tags   = { Name = "${var.service_name}-nat-eip" }
}

resource "aws_nat_gateway" "nat" {
  allocation_id = aws_eip.nat.id
  subnet_id     = aws_subnet.public[0].id
  tags          = { Name = "${var.service_name}-nat" }

  depends_on = [aws_internet_gateway.igw]
}

# ── Route Tables ─────────────────────────────────────────────────────────────

resource "aws_route_table" "public" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.igw.id
  }

  tags = { Name = "${var.service_name}-public-rt" }
}

resource "aws_route_table_association" "public" {
  count          = 2
  subnet_id      = aws_subnet.public[count.index].id
  route_table_id = aws_route_table.public.id
}

resource "aws_route_table" "private" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block     = "0.0.0.0/0"
    nat_gateway_id = aws_nat_gateway.nat.id
  }

  tags = { Name = "${var.service_name}-private-rt" }
}

resource "aws_route_table_association" "private" {
  count          = 2
  subnet_id      = aws_subnet.private[count.index].id
  route_table_id = aws_route_table.private.id
}

# ── Security Group (ECS tasks) ───────────────────────────────────────────────
# Allows inbound on container_port from ALB, and all outbound (for SNS/SQS/ECR).

resource "aws_security_group" "ecs" {
  name   = "${var.service_name}-ecs-sg"
  vpc_id = aws_vpc.main.id

  ingress {
    from_port   = var.container_port
    to_port     = var.container_port
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = { Name = "${var.service_name}-ecs-sg" }
}
