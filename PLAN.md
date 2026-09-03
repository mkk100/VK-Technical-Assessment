# PLAN

## Implementation approach
Four stages, wired together in `cmd/pipeline/main.go`:
```
FETCH            NORMALIZE         COMBINE            REPORT
per-source   ->  map to Product -> dedupe         -> summary + logs
fetchers                            (main + dedupe.go)  (summary.go)
```

- internal/pipeline/ holds all logic; cmd/pipeline/ is just wiring.
- Each source has its own file (source_a.go, source_b.go, source_c.go) because the three APIs differ in pagination and schema. They share one HTTP helper (client.go) and one output type (Product).

## Major technical decisions

- Go for static typing and because it's the language I'm more familiar with.
- Sequential implementation but I would consider concurrency if more time.
- Retry: 429 + 5xx errors, backoff for 300ms, cap 4 retries, This behavior matches the mock's transient-failure and rate-limit behavior
- For malformed records, drop the record, count it, log WARN, keep going 
- Partial failure source failure is isolated, run is flagged as partial_run = true and exits 2 
- For dedupe first occurence wins
- log/slog for logging

## Important tradeoffs

- Sequential vs concurrent — chose simplicity
- float64 prices— simple, but not safe for real currency (rounding). 
- Taking the incomplete source rather than dropping the entire source due to one corrupt data field.

## Testing and verification strategy

- internal/pipeline — unit tests for `dedupe` (first-wins, source-scoped).

- acceptance/ - where tests are at
  1. Normalization per source (exact `Product` out; cents-to-dollars; string-price parse)
  2. Malformed records (dropped not fatal; `fetched == kept + dropped` invariant)
  3. Retry (429-then-200 recovers; gives up after `MaxRetries+1`; no retry on 4xx)
  4. Partial failure (one source 503s, others keep their products, `partial_run` true)
  5. Pagination termination (each style walks all pages and stops; single-page = 1 request)


## What I'd do with more time

- Try out concurrency
- Better reinforcements for malformed records and handling dupes.

