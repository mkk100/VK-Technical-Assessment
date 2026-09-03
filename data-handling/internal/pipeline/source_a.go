package pipeline

import "fmt"

type sourceAPage struct {
	Page       int `json:"page"`
	TotalPages int `json:"total_pages"`
	Products   []struct {
		ID       string  `json:"id"`
		Name     string  `json:"name"`
		Price    float64 `json:"price"`
		Category string  `json:"category"`
	} `json:"products"`
}

// FetchSourceA walks page-based pagination and returns normalized products.
func FetchSourceA() ([]Product, error) {
	var out []Product

	for page := 1; ; page++ {
		url := fmt.Sprintf("%s/source-a/products?page=%d", BaseURL, page)

		var p sourceAPage
		if err := getJSON(url, &p); err != nil {
			return nil, err
		}

		for _, it := range p.Products {
			if it.ID == "" {
				dropped["source_a"]++
				logger.Warn("dropping malformed record", "source", "source_a", "reason", "empty id")
				continue
			}
			out = append(out, Product{
				ID:       it.ID,
				Name:     it.Name,
				Source:   "source_a",
				Price:    it.Price,
				Category: it.Category,
			})
		}

		if page >= p.TotalPages {
			break
		}
	}
	return out, nil
}
