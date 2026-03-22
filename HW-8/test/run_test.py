#!/usr/bin/env python3
"""
HW-8 150-operation test script.

Runs exactly:
  50 × POST /shopping-carts       (create_cart)
  50 × POST /shopping-carts/{id}/items (add_items)
  50 × GET  /shopping-carts/{id}  (get_cart)

Each result is written as one JSON object per line to the output file.
The script then merges them into the required flat JSON array.

Usage:
    python test/run_test.py <ALB_URL> <backend>

Examples:
    python test/run_test.py http://hw8-cart-alb-123456789.us-east-1.elb.amazonaws.com mysql
    python test/run_test.py http://hw8-cart-alb-123456789.us-east-1.elb.amazonaws.com dynamo
"""

import sys
import json
import time
import random
import datetime
import urllib.request
import urllib.error

# ── Config ────────────────────────────────────────────────────────────────────

N = 50  # operations per type

# ── Helpers ───────────────────────────────────────────────────────────────────

def now_iso() -> str:
    return datetime.datetime.utcnow().strftime("%Y-%m-%dT%H:%M:%SZ")


def post(base_url: str, path: str, body: dict) -> tuple[int, dict]:
    data = json.dumps(body).encode()
    req = urllib.request.Request(
        f"{base_url}{path}",
        data=data,
        headers={"Content-Type": "application/json"},
        method="POST",
    )
    try:
        with urllib.request.urlopen(req, timeout=10) as resp:
            raw = resp.read()
            payload = json.loads(raw) if raw else {}
            return resp.status, payload
    except urllib.error.HTTPError as e:
        raw = e.read()
        return e.code, {}


def get(base_url: str, path: str) -> tuple[int, dict]:
    req = urllib.request.Request(f"{base_url}{path}", method="GET")
    try:
        with urllib.request.urlopen(req, timeout=10) as resp:
            raw = resp.read()
            return resp.status, json.loads(raw) if raw else {}
    except urllib.error.HTTPError as e:
        return e.code, {}


def record(operation: str, start: float, status_code: int, success: bool) -> dict:
    return {
        "operation":     operation,
        "response_time": round((time.perf_counter() - start) * 1000, 2),  # ms
        "success":       success,
        "status_code":   status_code,
        "timestamp":     now_iso(),
    }


# ── Main ──────────────────────────────────────────────────────────────────────

def run(base_url: str, backend: str) -> list[dict]:
    results = []
    cart_ids: list[int] = []

    print(f"Target : {base_url}")
    print(f"Backend: {backend}")
    print(f"Running {N} × create_cart …")

    # ── Phase 1: create 50 carts ─────────────────────────────────────────────
    for i in range(N):
        customer_id = random.randint(1, 100_000)
        t = time.perf_counter()
        code, body = post(base_url, "/shopping-carts", {"customer_id": customer_id})
        success = code == 201
        results.append(record("create_cart", t, code, success))
        if success and "shopping_cart_id" in body:
            cart_ids.append(body["shopping_cart_id"])
        if (i + 1) % 10 == 0:
            print(f"  created {i + 1}/{N}  (last status={code})")

    if not cart_ids:
        print("ERROR: no carts were created successfully — aborting")
        sys.exit(1)

    print(f"Running {N} × add_items …")

    # ── Phase 2: add items to 50 carts ───────────────────────────────────────
    for i in range(N):
        cart_id = random.choice(cart_ids)
        body = {
            "product_id": random.randint(1, 1000),
            "quantity":   random.randint(1, 5),
        }
        t = time.perf_counter()
        code, _ = post(base_url, f"/shopping-carts/{cart_id}/items", body)
        success = code == 204
        results.append(record("add_items", t, code, success))
        if (i + 1) % 10 == 0:
            print(f"  added {i + 1}/{N}  (last status={code})")

    print(f"Running {N} × get_cart …")

    # ── Phase 3: retrieve 50 carts ───────────────────────────────────────────
    for i in range(N):
        cart_id = random.choice(cart_ids)
        t = time.perf_counter()
        code, _ = get(base_url, f"/shopping-carts/{cart_id}")
        success = code == 200
        results.append(record("get_cart", t, code, success))
        if (i + 1) % 10 == 0:
            print(f"  retrieved {i + 1}/{N}  (last status={code})")

    return results


def summarise(results: list[dict]) -> None:
    by_op: dict[str, list[float]] = {}
    for r in results:
        by_op.setdefault(r["operation"], []).append(r["response_time"])

    print("\n── Summary ──────────────────────────────────────────────────")
    for op, times in sorted(by_op.items()):
        times_sorted = sorted(times)
        n = len(times_sorted)
        avg = sum(times_sorted) / n
        p50 = times_sorted[int(n * 0.50)]
        p95 = times_sorted[int(n * 0.95)]
        p99 = times_sorted[min(int(n * 0.99), n - 1)]
        print(f"  {op:<12}  avg={avg:6.1f}ms  p50={p50:6.1f}ms  "
              f"p95={p95:6.1f}ms  p99={p99:6.1f}ms  n={n}")

    total = len(results)
    ok = sum(1 for r in results if r["success"])
    print(f"\n  Total: {total}  Success: {ok}  Failure: {total - ok}")
    print("─────────────────────────────────────────────────────────────\n")


if __name__ == "__main__":
    if len(sys.argv) < 3:
        print("Usage: python run_test.py <ALB_URL> <backend>")
        sys.exit(1)

    base_url = sys.argv[1].rstrip("/")
    backend  = sys.argv[2]          # "mysql" or "dynamo"
    out_file = f"{backend}_test_results.json"

    results = run(base_url, backend)
    summarise(results)

    with open(out_file, "w") as f:
        json.dump(results, f, indent=2)

    print(f"Results saved → {out_file}  ({len(results)} records)")
