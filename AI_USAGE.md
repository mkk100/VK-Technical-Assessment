# AI Usage

## Tools
Claude Code

## What I delegated

- Back and forth to set up sourceA, sourceB, sourceC go files.
- Code restructure
- Run-level summary, Observability and Duplication
- The acceptance tests once I had settled the behavior I wanted.

## Spec review feedback and how I responded

Asking the AI to check the work against the challenge surfaced several gaps I acted on:

- No overall run deadline only a per-request timeout existed. Documented as a known gap in SPEC/PLAN rather than rushing a half-done `context` change.
- Thin test coverage — only dedupe was tested at first. Added acceptance tests for normalization, malformed records, retry, partial failure, and pagination.
- source field missing from the Product.


## How I verified AI-generated work

- Run the program every time it makes a change, make sure genral flow of code makes sense.
- Actually manually count the results and check the logs and observablity.
- See if it gives me the desired result

## Something I challenged / rejected

- The AI not knowing how to call the sources and cursors in source B and C. I have to intervene and challenge its codes to better fit the mock api calls.

## Final review findings

- `float64` for currency flagged again as unsafe for real money; kept for the exercise.
- Use context deadline instead of per-request timeout and bounded retries

## With more time
- `requests` / `retries` in the summary
- concurrent fetch
