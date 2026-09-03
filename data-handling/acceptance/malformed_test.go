package acceptance

import (
	"net/http"
	"testing"

	"data-handling/internal/pipeline"
)

// 2. Malformed record handling.
//
// "One bad record kills the batch" is the classic data-pipeline failure. A
// regression here silently drops good data or crashes the run. Feed a page
// mixing valid records with every malformed variant and assert the accounting.
//
// The invariant fetched == kept + dropped is the highest-value assertion: if it
// fails, records are being lost without anyone knowing.
func TestMalformed_RecordsDroppedNotFatal(t *testing.T) {
	const fetched = 5 // records the server returns below

	serve(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("cursor") == "" {
			write(w, http.StatusOK, `{"items":[
				{"sku":"b-1","title":"Valid","amount_cents":1000,"department":"home"},
				{"sku":"b-2","title":"NonNumeric","amount_cents":"not-a-number","department":"home"},
				{"sku":"b-3","title":"Null","amount_cents":null,"department":"home"}
			],"next_cursor":"tok-2"}`)
			return
		}
		write(w, http.StatusOK, `{"items":[
			{"sku":"","title":"MissingSKU","amount_cents":2000,"department":"home"},
			{"sku":"b-5","title":"Valid","amount_cents":2499,"department":"kitchen"}
		],"next_cursor":null}`)
	})

	products, res := pipeline.RunSource("source_b", pipeline.FetchSourceB)

	if !res.OK {
		t.Fatalf("source should not fail on malformed records: %+v", res)
	}
	if res.Kept != 2 {
		t.Fatalf("expected 2 kept, got %d (%+v)", res.Kept, products)
	}
	if res.Dropped != 3 {
		t.Fatalf("expected 3 dropped (non-numeric, null, missing sku), got %d", res.Dropped)
	}
	if res.Kept+res.Dropped != fetched {
		t.Fatalf("record accounting broken: kept(%d) + dropped(%d) != fetched(%d)",
			res.Kept, res.Dropped, fetched)
	}

	// The good records must survive intact.
	ids := []string{products[0].ID, products[1].ID}
	if ids[0] != "b-1" || ids[1] != "b-5" {
		t.Fatalf("wrong records kept: %v", ids)
	}
	if products[0].Price != 10.0 {
		t.Fatalf("valid record price wrong: %v", products[0].Price)
	}
}
