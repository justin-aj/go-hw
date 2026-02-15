# Product API

A Go-based REST API for managing products, built from an OpenAPI 3.0 specification.

## Setup

### Run Locally
```bash
go mod tidy
go run main.go
```

### Run with Docker
```bash
docker build -t product-api .
docker run -p 8080:8080 product-api
```

Server starts on `http://localhost:8080`

## API Endpoints

### 1. Add Product Details
**POST** `/products/{productId}/details`

```bash
curl -X POST http://localhost:8080/products/1/details -H "Content-Type: application/json" -d "{\"product_id\": 1, \"sku\": \"ABC-123-XYZ\", \"manufacturer\": \"Acme Corporation\", \"category_id\": 456, \"weight\": 1250, \"some_other_id\": 789}"
```
**Response:** `204 No Content`

### 2. Get Product by ID
**GET** `/products/{productId}`

```bash
curl http://localhost:8080/products/1
```
**Response:** `200 OK`
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

### Error Examples

**Product not found:**
```bash
curl http://localhost:8080/products/999
```
Response: `404`

**Invalid input:**
```bash
curl -X POST http://localhost:8080/products/1/details -H "Content-Type: application/json" -d "{\"product_id\": 1, \"sku\": \"\", \"manufacturer\": \"Acme\", \"category_id\": 1, \"weight\": 100, \"some_other_id\": 1}"
```
Response: `400`

## Status Codes

| Code | Meaning |
|------|---------|
| 200  | Success (GET) |
| 204  | Success (POST, no body) |
| 400  | Bad Request (invalid input) |
| 404  | Not Found |
| 500  | Internal Server Error |
