// load-tester — measures latency and stale reads across all KV cluster modes.
//
// "Local-in-time" strategy: use a small key space (default 100 keys).
// With a small key space, reads and writes hit the same keys repeatedly and
// are naturally clustered closely together in time — exactly what is needed
// to expose stale reads during the replication window.
//
// Usage examples:
//
//	# W=5 R=1, 10% writes
//	./load-tester -leader http://localhost:9080 \
//	              -nodes  http://localhost:9080,http://localhost:9081,...,9084 \
//	              -mode w5r1 -write-pct 10 -output w5r1_10w.csv
//
//	# Leaderless, 50% writes
//	./load-tester -nodes http://localhost:9085,...,9089 \
//	              -mode leaderless -write-pct 50 -output leaderless_50w.csv
//
// Flags:
//
//	-leader     URL writes go to (leader-follower modes). Omit for leaderless.
//	-nodes      Comma-separated list of ALL node URLs. Reads pick one randomly.
//	            In leaderless mode writes also pick randomly from this list.
//	-mode       Label written into the CSV: w5r1 | w1r5 | w3r3 | leaderless
//	-write-pct  Integer percentage of requests that are writes: 1 | 10 | 50 | 90
//	-duration   Seconds to run (default 30)
//	-workers    Concurrent goroutines (default 10)
//	-keys       Key-space size (default 100; smaller = more local-in-time clustering)
//	-output     Path to the output CSV file (default results.csv)
package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ---- result record ----------------------------------------------------------

type record struct {
	ts          time.Time
	reqType     string  // "read" or "write"
	key         string
	latencyMs   float64
	status      int
	version     int64
	stale       bool
	rwIntervalMs float64 // ms since the last write to this key; -1 if unknown
}

// ---- version tracker --------------------------------------------------------
// The tracker maintains the latest version the client has written for each key
// so that a read response can be compared against it to detect staleness.
// It also records the wall-clock time of each write so that the read-write
// interval for the same key can be computed.

type tracker struct {
	mu            sync.Mutex
	knownVersion  map[string]int64
	lastWriteTime map[string]time.Time
}

func newTracker() *tracker {
	return &tracker{
		knownVersion:  make(map[string]int64),
		lastWriteTime: make(map[string]time.Time),
	}
}

func (t *tracker) recordWrite(key string, version int64) {
	t.mu.Lock()
	t.knownVersion[key] = version
	t.lastWriteTime[key] = time.Now()
	t.mu.Unlock()
}

// checkRead returns whether the read is stale and the ms since the last write
// to this key (-1 if the key has never been written by this client).
func (t *tracker) checkRead(key string, returnedVersion int64) (stale bool, rwIntervalMs float64) {
	t.mu.Lock()
	known := t.knownVersion[key]
	writeTime, hasWrite := t.lastWriteTime[key]
	t.mu.Unlock()

	// Stale: client wrote a newer version than what the node returned.
	stale = known > 0 && returnedVersion < known
	rwIntervalMs = -1
	if hasWrite {
		rwIntervalMs = float64(time.Since(writeTime).Milliseconds())
	}
	return
}

// ---- HTTP helpers -----------------------------------------------------------

var httpClient = &http.Client{Timeout: 10 * time.Second}

func doSet(nodeURL, key, value string) (version int64, status int, latencyMs float64) {
	body, _ := json.Marshal(map[string]string{"key": key, "value": value})
	start := time.Now()
	resp, err := httpClient.Post(nodeURL+"/set", "application/json", bytes.NewReader(body))
	latencyMs = float64(time.Since(start).Milliseconds())
	if err != nil {
		return 0, 0, latencyMs
	}
	defer resp.Body.Close()
	var sr struct {
		Version int64 `json:"version"`
	}
	json.NewDecoder(resp.Body).Decode(&sr)
	return sr.Version, resp.StatusCode, latencyMs
}

func doGet(nodeURL, key string) (version int64, status int, latencyMs float64) {
	start := time.Now()
	resp, err := httpClient.Get(nodeURL + "/get/" + key)
	latencyMs = float64(time.Since(start).Milliseconds())
	if err != nil {
		return 0, 0, latencyMs
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, resp.StatusCode, latencyMs
	}
	var kv struct {
		Version int64 `json:"version"`
	}
	json.NewDecoder(resp.Body).Decode(&kv)
	return kv.Version, resp.StatusCode, latencyMs
}

// ---- main -------------------------------------------------------------------

