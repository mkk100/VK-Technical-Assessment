package acceptance

import (
	"bytes"
	"net/http"
	"strings"
	"testing"
	"time"

	"data-handling/internal/pipeline"
)

// 4. Partial failure.
//
// The spec's headline requirement: "handle partial upstream failure without
// losing successful results." If source B throwing takes A and C down with it,
// the main ask has failed. Stub B to return 503; A and C return data.
func TestPartialFailure_SuccessfulSourcesSurvive(t *testing.T) {
	serve(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/source-a/"):
			write(w, http.StatusOK, `{"page":1,"total_pages":1,"products":[
				{"id":"a-1","name":"A","price":1,"category":"c"}]}`)
		case strings.HasPrefix(r.URL.Path, "/source-b/"):
			write(w, http.StatusServiceUnavailable, `{"error":"upstream_unavailable"}`)
		case strings.HasPrefix(r.URL.Path, "/source-c/"):
			write(w, http.StatusOK, `{"data":[
				{"product_id":"c-1","product_name":"C","price":"2.00","type":"t"}],
				"next_offset":null}`)
		default:
			write(w, http.StatusNotFound, `{}`)
		}
	})

	pA, rA := pipeline.RunSource("source_a", pipeline.FetchSourceA)
	pB, rB := pipeline.RunSource("source_b", pipeline.FetchSourceB)
	pC, rC := pipeline.RunSource("source_c", pipeline.FetchSourceC)

	// A and C succeeded and kept their products.
	if !rA.OK || len(pA) != 1 {
		t.Fatalf("source_a should have survived: ok=%v products=%v", rA.OK, pA)
	}
	if !rC.OK || len(pC) != 1 {
		t.Fatalf("source_c should have survived: ok=%v products=%v", rC.OK, pC)
	}

	// B failed cleanly: no products, not OK, error recorded.
	if rB.OK || len(pB) != 0 || rB.Error == "" {
		t.Fatalf("source_b should have failed with an error: %+v", rB)
	}

	var all []pipeline.Product
	all = append(all, pA...)
	all = append(all, pC...)

	var buf bytes.Buffer
	summary := pipeline.EmitSummary(&buf, []pipeline.SourceResult{rA, rB, rC}, time.Second, len(all))

	// partial_run drives exit code 2 in main().
	if !summary.PartialRun {
		t.Fatal("summary.PartialRun must be true when a source failed")
	}
	if summary.TotalProducts != 2 {
		t.Fatalf("expected 2 surviving products in summary, got %d", summary.TotalProducts)
	}
}
