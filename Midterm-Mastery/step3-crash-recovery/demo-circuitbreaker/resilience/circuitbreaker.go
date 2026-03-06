// ===================================================================
// CIRCUIT BREAKER PATTERN
// ===================================================================
//
// CONCEPT: "Stop calling something that's clearly broken."
//
// Problem:  Your downstream service (database, API) is down.
//           Every call fails after 5 seconds. But you keep calling
//           it — wasting 5 seconds × 1000 requests = 5000 seconds.
//
// Solution: Count consecutive failures. After 5 in a row, STOP
//           calling. Reject instantly. Wait 10 seconds, then try
//           ONE call to check if it's back.
//
// Three states:
//   CLOSED    → Normal. Calls go through.
//   OPEN      → Broken. All calls skipped instantly.
//   HALF-OPEN → Testing. One call goes through to check recovery.
//
// State transitions:
//   CLOSED  → OPEN:      5 consecutive failures
//   OPEN    → HALF-OPEN: 10 seconds elapsed
//   HALF-OPEN → CLOSED:  3 consecutive successes
//   HALF-OPEN → OPEN:    any failure
//
// In our demo: Wraps the __slow downstream call inside search.Execute().
//              Normal queries (?q=beauty) NEVER touch the circuit breaker.
// ===================================================================

package resilience

import (
	"fmt"
	"log"
	"sync"
	"time"
)

type CircuitState int

const (
	StateClosed   CircuitState = iota // Normal
	StateOpen                         // Broken — reject everything
	StateHalfOpen                     // Testing — let one through
)

func (s CircuitState) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateOpen:
		return "OPEN"
	case StateHalfOpen:
		return "HALF-OPEN"
	default:
		return "UNKNOWN"
	}
}

// CircuitBreaker tracks failures and decides whether to let calls through.
type CircuitBreaker struct {
	mu               sync.Mutex
	state            CircuitState
	failureCount     int
	successCount     int
	failureThreshold int           // trip after N failures (5)
	successThreshold int           // close after N successes (3)
	cooldown         time.Duration // wait before testing (10s)
	lastFailure      time.Time
	totalTrips       int
}

var ErrCircuitOpen = fmt.Errorf("circuit breaker is OPEN")

// NewCircuitBreaker creates a circuit breaker.
func NewCircuitBreaker(failureThreshold, successThreshold int, cooldown time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:            StateClosed,
		failureThreshold: failureThreshold,
		successThreshold: successThreshold,
		cooldown:         cooldown,
	}
}

// Call runs your function through the circuit breaker.
//
//	Circuit CLOSED  → runs fn
//	Circuit OPEN    → skips fn, returns ErrCircuitOpen
//	Circuit HALF-OPEN → runs fn as a test
func (cb *CircuitBreaker) Call(fn func() error) error {
	cb.mu.Lock()

	// OPEN + cooldown elapsed → move to HALF-OPEN
	if cb.state == StateOpen && time.Since(cb.lastFailure) > cb.cooldown {
		log.Println("[CIRCUIT-BREAKER] Cooldown elapsed — trying HALF-OPEN")
		cb.state = StateHalfOpen
		cb.successCount = 0
	}

	// Still OPEN → reject without calling fn
	if cb.state == StateOpen {
		cb.mu.Unlock()
		return ErrCircuitOpen
	}
	cb.mu.Unlock()

	// Call the downstream function
	err := fn()

	// Record the result
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failureCount++
		cb.successCount = 0
		cb.lastFailure = time.Now()

		if cb.failureCount >= cb.failureThreshold {
			log.Printf("[CIRCUIT-BREAKER] %d failures — circuit OPEN\n", cb.failureCount)
			cb.state = StateOpen
			cb.totalTrips++
		}
		return err
	}

	cb.successCount++
	cb.failureCount = 0
	if cb.state == StateHalfOpen && cb.successCount >= cb.successThreshold {
		log.Println("[CIRCUIT-BREAKER] Recovered — circuit CLOSED")
		cb.state = StateClosed
	}
	return nil
}

// Stats returns current state for /metrics.
type CircuitBreakerStats struct {
	State        string `json:"state"`
	Failures     int    `json:"consecutive_failures"`
	Successes    int    `json:"consecutive_successes"`
	TotalTrips   int    `json:"total_trips"`
	Threshold    int    `json:"failure_threshold"`
	CooldownSecs int    `json:"cooldown_seconds"`
}

func (cb *CircuitBreaker) Stats() CircuitBreakerStats {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return CircuitBreakerStats{
		State:        cb.state.String(),
		Failures:     cb.failureCount,
		Successes:    cb.successCount,
		TotalTrips:   cb.totalTrips,
		Threshold:    cb.failureThreshold,
		CooldownSecs: int(cb.cooldown.Seconds()),
	}
}
