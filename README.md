# CS 6650: Advanced Design of Large-Scale Distributed Systems
## Homework Repository

Welcome to the course repository containing all homework assignments exploring distributed systems, scalability, and cloud architecture. Each assignment builds on core concepts with hands-on implementation and performance analysis.

## 📚 Assignments Overview

### [HW-1: Album Web Service with Gin Framework](./HW-1)
A foundational REST API project demonstrating web service basics and load testing.

**Key Topics:**
- RESTful API design patterns
- Go HTTP framework (Gin)
- In-memory data structures
- Load testing and performance metrics
- Single vs multi-instance deployment

**Tech Stack:** Go, Gin, Docker, Terraform, Python/Locust

**Key Files:**
- `HW-1/web-service-gin/main.go` - Service implementation
- `HW-1/web-service-gin/load_testing.py` - Basic load tests
- `HW-1/web-service-gin/advanced_load_testing.py` - Advanced metrics

---

### [HW-2: AWS EC2 Infrastructure Setup](./HW-2)
Infrastructure as Code project for cloud resource provisioning and management.

**Key Topics:**
- Terraform fundamentals
- AWS EC2 instance provisioning
- Security groups and network configuration
- Infrastructure state management
- SSH access and remote management

**Tech Stack:** Terraform, AWS (EC2, VPC), SSH

**Key Files:**
- `HW-2/main.tf` - Infrastructure definition
- `HW-2/terraform.tfvars` - Configuration variables
- `HW-2/hw2.pem` - SSH key pair

---

### [HW-3: Go Concurrency & Synchronization Experiments](./HW-3)
Deep dive into concurrent programming patterns and performance tradeoffs in Go.

**Key Topics:**
- Race condition detection and prevention
- Synchronization primitives (Mutex, RWMutex)
- Concurrent-safe data structures (sync.Map)
- Context switching analysis
- I/O blocking patterns
- Atomic operations and lock-free programming

**Tech Stack:** Go, Docker, Python/Locust

**Key Experiments:**
- Race conditions (4 implementations with varying fixes)
- Context switching overhead analysis
- File I/O buffering strategies
- Load testing with Locust framework

**Key Files:**
- `HW-3/race-condition*.go` - Race condition examples
- `HW-3/context-switching-experiment.go` - Context switching analysis
- `HW-3/file-io-experiment.go` - I/O pattern testing
- `HW-3/*_EXPERIMENT.md` - Detailed analysis documents

---

### [HW-4: MapReduce Implementation & Scaling Analysis](./HW-4)
Distributed data processing pipeline demonstrating big data processing patterns.

**Key Topics:**
- MapReduce programming model
- Data partitioning and distributed processing
- Container orchestration
- Docker networking and communication
- Horizontal scaling analysis
- Fault tolerance and retry mechanisms

**Tech Stack:** Go, Docker, Docker Compose, Python

**Architecture:**
- **Splitter**: Partitions input data into chunks
- **Mapper**: Processes chunks in parallel (word counting)
- **Reducer**: Aggregates results from mappers

**Key Files:**
- `HW-4/mapreduce/orchestrator.py` - Pipeline orchestration
- `HW-4/mapreduce/performance.py` - Metrics collection
- `HW-4/mapreduce/mapper/main.go` - Mapper implementation
- `HW-4/mapreduce/reducer/main.go` - Reducer implementation
- `HW-4/mapreduce/splitter/main.go` - Data splitter

**Running the Pipeline:**
```bash
cd HW-4/mapreduce
python orchestrator.py run        # Run full pipeline
python orchestrator.py scale      # Test scaling performance
```

---

### [HW-5: Product API with OpenAPI Specification](./HW-5)
API-first development using OpenAPI specification with comprehensive deployment options.

**Key Topics:**
- API-first development methodology
- OpenAPI 3.0 specification design
- RESTful endpoint implementation
- Docker containerization
- Infrastructure as Code (Terraform)
- Load testing and performance validation

**Tech Stack:** Go, OpenAPI 3.0, Docker, Terraform, AWS, Python/Locust

**API Endpoints:**
- `POST /products/{productId}/details` - Add product details
- `GET /products/{productId}` - Retrieve product information

