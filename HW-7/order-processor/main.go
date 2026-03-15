// Order Processor Service
//
// Polls an SQS queue for orders published by the Order Receiver and
// processes them with a configurable number of concurrent worker goroutines.
//
// Environment variables:
//   SQS_QUEUE_URL  - URL of the SQS queue to poll
//   NUM_WORKERS    - number of concurrent worker goroutines (default: 1)
//
// Architecture:
//
//   main goroutine
//     └─ poll loop: ReceiveMessage (long-poll, up to 20s)
//          └─ for each message → dispatch to worker pool via jobs channel
//
//   worker goroutines (NUM_WORKERS of them)
//     └─ read from jobs channel
//          └─ processPayment (3s delay)
//          └─ DeleteMessage from SQS (acknowledge)
//
// Why a semaphore / worker pool?
//   Each goroutine costs ~2KB of stack.  100 goroutines = ~200KB — very
//   cheap.  But if we spawned an unbounded number we could overwhelm the
//   CPU.  A fixed pool lets us tune throughput vs resource usage.
//
// Phase 3 → 1 worker:   0.33 orders/s  (baseline)
// Phase 5 → 5 workers:  ~1.67 orders/s
//           20 workers: ~6.67 orders/s
//           100 workers: ~33 orders/s  (near the 60/s target)
//           180 workers: ~60 orders/s  (matches flash sale demand)
package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

// snsWrapper is the envelope SNS wraps around messages before putting them
// in SQS.  The actual order JSON lives in the "Message" field.
type snsWrapper struct {
	Message string `json:"Message"`
}

func main() {
	queueURL := mustEnv("SQS_QUEUE_URL")
	numWorkers := envInt("NUM_WORKERS", 1)

	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("failed to load AWS config: %v", err)
	}
	client := sqs.NewFromConfig(cfg)

	// jobs is an unbuffered channel: the poll loop blocks until a worker is
	// free, providing natural back-pressure.
	jobs := make(chan types.Message)

	// Start the worker pool.
	log.Printf("Starting %d worker goroutine(s)", numWorkers)
	for i := 0; i < numWorkers; i++ {
		go worker(i, client, queueURL, jobs)
	}

	// Poll SQS forever.
	log.Printf("Polling SQS queue: %s", queueURL)
	for {
		result, err := client.ReceiveMessage(context.Background(), &sqs.ReceiveMessageInput{
			QueueUrl: aws.String(queueURL),
			// WaitTimeSeconds=20: long polling — waits up to 20s for messages
			// rather than returning immediately.  This reduces API calls and
			// cost when the queue is quiet.
			WaitTimeSeconds: 20,
			// Fetch up to 10 messages per call (SQS maximum).
			MaxNumberOfMessages: 10,
		})
		if err != nil {
			log.Printf("ReceiveMessage error: %v — retrying in 5s", err)
			time.Sleep(5 * time.Second)
			continue
		}

		for _, msg := range result.Messages {
			// Send to the jobs channel.  This blocks if all workers are busy,
			// which is fine — it limits how many messages we hold in memory.
			jobs <- msg
		}
	}
}

// worker drains the jobs channel, processing each message with a 3-second
// simulated payment delay, then deletes it from SQS.
func worker(id int, client *sqs.Client, queueURL string, jobs <-chan types.Message) {
	for msg := range jobs {
		processMessage(id, client, queueURL, msg)
	}
}

func processMessage(workerID int, client *sqs.Client, queueURL string, msg types.Message) {
	// SNS wraps the original message body.  Unwrap it.
	var wrapper snsWrapper
	if err := json.Unmarshal([]byte(*msg.Body), &wrapper); err != nil {
		log.Printf("[worker %d] failed to unwrap SNS envelope: %v", workerID, err)
		return
	}

	// Parse the order from the inner JSON.
	var order map[string]interface{}
	if err := json.Unmarshal([]byte(wrapper.Message), &order); err != nil {
		log.Printf("[worker %d] failed to parse order: %v", workerID, err)
		return
	}

	orderID, _ := order["order_id"].(string)
	log.Printf("[worker %d] processing order %s", workerID, orderID)

	// ── Simulated payment processing ────────────────────────────────────────
	// Same 3-second delay as the sync endpoint — same backend, just async.
	time.Sleep(3 * time.Second)
	// ─────────────────────────────────────────────────────────────────────────

	log.Printf("[worker %d] order %s completed", workerID, orderID)

	// Delete the message from SQS to acknowledge successful processing.
	// If we crash before this, SQS will redeliver after the visibility timeout.
	_, err := client.DeleteMessage(context.Background(), &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(queueURL),
		ReceiptHandle: msg.ReceiptHandle,
	})
	if err != nil {
		log.Printf("[worker %d] failed to delete message: %v", workerID, err)
	}
}

// mustEnv reads a required environment variable or exits.
func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("required env var %s is not set", key)
	}
	return v
}

// envInt reads an integer env var with a default fallback.
func envInt(key string, def int) int {
	s := os.Getenv(key)
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		log.Printf("invalid %s=%q, using default %d", key, s, def)
		return def
	}
	return n
}
