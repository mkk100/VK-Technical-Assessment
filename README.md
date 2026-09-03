# Reliable Data Pipeline

Aggregates product data from three mock upstream APIs, normalizes it into one
schema, and writes a consolidated JSON list plus a run summary.

## Language
**Go** - static typing, catches the per-source schema-mapping mistakes

## Run

Start the mock API (terminal 1):

Run the pipeline (terminal 2):

```sh
cd data-handling
go run ./cmd/pipeline
```

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
- One source failing is a partial run, not a total failure — the other
  sources' products are still returned.
- Malformed individual records are dropped (counted, logged), never fatal.
- Dedup identity is source-scoped (`source|id`), first occurrence wins. 


## Known limitations

- No overall run deadline — bounded only by per-request timeout (5s) x
  retries x pages. A `context` deadline is the intended fix.
- Prices are float64 — fine here, not safe for real currency (rounding).
- Basic rigorous deduplication/ conflict resolution

## Not implemented (time limit)

- Concurrent source fetching (`errgroup`) — code is structured for it
- `context`-based run budget

