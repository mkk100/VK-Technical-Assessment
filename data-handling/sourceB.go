package main

import (
	"encoding/json"
	"fmt"
	"log"
)

type sourceBcursor struct {
	Items   []struct {
		Sku   string `json:"sku"`
		Title  string `json:"title"`
		Amount_Cents json.RawMessage `json:"amount_cents"`
		Department string `json:"department"`
	} `json:"items"`
	Next_Cursor string `json:"next_cursor"`
}

func fetchSourceB() ([]Product, error) {
	var out []Product

	for cursor := 1; ; cursor++ {
		url := baseURL + "/source-b/products"

		if cursor > 1 {
			url += fmt.Sprintf("?cursor=cursor-%d", cursor)
		}

		var p sourceBcursor
		if err := getJSON(url, &p); err != nil {
			return nil, err
		}
		
		for _, it := range p.Items {
			cents, ok := parseCents(it.Amount_Cents) // parse the invalid Amount_Cents variable
			if !ok || it.Sku == "" {
				dropped["source_b"]++
				log.Printf("source_b: dropping malformed record sku=%q", it.Sku)
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

		if p.Next_Cursor == "" {
			break
		}
	}
	return out, nil
}


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

