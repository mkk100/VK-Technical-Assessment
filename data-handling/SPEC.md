Run-level summary info
- Partially successful run if at least one of the sources failed
- Retrying transient upstream error in source B and rate limit per second error in source C, both are solved by waiting a couple seconds before retrying
- Dropping malformed records (What constitutes malformed records?)
- Dupes and conflicting records
- How should timeouts or run limits be handled?