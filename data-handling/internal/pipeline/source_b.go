package pipeline

import "encoding/json"

type sourceBcursor struct {
	Items []struct {
		Sku         string          `json:"sku"`
		Title       string          `json:"title"`
		AmountCents json.RawMessage `json:"amount_cents"`
		Department  string          `json:"department"`
	} `json:"items"`
	NextCursor string `json:"next_cursor"`
}

// FetchSourceB follows cursor pagination. amount_cents may be null or a
// non-numeric string; those records are dropped as malformed.
func FetchSourceB() ([]Product, error) {
	var out []Product
	cursor := ""

	for {
		url := BaseURL + "/source-b/products"
		if cursor != "" {
			url += "?cursor=" + cursor
		}

		var p sourceBcursor
		if err := getJSON(url, &p); err != nil {
			return nil, err
		}

		for _, it := range p.Items {
			cents, ok := parseCents(it.AmountCents)
			if !ok || it.Sku == "" {
				dropped["source_b"]++
				logger.Warn("dropping malformed record", "source", "source_b", "sku", it.Sku, "reason", "bad amount_cents or missing sku")
				continue
			}
			out = append(out, Product{
				ID:       it.Sku,
				Name:     it.Title,
				Source:   "source_b",
				Price:    float64(cents) / 100.0,
				Category: it.Department,
			})
		}

		if p.NextCursor == "" {
			break
		}
		cursor = p.NextCursor
	}
	return out, nil
}

// parseCents accepts only a JSON integer; null, strings and floats are rejected.
func parseCents(raw json.RawMessage) (int, bool) {
	if len(raw) == 0 || string(raw) == "null" {
		return 0, false
	}
	var n int
	if err := json.Unmarshal(raw, &n); err != nil {
		return 0, false
	}
	return n, true
}
