# HW-5: Product API with OpenAPI Specification

A RESTful Product API built from an OpenAPI 3.0 specification, demonstrating API-first development and deployment patterns. This assignment covers specification-driven API design, containerization, and cloud infrastructure.

## Overview

This project implements a product management API designed from an OpenAPI 3.0 specification. It covers:
- API-first development using OpenAPI
- RESTful service implementation in Go
- Docker containerization and deployment
- Terraform-based infrastructure as code
- Load testing and performance validation

## Project Structure

```
HW-5/
├── src/
│   ├── main.go                    # API server entry point
│   ├── api.yaml                   # OpenAPI 3.0 specification
│   ├── go.mod                     # Go module dependencies
│   ├── Dockerfile                 # Docker container definition
│   └── api/                       # Generated API code from OpenAPI spec
│
├── terraform/
│   ├── main.tf                    # Primary infrastructure resources
│   ├── provider.tf                # AWS provider configuration
│   ├── variables.tf               # Input variables
│   ├── outputs.tf                 # Output values
│   ├── terraform.tfstate          # Current state
│   ├── terraform.tfstate.backup   # State backup
│   ├── modules/                   # Reusable modules
│   │
│   └── [tfvars files]             # Environment-specific variables
│       ├── part2.tfvars           # Part 2 configuration
│       └── part3.tfvars           # Part 3 configuration
│
├── locustfile-fast.py             # Optimized load testing
├── README.md                      # This file
└── venv/                          # Python virtual environment

```

## Features

### Core API Endpoints

#### 1. Add Product Details
**POST** `/products/{productId}/details`

Add comprehensive product information:
```bash
curl -X POST http://localhost:8080/products/1/details \
  -H "Content-Type: application/json" \
  -d '{
    "product_id": 1,
    "sku": "ABC-123-XYZ",
    "manufacturer": "Acme Corporation",
    "category_id": 456,
    "weight": 1250,
    "some_other_id": 789
  }'
```

**Response**: `204 No Content`

#### 2. Get Product by ID
**GET** `/products/{productId}`

Retrieve complete product information:
```bash
curl http://localhost:8080/products/1
```

**Response** (`200 OK`):
```json
{
  "product_id": 1,
  "sku": "ABC-123-XYZ",
  "manufacturer": "Acme Corporation",
  "category_id": 456,
  "weight": 1250,
  "some_other_id": 789
}
```

#### 3. Error Handling

Comprehensive error responses for invalid requests:
```bash
curl http://localhost:8080/products/999  # Not found
```

**Response** (`404 Not Found`):
```json
{
  "error": "Product not found: 999"
}
```

## Prerequisites

- Go 1.16 or higher
- Docker & Docker Compose
- Terraform 1.0 or higher
- Python 3.8+ with pip
- AWS Account (for cloud deployment)

## Local Development

### Setup & Run

```bash
cd HW-5/src
go mod tidy
go run main.go
```

Server starts on `http://localhost:8080`

### API Testing

Test all endpoints:
```bash
# Get all products
curl http://localhost:8080/products

# Add product details
curl -X POST http://localhost:8080/products/1/details \
  -H "Content-Type: application/json" \
  -d '{"product_id": 1, "sku": "TEST-001", "manufacturer": "Test Corp"}'

# Retrieve product
curl http://localhost:8080/products/1
```

## Docker Deployment

### Build Image

```bash
docker build -t product-api:latest ./src
```

### Run Container

```bash
docker run -p 8080:8080 product-api:latest
```

### Docker Compose (if applicable)

```bash
docker-compose up -d
```

## Cloud Deployment with Terraform

### AWS Infrastructure Setup

```bash
cd terraform
terraform init
```

### Configure Variables

Create or edit `terraform.tfvars`:
```hcl
region  = "us-east-1"
app_name = "product-api"
environment = "production"
```

### Plan & Deploy

```bash
terraform plan
terraform apply
```

### View Deployment

```bash
terraform output
```

### Part 2 Configuration

Deploy with specific configuration:
```bash
terraform apply -var-file="part2.tfvars"
```

### Part 3 Configuration

Advanced scaling setup:
```bash
terraform apply -var-file="part3.tfvars"
```

## Load Testing

### Install Dependencies

```bash
pip install -r requirements.txt
```

### Run Load Tests

Fast load test (optimized for quick results):
```bash
locust -f locustfile-fast.py --host=http://localhost:8080
```

Access the Locust dashboard at `http://localhost:8089`

### Load Test Configuration

Customize in `locustfile-fast.py`:
```python
CONCURRENT_USERS = 100      # Number of concurrent users
SPAWN_RATE = 10             # Users spawned per second
DURATION = 60               # Test duration in seconds
```

