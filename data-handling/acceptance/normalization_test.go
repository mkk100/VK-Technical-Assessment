package acceptance

import (
	"net/http"
	"testing"

	"data-handling/internal/pipeline"
)

// 1. Normalization per source.
//
// Each source has a different wire schema. A wrong field mapping produces
// valid-looking JSON with wrong data. Feed one canned page per source, assert
// the exact Product out.
func TestNormalization_PerSource(t *testing.T) {
	t.Run("source_a maps id/name/price/category and stamps source", func(t *testing.T) {
		serve(t, func(w http.ResponseWriter, r *http.Request) {
			write(w, http.StatusOK, `{"page":1,"total_pages":1,"products":[
				{"id":"a-1","name":"Keyboard","price":89.99,"category":"electronics"}]}`)
		})

		got, err := pipeline.FetchSourceA()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := pipeline.Product{ID: "a-1", Name: "Keyboard", Source: "source_a", Price: 89.99, Category: "electronics"}
		if len(got) != 1 || got[0] != want {
			t.Fatalf("got %+v, want %+v", got, want)
		}
	})

	t.Run("source_b converts amount_cents 3499 -> price 34.99", func(t *testing.T) {
		serve(t, func(w http.ResponseWriter, r *http.Request) {
			write(w, http.StatusOK, `{"items":[
				{"sku":"b-1","title":"Desk Lamp","amount_cents":3499,"department":"home"}],
				"next_cursor":null}`)
		})

		got, err := pipeline.FetchSourceB()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := pipeline.Product{ID: "b-1", Name: "Desk Lamp", Source: "source_b", Price: 34.99, Category: "home"}
		if len(got) != 1 || got[0] != want {
			t.Fatalf("got %+v, want %+v", got, want)
		}
	})

	t.Run("source_c parses string price \"49.50\" -> 49.5", func(t *testing.T) {
		serve(t, func(w http.ResponseWriter, r *http.Request) {
			write(w, http.StatusOK, `{"data":[
				{"product_id":"c-1","product_name":"USB-C Hub","price":"49.50","type":"electronics"}],
				"next_offset":null}`)
		})

		got, err := pipeline.FetchSourceC()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := pipeline.Product{ID: "c-1", Name: "USB-C Hub", Source: "source_c", Price: 49.5, Category: "electronics"}
		if len(got) != 1 || got[0] != want {
			t.Fatalf("got %+v, want %+v", got, want)
		}
	})
}
