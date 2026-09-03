package pipeline

import "testing"

func TestDedupe(t *testing.T) {
	in := []Product{
		{ID: "1", Source: "source_a"},
		{ID: "1", Source: "source_a"}, // duplicate within a source -> dropped
		{ID: "1", Source: "source_b"}, // same id, different source -> kept
		{ID: "2", Source: "source_a"},
	}

	got := Dedupe(in)

	if len(got) != 3 {
		t.Fatalf("want 3 products, got %d: %+v", len(got), got)
	}
	if got[0].ID != "1" || got[0].Source != "source_a" {
		t.Fatalf("first-wins broken, got[0] = %+v", got[0])
	}
}