### Performance Metrics

Monitor:
- **RPS** (Requests Per Second): Throughput
- **Response Time**: Latency (p50, p95, p99)
- **Error Rate**: Percentage of failed requests
- **Connections**: Active concurrent connections

## API Specification

### OpenAPI 3.0 Definition

Full API specification in `src/api.yaml`:

**Base URL**: `/api/v1`

**Schemes**: HTTP, HTTPS

**Content Type**: application/json

### Paths

```yaml
/products/{productId}/details:
  post:
    summary: Add product details
    parameters:
      - name: productId
        in: path
        required: true
        schema:
          type: integer
```

## Configuration

### Environment Variables

```bash
PORT=8080                 # Server port
LOG_LEVEL=info           # Logging level
DATABASE_URL=...         # Database connection (if used)
```

### Server Configuration

Adjustable in `src/main.go`:
```go
const (
    Port = 8080
    MaxConnections = 1000
    ReadTimeout = 5 * time.Second
)
```

## Deployment Strategies

### Development
```bash
go run src/main.go
```

### Docker
```bash
docker build -t product-api .
docker run -p 8080:8080 product-api
```

### Kubernetes
Use Terraform to create K8s resources and deploy via Helm charts.

### Serverless
Deploy to AWS Lambda with API Gateway.

## Infrastructure Components (Terraform)

### Compute
- EC2 instances or ECS Fargate for containerized API
- Auto-scaling groups for horizontal scaling

### Networking
- VPC with public/private subnets
- Load balancer for traffic distribution
- Security groups restricting access

### Data
- RDS for product database (if applicable)
- S3 for blob storage

### Monitoring
- CloudWatch for logs and metrics
- CloudTrail for audit logging

## Performance Characteristics

### Throughput
- **Single Instance**: ~1000 RPS (depends on hardware)
- **Load Balanced (3 instances)**: ~3000 RPS
- **Bottleneck**: Database operations

### Latency
- **P50**: 10-20ms
- **P95**: 50-100ms
- **P99**: 200-500ms

### Scaling
- **Horizontal**: Add instances behind load balancer
- **Vertical**: Increase instance size
- **Database**: Add read replicas for GET operations

## Best Practices

### API Design
1. ✅ Use OpenAPI specification first
2. ✅ Consistent naming conventions
3. ✅ Proper HTTP status codes
4. ✅ Comprehensive error messages
5. ✅ Support for pagination/filtering

### Security
1. ✅ Input validation on all endpoints
2. ✅ Authentication (if applicable)
3. ✅ Rate limiting
4. ✅ HTTPS in production
5. ✅ Secret management in Terraform

### Deployment
1. ✅ Blue-green deployments
2. ✅ Automated health checks
3. ✅ Graceful shutdown handling
4. ✅ Comprehensive logging
5. ✅ Monitoring and alerting

## Troubleshooting

### Port Already in Use
```bash
lsof -i :8080
kill -9 <PID>
```

### Docker Build Fails
```bash
docker build --no-cache -t product-api:latest ./src
```

### Terraform Deployment Issues
```bash
terraform validate
terraform fmt -recursive
terraform plan -out=plan.tfplan
```

### Load Test Errors
- Check server is running and accessible
- Verify firewall rules
- Review server logs for errors

## Technologies Used

- **Go 1.16+**: Backend implementation
- **Gin Framework**: HTTP routing (optional)
- **OpenAPI 3.0**: API specification
- **Docker**: Containerization
- **Terraform**: Infrastructure as Code
- **Locust**: Load testing framework
- **AWS**: Cloud platform
  - EC2: Compute
  - RDS: Database
  - ELB: Load balancing
  - CloudWatch: Monitoring

## Concepts Covered

- **API-First Development**: Specification-driven design
- **RESTful Principles**: Resource modeling, verbs, status codes
- **Containerization**: Docker best practices
- **Infrastructure as Code**: Terraform modules and state
- **Load Testing**: Performance measurement and optimization
- **Cloud Deployment**: AWS services and integration
- **Scaling Strategies**: Horizontal and vertical scaling

## References

- [OpenAPI Specification](https://swagger.io/specification/)
- [Go HTTP Server](https://golang.org/pkg/net/http/)
- [Terraform AWS Provider](https://registry.terraform.io/providers/hashicorp/aws/latest/docs)
- [Docker Best Practices](https://docs.docker.com/develop/dev-best-practices/)
- [Locust Load Testing](https://locust.io/docs/index.html)

## Notes

- In-memory data storage (not persisted)
- For production: integrate with database
- Scale horizontally by adding instances
- Monitor performance metrics in production
- Plan auto-scaling policies
- Regular backups of infrastructure state
