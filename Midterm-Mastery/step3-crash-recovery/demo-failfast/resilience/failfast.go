// ===================================================================
// FAIL-FAST PATTERN
// ===================================================================
//
// CONCEPT: "Don't wait forever for a slow response."
//
// Problem:  A request is stuck (slow database, unresponsive API).
//           Without a timeout, the caller waits forever.
//
// Solution: Set a timer. If the response doesn't come back in time,
//           give up and return an error immediately.
//
// In our demo: The __slow query takes 5 seconds. Fail-fast kills it
//              at 500ms, freeing the server to handle other requests.
// ===================================================================

package resilience

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// FailFast adds a timeout to any HTTP handler.
type FailFast struct {
	next    http.Handler  // the handler to protect
	timeout time.Duration // max allowed time (500ms)
}

// NewFailFast wraps a handler with a timeout.
func NewFailFast(next http.Handler, timeout time.Duration) *FailFast {
	return &FailFast{next: next, timeout: timeout}
}

// ServeHTTP runs on every request. It races the handler against a timer.
func (ff *FailFast) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Start a countdown timer.
	ctx, cancel := context.WithTimeout(r.Context(), ff.timeout)
	defer cancel()

	// Run the handler in the background, capturing its response.
	rec := &responseRecorder{header: make(http.Header), status: 200}
	done := make(chan struct{})
	go func() {
		ff.next.ServeHTTP(rec, r.WithContext(ctx))
		close(done)
	}()

	// Wait: who finishes first — the handler or the timer?
	select {
	case <-done:
		// Handler finished in time. Send its response to the client.
		for k, v := range rec.header {
			w.Header()[k] = v
		}
		w.WriteHeader(rec.status)
		w.Write(rec.body)

	case <-ctx.Done():
		// Timer expired. Handler is too slow. Return 503.
		log.Printf("[FAIL-FAST] Request timed out — returning 503\n")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(503)
		json.NewEncoder(w).Encode(map[string]string{
			"error":   "request timed out",
			"pattern": "fail-fast",
		})
	}
}

// responseRecorder captures a handler's output so we can discard it on timeout.
type responseRecorder struct {
	header http.Header
	body   []byte
	status int
}

func (r *responseRecorder) Header() http.Header        { return r.header }
func (r *responseRecorder) WriteHeader(statusCode int) { r.status = statusCode }
func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body = append(r.body, b...)
	return len(b), nil
}
