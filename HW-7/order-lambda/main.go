// Order Lambda — Part III
//
// Replaces the ECS Order Processor with a serverless function.
// SNS invokes this function directly for each order event.
//
// Cold starts:
//   The first invocation after a cold start includes an "Init Duration" line
//   in the CloudWatch REPORT.  Example:
//     REPORT RequestId: xyz Duration: 3005ms Init Duration: 73ms
//   Subsequent "warm" invocations omit Init Duration:
//     REPORT RequestId: abc Duration: 3001ms
//
//   For a 3-second processing job, 73ms cold start = 2.4% overhead — negligible.
//   Cold starts happen: first ever invocation, after ~5+ minutes idle,
//   or when concurrency spikes beyond current warm instances.
package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// snsWrapper mirrors the SNS-to-Lambda message envelope.
type snsWrapper struct {
	Message string `json:"Message"`
}

// order is a minimal representation for logging.
type order struct {
	OrderID    string `json:"order_id"`
	CustomerID int    `json:"customer_id"`
	Status     string `json:"status"`
}

// handler is invoked by Lambda for each SNS notification.
// SNS sends a batch of records; each record contains one order.
func handler(ctx context.Context, event events.SNSEvent) error {
	for _, record := range event.Records {
		msg := record.SNS.Message

		var o order
		if err := json.Unmarshal([]byte(msg), &o); err != nil {
			log.Printf("failed to parse order: %v — skipping", err)
			continue
		}

		log.Printf("Lambda processing order %s (customer %d)", o.OrderID, o.CustomerID)

		// ── Simulated payment processing ────────────────────────────────────
		// Same 3-second delay as ECS workers.
		time.Sleep(3 * time.Second)
		// ────────────────────────────────────────────────────────────────────

		log.Printf("Lambda completed order %s", o.OrderID)
	}
	return nil
}

func main() {
	lambda.Start(handler)
}
