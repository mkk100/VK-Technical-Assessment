package pipeline

import (
	"fmt"
	"strconv"
)

type sourceCoffset struct {
	Data []struct {
		ProductID   string `json:"product_id"`
		ProductName string `json:"product_name"`
		Price       string `json:"price"` // decimal string
		Type        string `json:"type"`
	} `json:"data"`
	NextOffset *int `json:"next_offset"`
}

// FetchSourceC walks offset/limit pagination. The source is rate limited to
// 2 req/s; getJSON absorbs the resulting 429s via retry.
func FetchSourceC() ([]Product, error) {
	var out []Product
	offset := 0

	for {
		url := fmt.Sprintf("%s/source-c/products?offset=%d&limit=2", BaseURL, offset)

		var p sourceCoffset
		if err := getJSON(url, &p); err != nil {
			return nil, err
		}

		for _, it := range p.Data {
			price, err := strconv.ParseFloat(it.Price, 64)
			if err != nil || it.ProductID == "" {
				dropped["source_c"]++
				logger.Warn("dropping malformed record", "source", "source_c", "id", it.ProductID, "price", it.Price, "reason", "unparseable price or missing id")
				continue
			}
			out = append(out, Product{
				ID:       it.ProductID,
				Name:     it.ProductName,
				Source:   "source_c",
				Price:    price,
				Category: it.Type,
			})
		}

		if p.NextOffset == nil {
			break
		}
		offset = *p.NextOffset
	}
	return out, nil
}
