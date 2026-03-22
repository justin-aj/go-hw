#!/usr/bin/env python3
"""
combine_results.py — merges mysql_test_results.json + dynamodb_test_results.json
into combined_results.json with a 'backend' field added to each record.

Also prints the comparison table required for STEP III.

Usage:
    python test/combine_results.py
"""

import json
import statistics

FILES = {
    "mysql":  "mysql_test_results.json",
    "dynamo": "dynamodb_test_results.json",
}

combined = []
by_backend: dict[str, list[dict]] = {}

for backend, filename in FILES.items():
    try:
        with open(filename) as f:
            records = json.load(f)
    except FileNotFoundError:
        print(f"WARNING: {filename} not found — skipping")
        continue
    for r in records:
        r["backend"] = backend
        combined.append(r)
    by_backend[backend] = records

with open("combined_results.json", "w") as f:
    json.dump(combined, f, indent=2)
print(f"Wrote combined_results.json ({len(combined)} records)\n")


def percentile(data: list[float], p: float) -> float:
    s = sorted(data)
    idx = int(len(s) * p / 100)
    return s[min(idx, len(s) - 1)]


# ── Overall comparison table ──────────────────────────────────────────────────
print("Overall Comparison (all operations)")
print(f"{'Metric':<28} {'MySQL':>10} {'DynamoDB':>10} {'Winner':>10}")
print("-" * 62)

metrics: dict[str, dict] = {}
for backend, records in by_backend.items():
    times = [r["response_time"] for r in records]
    ok    = sum(1 for r in records if r["success"])
    metrics[backend] = {
        "avg":     statistics.mean(times),
        "p50":     percentile(times, 50),
        "p95":     percentile(times, 95),
        "p99":     percentile(times, 99),
        "success": ok / len(records) * 100,
    }

rows = [
    ("Avg Response Time (ms)", "avg"),
    ("P50 Response Time (ms)", "p50"),
    ("P95 Response Time (ms)", "p95"),
    ("P99 Response Time (ms)", "p99"),
    ("Success Rate (%)",       "success"),
]
for label, key in rows:
    m  = metrics.get("mysql",  {}).get(key, float("nan"))
    d  = metrics.get("dynamo", {}).get(key, float("nan"))
    winner = "MySQL" if m < d else "DynamoDB"
    print(f"  {label:<26} {m:>10.1f} {d:>10.1f} {winner:>10}")

# ── Per-operation breakdown ───────────────────────────────────────────────────
print("\nPer-Operation Breakdown")
print(f"{'Operation':<14} {'MySQL Avg':>12} {'DynamoDB Avg':>14} {'Faster By':>12}")
print("-" * 56)

ops = ["create_cart", "add_items", "get_cart"]
for op in ops:
    avgs = {}
    for backend, records in by_backend.items():
        times = [r["response_time"] for r in records if r["operation"] == op]
        avgs[backend] = statistics.mean(times) if times else float("nan")
    m = avgs.get("mysql",  float("nan"))
    d = avgs.get("dynamo", float("nan"))
    diff = abs(m - d)
    faster = "MySQL" if m < d else "DynamoDB"
    print(f"  {op:<12} {m:>12.1f} {d:>14.1f}   {faster} by {diff:.1f}ms")
