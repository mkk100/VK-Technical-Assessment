package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

func main() {
	start := time.Now()

	pA, rA := runSource("source_a", fetchSourceA)
	pB, rB := runSource("source_b", fetchSourceB)
	pC, rC := runSource("source_c", fetchSourceC)

	var all []Product
	all = append(all, pA...)
	all = append(all, pB...)
	all = append(all, pC...)
	all = dedupe(all)

	// run summary -> stderr
	summary := emitSummary([]SourceResult{rA, rB, rC}, time.Since(start), len(all))

	// products -> stdout (the deliverable)
	out, err := json.MarshalIndent(all, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(out))

	if summary.PartialRun {
		os.Exit(2) // ran, but at least one source failed
	}
}
