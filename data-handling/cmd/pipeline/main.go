// Command pipeline aggregates product data from the three mock upstream APIs,
// normalizes it, and writes a JSON product list to stdout plus a run summary
// (and structured logs) to stderr.
//
// Exit codes: 0 = full success, 1 = could not write output, 2 = partial run
// (at least one source failed).
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"data-handling/internal/pipeline"
)

func main() {
	start := time.Now()

	pA, rA := pipeline.RunSource("source_a", pipeline.FetchSourceA)
	pB, rB := pipeline.RunSource("source_b", pipeline.FetchSourceB)
	pC, rC := pipeline.RunSource("source_c", pipeline.FetchSourceC)

	var all []pipeline.Product
	all = append(all, pA...)
	all = append(all, pB...)
	all = append(all, pC...)
	all = pipeline.Dedupe(all)

	summary := pipeline.EmitSummary(os.Stderr, []pipeline.SourceResult{rA, rB, rC}, time.Since(start), len(all))

	out, err := json.MarshalIndent(all, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, "marshal products failed:", err)
		os.Exit(1)
	}
	fmt.Println(string(out))

	if summary.PartialRun {
		os.Exit(2)
	}
}