**Deployment Options:**
- Local development with `go run`
- Containerized with Docker
- Cloud deployment with Terraform (Part 2 & 3 configurations)

**Key Files:**
- `HW-5/src/main.go` - API implementation
- `HW-5/src/api.yaml` - OpenAPI specification
- `HW-5/terraform/` - Infrastructure definitions
- `HW-5/locustfile-fast.py` - Load testing

---

### [HW-6: Product Search Service - Modularized Architecture](./HW-6)
Advanced software architecture following Parnas' decomposition principles for clean module design.

**Key Topics:**
- Modular architecture design (Parnas decomposition)
- Interface-based design and dependency injection
- Concurrent data structures (sync.Map)
- Clean code and separation of concerns
- Single responsibility principle
- Testable module design

**Tech Stack:** Go, Docker, Terraform, AWS, Python/Locust

**Module Architecture:**
```
main (composition root)
├── model       (data representation)
├── store       (concurrent-safe storage - sync.Map)
├── seeddata    (seed data source)
├── generator   (100K product generation)
├── search      (search algorithm)
└── handler     (HTTP transport layer)
```

**Key Concept:**
Each module hides one design decision behind a stable interface:
- `model` → Product data representation
- `store` → Storage mechanism & concurrency
- `seeddata` → Seed catalog source
- `generator` → Expansion strategy
- `search` → Matching algorithm
- `handler` → HTTP serialization

**Performance:**
- 100K concurrent products
- ~5000 RPS single instance
- ~15000 RPS with 3-instance load balancing
- P50 latency: 5-10ms

**Key Files:**
- `HW-6/main.go` - Service composition
- `HW-6/model/product.go` - Data model
- `HW-6/store/store.go` - Concurrent storage
- `HW-6/search/search.go` - Search implementation
- `HW-6/handler/handler.go` - HTTP handlers
- `HW-6/terraform/` - Infrastructure

---

## 🚀 Quick Start

### Prerequisites
- **Go**: 1.16+
- **Docker & Docker Compose**: Latest stable
- **Terraform**: 1.0+
- **Python**: 3.8+ with pip
- **AWS Account**: For HW-2, HW-5, HW-6 deployments
- **Git**: For version control

### Running Individual Assignments

**HW-1: Album Service**
```bash
cd HW-1/web-service-gin
go mod tidy
go run main.go
curl http://localhost:8080/albums
```

**HW-2: EC2 Infrastructure**
```bash
cd HW-2
terraform init
terraform plan
terraform apply
```

**HW-3: Concurrency Experiments**
```bash
cd HW-3
go run -race race-condition-2.go
go run context-switching-experiment.go
docker-compose -f docker-compose-locust.yml up
```

**HW-4: MapReduce Pipeline**
```bash
cd HW-4/mapreduce
docker-compose build
python orchestrator.py run
python orchestrator.py scale
```

**HW-5: Product API**
```bash
cd HW-5/src
go mod tidy
go run main.go
curl -X POST http://localhost:8080/products/1/details -H "Content-Type: application/json" -d '{"product_id":1,"sku":"ABC-123"}'
```

**HW-6: Product Search**
```bash
cd HW-6
go mod tidy
go run main.go
curl "http://localhost:8080/search?q=laptop"
```

---

## 📊 Learning Progression

1. **HW-1 & HW-2**: Foundations
   - Basic REST API design
   - Infrastructure provisioning

2. **HW-3 & HW-4**: Concurrency & Scalability
   - Concurrent programming patterns
   - Distributed data processing

3. **HW-5 & HW-6**: Advanced Architecture
   - API specification-driven design
   - Modular software architecture

---

## 🔑 Key Concepts Covered

### Distributed Systems
- Data partitioning and distribution
- MapReduce programming model
- Horizontal scaling strategies
- Load balancing and failover

### Concurrency & Performance
- Race conditions and synchronization
- Mutex vs RWMutex tradeoffs
- Lock-free data structures
- Context switching overhead
- Load testing methodologies

### Software Architecture
- Modular design principles (Parnas decomposition)
- Interface-based design
- Dependency injection
- Single responsibility principle
- Clean code practices

