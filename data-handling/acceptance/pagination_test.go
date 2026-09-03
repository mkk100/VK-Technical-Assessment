package acceptance

import (
	"fmt"
	"net/http"
	"sync/atomic"
	"testing"

	"data-handling/internal/pipeline"
)

// 5. Pagination termination.
//
// A wrong stop condition (page >= total_pages off-by-one, a cursor that never
// empties) is an infinite loop that hangs the run. Assert each source walks
// every page, collects every record, and stops.

func TestPagination_SourceA_WalksAllPagesAndStops(t *testing.T) {
	var requests atomic.Int32
	serve(t, func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		page := r.URL.Query().Get("page")
		if page == "" || page > "3" {
			t.Errorf("unexpected page %q", page)
			write(w, http.StatusNotFound, `{}`)
			return
		}
		write(w, http.StatusOK, fmt.Sprintf(
			`{"page":%s,"total_pages":3,"products":[{"id":"a-%s","name":"n","price":1,"category":"c"}]}`,
			page, page))
	})

	got, err := pipeline.FetchSourceA()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n := requests.Load(); n != 3 {
		t.Fatalf("expected exactly 3 page requests, got %d", n)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 products collected, got %d", len(got))
	}
}

func TestPagination_SourceA_SinglePage(t *testing.T) {
	var requests atomic.Int32
	serve(t, func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		write(w, http.StatusOK, `{"page":1,"total_pages":1,"products":[
			{"id":"a-1","name":"n","price":1,"category":"c"}]}`)
	})

	if _, err := pipeline.FetchSourceA(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n := requests.Load(); n != 1 {
		t.Fatalf("total_pages:1 should mean 1 request, got %d", n)
	}
}

func TestPagination_SourceB_FollowsCursorUntilNull(t *testing.T) {
	var requests atomic.Int32
	serve(t, func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		switch r.URL.Query().Get("cursor") {
		case "":
			write(w, http.StatusOK, `{"items":[{"sku":"b-1","title":"n","amount_cents":100,"department":"d"}],"next_cursor":"c2"}`)
		case "c2":
			write(w, http.StatusOK, `{"items":[{"sku":"b-2","title":"n","amount_cents":100,"department":"d"}],"next_cursor":"c3"}`)
		case "c3":
			write(w, http.StatusOK, `{"items":[{"sku":"b-3","title":"n","amount_cents":100,"department":"d"}],"next_cursor":null}`)
		default:
			t.Errorf("unexpected cursor %q", r.URL.Query().Get("cursor"))
			write(w, http.StatusBadRequest, `{}`)
		}
	})

	got, err := pipeline.FetchSourceB()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n := requests.Load(); n != 3 {
		t.Fatalf("expected 3 requests following the cursor chain, got %d", n)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 products, got %d", len(got))
	}
}

func TestPagination_SourceC_FollowsOffsetUntilNull(t *testing.T) {
	var requests atomic.Int32
	serve(t, func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		switch r.URL.Query().Get("offset") {
		case "0":
			write(w, http.StatusOK, `{"data":[{"product_id":"c-1","product_name":"n","price":"1.00","type":"t"}],"next_offset":2}`)
		case "2":
			write(w, http.StatusOK, `{"data":[{"product_id":"c-2","product_name":"n","price":"1.00","type":"t"}],"next_offset":null}`)
		default:
			t.Errorf("unexpected offset %q", r.URL.Query().Get("offset"))
			write(w, http.StatusBadRequest, `{}`)
		}
	})

	got, err := pipeline.FetchSourceC()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n := requests.Load(); n != 2 {
		t.Fatalf("expected 2 requests, got %d", n)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 products, got %d", len(got))
	}
}
