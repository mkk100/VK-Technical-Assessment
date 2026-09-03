# Reliable Data Pipeline

Aggregates product data from three mock upstream APIs, normalizes it into one
schema, and writes a consolidated JSON list plus a run summary.

See [`data-handling/SPEC.md`](data-handling/SPEC.md) for behavior,
[`PLAN.md`](PLAN.md) for approach and tradeoffs.

## Language

**Go** (stdlib only, no dependencies). Chosen over Python because static typing
catches the per-source schema-mapping mistakes this exercise centers on, and the
stdlib covers HTTP, retry, and JSON without a framework.

## Setup

- Go 1.24+
- Python 3 (only to run the provided mock service)

## Run

Start the mock API (terminal 1):

```sh
cd mock-api-service
python3 server.py --port 8080
```

Run the pipeline (terminal 2):

```sh
cd data-handling
go run ./cmd/pipeline
```

Optional flags / env: `MOCK_BASE_URL` (default `http://localhost:8080`).

### Output

- **stdout** — a JSON array of normalized products (the deliverable):

  ```json
  [
    { "id": "a-101", "name": "Mechanical Keyboard", "source": "source_a",
      "price": 89.99, "category": "electronics" }
  ]
  ```

- **stderr** — structured logs (one line per retry / dropped record / source
  start-finish) followed by a run summary:

  ```json
  {
    "elapsed": "1.9s",
    "total_products": 17,
    "partial_run": false,
    "sources": [
      { "source": "source_b", "ok": true, "kept": 5, "dropped": 1, "elapsed": "380ms" }
    ]
  }
  ```

- **Exit codes** — `0` full success, `1` could not write output, `2` partial run
  (at least one source failed; other sources' data is still emitted).

## Test

```sh
cd data-handling
go test ./...
```

No mock service needed — tests drive the fetchers against in-process
`httptest` servers. `data-handling/acceptance/` holds the behavior tests
(normalization, malformed records, retry, partial failure, pagination);
`internal/pipeline/` holds unit tests.

## Important assumptions

- One source failing is a **partial run**, not a total failure — the other
  sources' products are still returned.
- Malformed individual records are dropped (counted, logged), never fatal.
  "Malformed" = missing id or unparseable/absent price.
- Dedup identity is source-scoped (`source|id`), first occurrence wins. The same
  real product from two sources is treated as two products.
- `next_cursor` / `next_offset` are opaque and followed by value; `total_pages`
  is trusted as source A's stop signal.
- The mock is deterministic and re-armed via `POST /admin/reset` between runs.

## Known limitations

- **No overall run deadline** — bounded only by per-request timeout (5s) x
  retries x pages. A `context` deadline is the intended fix.
- **Prices are `float64`** — fine here, not safe for real currency (rounding).
- **Retry/request counts not in the summary** — visible in logs only.
- **No cross-source deduplication / conflict resolution.**

## Not implemented (time limit)

- Concurrent source fetching (`errgroup`) — code is structured for it
- `context`-based run budget
- Circuit breaker for a persistently-down source
- Shared `Product` validator so "malformed" is defined once
