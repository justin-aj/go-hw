// Package handler implements HTTP endpoints for the Order Receiver service.
//
// Two key endpoints:
//   POST /orders/sync  - synchronous: blocks until payment verified (3s)
//   POST /orders/async - asynchronous: publishes to SNS, returns 202 immediately
//
// The sync bottleneck is intentional and simulated with a buffered channel.
// A buffered channel of size 1 means only ONE payment can be "in-flight" at a
// time, forcing others to queue or fail — this reproduces a single-threaded
// payment processor.
package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/google/uuid"

	"order-receiver/model"
)

// Handler holds shared dependencies (SNS client, config).
type Handler struct {
	snsClient *sns.Client
	topicARN  string

	// paymentSlot is a buffered channel of size 1.
	// Acquiring the slot = entering the (simulated) payment processor.
	// Only one goroutine can hold the slot at a time, so throughput is
	// capped at 1 / 3s ≈ 0.33 orders/second — exactly like a single-threaded
	// payment backend.
	paymentSlot chan struct{}
}

// New constructs a Handler, initialising the AWS SNS client and the
// payment-slot channel.  SNS_TOPIC_ARN must be set in the environment.
func New() *Handler {
	topicARN := os.Getenv("SNS_TOPIC_ARN")

	// Load AWS SDK config from environment / instance metadata.
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Printf("WARNING: could not load AWS config: %v", err)
	}

	return &Handler{
		snsClient:   sns.NewFromConfig(cfg),
		topicARN:    topicARN,
		paymentSlot: make(chan struct{}, 1), // capacity = 1 = single processor lane
	}
}

// RegisterRoutes wires all endpoints onto mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", h.health)
	mux.HandleFunc("/orders/sync", h.syncOrder)
	mux.HandleFunc("/orders/async", h.asyncOrder)
}

// health is a simple liveness probe used by the ALB / ECS health check.
func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "ok")
}

// syncOrder processes an order synchronously.
//
// Flow:
//   1. Decode the incoming JSON body into an Order.
//   2. Acquire the payment slot (blocks if slot is full — another order is
//      already being processed).
//   3. Sleep 3 seconds to simulate payment verification latency.
//   4. Release the slot.
//   5. Return 200 OK with the completed order.
//
// During a flash sale, concurrent requests pile up waiting for the slot.
// The OS/Go runtime will queue goroutines, but from the client's perspective
// the response time explodes (5 concurrent × 3s = 15s for the last customer).
func (h *Handler) syncOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	order, err := decodeOrder(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	order.Status = "processing"
	start := time.Now()

	// ── Bottleneck simulation ────────────────────────────────────────────────
	// send to slot (blocks if full), sleep, then drain the slot.
	h.paymentSlot <- struct{}{}       // acquire
	time.Sleep(3 * time.Second)       // simulate payment processor
	<-h.paymentSlot                   // release
	// ─────────────────────────────────────────────────────────────────────────

	order.Status = "completed"
	log.Printf("[sync] order %s completed in %v", order.OrderID, time.Since(start))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(order)
}

// asyncOrder acknowledges the order immediately and offloads processing to SNS.
//
// Flow:
//   1. Decode the order.
//   2. Publish the order JSON as a message to the SNS topic.
//   3. Return 202 Accepted — no waiting for payment.
//
// The background Order Processor service reads from the SQS queue that is
// subscribed to this SNS topic and performs the actual 3-second payment step.
func (h *Handler) asyncOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	order, err := decodeOrder(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	order.Status = "pending"

	// Serialize order to JSON for the SNS message body.
	body, err := json.Marshal(order)
	if err != nil {
		http.Error(w, "failed to marshal order", http.StatusInternalServerError)
		return
	}

	// Publish to SNS.  SNS fans out to all subscribed SQS queues (and Lambdas).
	_, err = h.snsClient.Publish(context.Background(), &sns.PublishInput{
		TopicArn: aws.String(h.topicARN),
		Message:  aws.String(string(body)),
	})
	if err != nil {
		log.Printf("[async] SNS publish failed: %v", err)
		http.Error(w, "failed to queue order", http.StatusInternalServerError)
		return
	}

	log.Printf("[async] order %s queued", order.OrderID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted) // 202 — "I got it, working on it"
	json.NewEncoder(w).Encode(map[string]string{
		"order_id": order.OrderID,
		"status":   "pending",
		"message":  "order queued for processing",
	})
}

// decodeOrder parses the request body into a model.Order, generating an
// OrderID and CreatedAt if the caller didn't provide them.
func decodeOrder(r *http.Request) (model.Order, error) {
	var order model.Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		return order, fmt.Errorf("invalid JSON: %w", err)
	}
	if order.OrderID == "" {
		order.OrderID = uuid.New().String()
	}
	if order.CreatedAt.IsZero() {
		order.CreatedAt = time.Now()
	}
	return order, nil
}