func main() {
	leaderURL := flag.String("leader", "", "Leader URL for writes (omit for leaderless)")
	nodesStr := flag.String("nodes", "", "Comma-separated node URLs for reads (required)")
	mode := flag.String("mode", "unknown", "Mode label: w5r1 | w1r5 | w3r3 | leaderless")
	writePct := flag.Int("write-pct", 10, "Percentage of requests that are writes (1/10/50/90)")
	duration := flag.Int("duration", 30, "Test duration in seconds")
	workers := flag.Int("workers", 10, "Concurrent goroutines")
	keys := flag.Int("keys", 100, "Key-space size — smaller means more local-in-time clustering")
	output := flag.String("output", "results.csv", "Output CSV file path")
	flag.Parse()

	if *nodesStr == "" {
		log.Fatal("-nodes is required")
	}
	nodes := strings.Split(*nodesStr, ",")

	tk := newTracker()
	var mu sync.Mutex
	var results []record

	deadline := time.Now().Add(time.Duration(*duration) * time.Second)
	var wg sync.WaitGroup

	for w := 0; w < *workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rng := rand.New(rand.NewSource(time.Now().UnixNano()))

			for time.Now().Before(deadline) {
				key := fmt.Sprintf("key%d", rng.Intn(*keys))
				isWrite := rng.Intn(100) < *writePct

				var r record
				r.ts = time.Now()
				r.key = key

				if isWrite {
					// Writes go to the leader; if no leader (leaderless) pick randomly.
					writeURL := *leaderURL
					if writeURL == "" {
						writeURL = nodes[rng.Intn(len(nodes))]
					}
					value := strconv.FormatInt(time.Now().UnixNano(), 10)
					version, status, latency := doSet(writeURL, key, value)
					if status != http.StatusCreated {
						continue
					}
					tk.recordWrite(key, version)
					r.reqType = "write"
					r.latencyMs = latency
					r.version = version
					r.status = status
				} else {
					// Reads go to a random node (exposes stale reads from followers).
					readURL := nodes[rng.Intn(len(nodes))]
					version, status, latency := doGet(readURL, key)
					if status == 0 {
						continue
					}
					stale, rwInterval := tk.checkRead(key, version)
					r.reqType = "read"
					r.latencyMs = latency
					r.version = version
					r.status = status
					r.stale = stale
					r.rwIntervalMs = rwInterval
				}

				mu.Lock()
				results = append(results, r)
				mu.Unlock()
			}
		}()
	}

	wg.Wait()
	writeCSV(*output, *mode, results)
	printSummary(*mode, results)
}

// ---- output -----------------------------------------------------------------

func writeCSV(path, mode string, results []record) {
	f, err := os.Create(path)
	if err != nil {
		log.Fatalf("create %s: %v", path, err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	w.Write([]string{"mode", "type", "key", "latency_ms", "status", "stale", "version", "rw_interval_ms"})
	for _, r := range results {
		w.Write([]string{
			mode,
			r.reqType,
			r.key,
			fmt.Sprintf("%.2f", r.latencyMs),
			strconv.Itoa(r.status),
			strconv.FormatBool(r.stale),
			strconv.FormatInt(r.version, 10),
			fmt.Sprintf("%.2f", r.rwIntervalMs),
		})
	}
	w.Flush()
	log.Printf("results written to %s (%d records)", path, len(results))
}

func printSummary(mode string, results []record) {
	var readLat, writeLat []float64
	staleCount := 0

	for _, r := range results {
		if r.reqType == "write" {
			writeLat = append(writeLat, r.latencyMs)
		} else {
			readLat = append(readLat, r.latencyMs)
			if r.stale {
				staleCount++
			}
		}
	}

	fmt.Printf("\n========== Load Test: mode=%s ==========\n", mode)
	fmt.Printf("Total requests : %d\n", len(results))
	fmt.Printf("Writes         : %d\n", len(writeLat))
	fmt.Printf("Reads          : %d\n", len(readLat))
	if len(writeLat) > 0 {
		fmt.Printf("Write latency  : avg=%.1fms  p50=%.1fms  p95=%.1fms  p99=%.1fms  max=%.1fms\n",
			avg(writeLat), pct(writeLat, 50), pct(writeLat, 95), pct(writeLat, 99), pct(writeLat, 100))
	}
	if len(readLat) > 0 {
		fmt.Printf("Read latency   : avg=%.1fms  p50=%.1fms  p95=%.1fms  p99=%.1fms  max=%.1fms\n",
			avg(readLat), pct(readLat, 50), pct(readLat, 95), pct(readLat, 99), pct(readLat, 100))
		fmt.Printf("Stale reads    : %d / %d  (%.2f%%)\n",
			staleCount, len(readLat), float64(staleCount)/float64(len(readLat))*100)
	}
}

func avg(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range vals {
		sum += v
	}
	return sum / float64(len(vals))
}

func pct(vals []float64, p int) float64 {
	if len(vals) == 0 {
		return 0
	}
	sorted := make([]float64, len(vals))
	copy(sorted, vals)
	sort.Float64s(sorted)
	if p >= 100 {
		return sorted[len(sorted)-1]
	}
	idx := int(float64(p) / 100 * float64(len(sorted)))
	return sorted[idx]
}
