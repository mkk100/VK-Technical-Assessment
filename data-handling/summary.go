package main

import (
	"encoding/json"
	"os"
	"time"
)

// SourceResult is the per-source outcome of one run.
type SourceResult struct {
	Source  string `json:"source"`
	OK      bool   `json:"ok"`
	Kept    int    `json:"kept"`    // products normalized cleanly
	Dropped int    `json:"dropped"` // malformed records skipped
	Error   string `json:"error,omitempty"`
	Elapsed string `json:"elapsed"`
}

// RunSummary is the top-level observability object, written to stderr.
type RunSummary struct {
	Elapsed       string         `json:"elapsed"`
	TotalProducts int            `json:"total_products"` // after dedupe
	PartialRun    bool           `json:"partial_run"`    // at least one source failed
	Sources       []SourceResult `json:"sources"`
}

// dropped counts malformed records skipped per source. The fetchers write to it;
// runSource reads it back into each SourceResult.
var dropped = map[string]int{}

// runSource times a fetcher and captures its result.
func runSource(name string, fn func() ([]Product, error)) ([]Product, SourceResult) {
	start := time.Now()
	items, err := fn()
	res := SourceResult{
		Source:  name,
		OK:      err == nil,
		Kept:    len(items),
		Dropped: dropped[name],
		Elapsed: time.Since(start).Round(time.Millisecond).String(),
	}
	if err != nil {
		res.Error = err.Error()
	}
	return items, res
}

// emitSummary assembles the RunSummary and writes it to stderr as JSON.
func emitSummary(sources []SourceResult, elapsed time.Duration, totalProducts int) RunSummary {
	s := RunSummary{
		Elapsed:       elapsed.Round(time.Millisecond).String(),
		TotalProducts: totalProducts,
		Sources:       sources,
	}
	for _, r := range sources {
		if !r.OK {
			s.PartialRun = true
		}
	}
	enc := json.NewEncoder(os.Stderr)
	enc.SetIndent("", "  ")
	enc.Encode(s)
	return s
}
