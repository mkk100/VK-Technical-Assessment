package pipeline

import (
	"encoding/json"
	"io"
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

// RunSummary is the top-level observability object.
type RunSummary struct {
	Elapsed       string         `json:"elapsed"`
	TotalProducts int            `json:"total_products"` // after dedupe
	PartialRun    bool           `json:"partial_run"`    // at least one source failed
	Sources       []SourceResult `json:"sources"`
}

// dropped counts malformed records skipped per source. The fetchers write to it;
// RunSource reads it back into each SourceResult.
var dropped = map[string]int{}

// RunSource times a fetcher, captures its result, and logs start/finish.
func RunSource(name string, fn func() ([]Product, error)) ([]Product, SourceResult) {
	logger.Info("source start", "source", name)
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
		logger.Error("source failed", "source", name, "kept", res.Kept, "dropped", res.Dropped, "elapsed", res.Elapsed, "err", err)
	} else {
		logger.Info("source done", "source", name, "kept", res.Kept, "dropped", res.Dropped, "elapsed", res.Elapsed)
	}
	return items, res
}

// EmitSummary assembles the RunSummary and writes it to w as indented JSON.
func EmitSummary(w io.Writer, sources []SourceResult, elapsed time.Duration, totalProducts int) RunSummary {
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

	logger.Info("run complete", "total_products", totalProducts, "partial_run", s.PartialRun, "elapsed", s.Elapsed)

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.Encode(s)
	return s
}
