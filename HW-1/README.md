# HW-1: Album Web Service with Gin Framework

A RESTful web service built with Go and the Gin framework for managing music albums. This assignment demonstrates fundamental REST API design patterns and HTTP request handling using industry-standard Go frameworks.

## Overview

This project implements a basic REST API for album management with CRUD operations. It serves as an introduction to building web services in Go using the Gin web framework and covers deployment/scaling considerations.

## Project Structure

```
HW-1/
├── web-service-gin/              # Main application directory
│   ├── main.go                   # Application entry point, route handlers
│   ├── main-syncmap.go           # Alternative implementation using sync.Map
│   ├── Dockerfile                # Docker containerization
│   ├── go.mod                    # Go module dependencies
│   ├── README.md                 # Project documentation
│   ├── load_testing.py           # Basic load testing script
│   ├── advanced_load_testing.py  # Advanced load testing with metrics
│   ├── test_two_instances.py     # Multi-instance testing
│   ├── advanced_load_testing.md  # Load testing documentation
│   ├── TWO_INSTANCE_EXPERIMENT.md # Scaling experiment results
│   ├── screenshot_results.md     # Performance screenshots
│   ├── terraform files           # Infrastructure as Code for AWS

```

## Features

- **Get all albums**: `GET /albums` - Retrieve complete album list
- **Get album by ID**: `GET /albums/{id}` - Retrieve specific album
- **Add new album**: `POST /albums` - Create new album record

## Prerequisites

- Go 1.16 or higher
- Docker (for containerized deployment)
- Terraform (for AWS deployment)
- Python 3.x with requests library (for load testing)

## Setup & Execution

### Local Development

```bash
cd HW-1/web-service-gin
go mod tidy
go run main.go
```

Server will start on `http://localhost:8080`

### Docker Deployment

```bash
docker build -t album-service .
docker run -p 8080:8080 album-service
```

### AWS Deployment with Terraform

```bash
cd HW-1/web-service-gin
terraform init
terraform plan
terraform apply
```

## API Examples

### Get All Albums
```bash
curl http://localhost:8080/albums
```

### Get Album by ID
```bash
curl http://localhost:8080/albums/1
```

### Add New Album
```bash
curl -X POST http://localhost:8080/albums \
  -H "Content-Type: application/json" \
  -d '{"id":"4","title":"Kind of Blue","artist":"Miles Davis","price":12.99}'
```

## Load Testing

Run basic load testing:
```bash
python load_testing.py
```

Run advanced load testing with detailed metrics:
```bash
python advanced_load_testing.py
```

Test multi-instance scaling:
```bash
python test_two_instances.py
```

## Key Experiments & Observations

### Scaling Analysis
- **Single Instance Performance**: Baseline throughput and latency measurements
- **Multi-Instance Scaling**: Testing with 2 instances to observe load distribution
- **Results**: See `TWO_INSTANCE_EXPERIMENT.md` for detailed findings

### Implementations
- **main.go**: Standard implementation using in-memory slice
- **main-syncmap.go**: Alternative using concurrent-safe `sync.Map`

## Results & Metrics

Load testing results and performance metrics are documented in:
- `advanced_load_testing.md` - Detailed load test analysis
- `screenshot_results.md` - Performance visualization
- `TWO_INSTANCE_EXPERIMENT.md` - Scaling test results

## Technologies Used

- **Go 1.16+**: Programming language
- **Gin Framework**: HTTP web framework
- **Docker**: Containerization
- **Terraform**: Infrastructure as Code
- **Python**: Load testing tooling
- **AWS EC2**: Cloud deployment platform

## Key Takeaways

This assignment covers:
- RESTful API design principles
- Go web framework usage (Gin)
- Containerization and deployment strategies
- Infrastructure as Code with Terraform
- Performance testing and load analysis
- Concurrent vs sequential request handling

## References

- [Gin Framework Documentation](https://gin-gonic.com/)
- [Go REST API Guide](https://golang.org/doc/)
- [Terraform AWS Provider](https://registry.terraform.io/providers/hashicorp/aws/latest/docs)

## Notes

- The service stores data in-memory; it will be lost on restart
- For production use, integrate with a persistent database
- SSL/TLS configuration recommended for production deployment