### Cloud & DevOps
- Infrastructure as Code (Terraform)
- Docker containerization
- Container orchestration
- AWS services (EC2, VPC, ELB, CloudWatch)
- CI/CD pipelines

### API Design
- RESTful principles
- OpenAPI specification
- Error handling
- Performance optimization
- Scaling patterns

---

## 📈 Performance Reference

| Assignment | Throughput | Latency (P50) | Optimization |
|------------|-----------|---------------|---------------|
| HW-1 | ~500 RPS | 20-30ms | Single instance baseline |
| HW-3 | Varies | Synchronization dependent | Race condition analysis |
| HW-4 | 1 file/sec | Split, map, reduce stages | Horizontal scaling |
| HW-5 | ~1000 RPS | 10-20ms | OpenAPI specification |
| HW-6 | ~5000 RPS | 5-10ms | Optimized architecture |

---

## 🛠️ Tools & Technologies

### Languages & Frameworks
- **Go** 1.16+ - Backend services
- **Python** 3.8+ - Load testing, orchestration
- **Bash/PowerShell** - Scripting

### Infrastructure & Deployment
- **Docker** - Containerization
- **Docker Compose** - Multi-container orchestration
- **Terraform** - Infrastructure as Code
- **AWS** - Cloud provider

### Testing & Monitoring
- **Locust** - Load testing framework
- **Prometheus** - Metrics collection (if applicable)
- **CloudWatch** - AWS monitoring

---

## 📝 Documentation

Each homework directory includes:
- **README.md** - Complete assignment overview
- **Experiment markdown files** - Detailed analysis and results
- **Source code** - Well-commented implementations
- **Configuration files** - Terraform, Docker compose, API specs

---

## 🎓 Learning Outcomes

After completing this course, you will understand:

✅ How to design and implement scalable distributed systems
✅ Concurrency patterns and performance tradeoffs in Go
✅ Data processing techniques for large-scale systems
✅ Cloud infrastructure provisioning and management
✅ API design using industry-standard specifications
✅ Modular architecture and clean code principles
✅ Performance testing and optimization strategies
✅ Container orchestration and deployment patterns

---

## 🔗 References & Resources

### Go Concurrency
- [Effective Go - Concurrency](https://golang.org/doc/effective_go#concurrency)
- [sync Package Documentation](https://pkg.go.dev/sync)
- [Context Package Guide](https://pkg.go.dev/context)

### Distributed Systems
- [MapReduce Paper](https://research.google/pubs/mapreduce-simplified-data-processing-on-large-clusters/)
- [Designing Data-Intensive Applications](https://dataintensive.net/)

### Architecture & Design
- [Parnas (1972) - Criteria for Decomposing Systems into Modules](https://www.win.tue.nl/~wstomv/edu/2ip30/references/criteria_for_modularization.pdf)
- [Clean Architecture by Uncle Bob](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)

### DevOps & Infrastructure
- [Terraform Documentation](https://www.terraform.io/docs)
- [Docker Best Practices](https://docs.docker.com/develop/dev-best-practices/)
- [AWS EC2 Guide](https://docs.aws.amazon.com/ec2/)

### API Design
- [OpenAPI Specification](https://swagger.io/specification/)
- [RESTful API Best Practices](https://restfulapi.net/)

---

## 📞 Support & Questions

For specific assignment help, refer to the individual README files:
- [HW-1 README](./HW-1/README.md)
- [HW-2 README](./HW-2/README.md)
- [HW-3 README](./HW-3/README.md)
- [HW-4 README](./HW-4/README.md)
- [HW-5 README](./HW-5/README.md)
- [HW-6 README](./HW-6/README.md)

---

## 📄 License

Course materials for CS 6650: Advanced Design of Large-Scale Distributed Systems

---

## 🎯 Next Steps

1. **Start with HW-1**: Get familiar with REST APIs and load testing
2. **Move to HW-2**: Learn infrastructure provisioning
3. **Deep dive HW-3**: Understand concurrency patterns
4. **Explore HW-4**: Master distributed processing
5. **Build HW-5**: Design with specifications
6. **Architect HW-6**: Apply modular design principles

Good luck! 🚀
