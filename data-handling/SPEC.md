## Goals
- Normalized records
- Handles upstream failure
- Handles rate limits
- Provide run-level summary
- Provide observability
- Tests 

## Non-Goals
- Duplication in data
- One data field or one source failed and operation comes to a halt.
- Bad testing, no run level summary and no observability 

## Important requirements
- Fetch all 3 sources and iterate every page
- Proper handling for transient errors and rate limits
- normalized data
- Despite the source failing, program should still give a partial result

## Normalized data, every data should follow this format.
  {
    "id": "b-203",
    "name": "Notebook Set",
    "source": "source_b",
    "price": 18.99,
    "category": "office"
  }

## Failure behavior
- if transient error, retry with backoff
- if permanent(4xx), fail source
- if malformed record, drop record
- if retries exhausted, only fail that source

## Assumptions
- Partially successful run if at least one of the sources failed
- Retrying transient upstream error in source B and rate limit per second error in source C, both are solved by waiting a couple seconds before retrying
- For now, malformed records is the one with invalid cost specification. What constitutes malformed records is the more rigorous question I can ask if given more time
- Dupes and conflicting records (first dupe gets it, again more discussion to be had if it's a good rule given more time)
- When one source fails completely, we still run other sources but we flag it in the log as partial_run = true
- Failures are handled by waiting for a few seconds before rerunning the GET request. 

## Testable Acceptance Criteria
- Given healthy sources, output should contain every product except the malformed ones.
- Break the loop if the pages are null.
- If a source gives 429 and then 200, the request has been retried and final result is complete
- If a source gives 429, retry after 1s


