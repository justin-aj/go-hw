// Shopping Cart Service
//
// Implements the HW-5 shopping-cart endpoints backed by either MySQL (RDS) or
// DynamoDB. The backend is selected at startup via the DB_BACKEND env var.
//
// Environment variables:
//   DB_BACKEND    "mysql" (default) or "dynamo"
//   PORT          HTTP listen port (default "8080")
//
// MySQL vars:
//   DB_HOST, DB_PORT, DB_NAME, DB_USER, DB_PASSWORD
//
// DynamoDB vars:
//   DYNAMO_TABLE, AWS_REGION
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"cart-service/handler"
	"cart-service/store"
)

func main() {
	backend := envOr("DB_BACKEND", "mysql")

	var s store.Store
	var err error

	switch backend {
	case "dynamo":
		region := envOr("AWS_REGION", "us-east-1")
		table := os.Getenv("DYNAMO_TABLE")
		s, err = store.NewDynamoDB(context.Background(), region, table)
	default: // mysql
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_HOST"),
			envOr("DB_PORT", "3306"),
			os.Getenv("DB_NAME"),
		)
		s, err = store.NewMySQL(dsn)
	}

	if err != nil {
		log.Fatalf("store init (%s): %v", backend, err)
	}
	defer s.Close()

	h := handler.New(s)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	port := envOr("PORT", "8080")
	log.Printf("shopping-cart service on :%s  backend=%s", port, backend)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
