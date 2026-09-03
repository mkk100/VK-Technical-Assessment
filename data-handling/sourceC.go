package main

import (
	"fmt"
	"log"
	"strconv"
)

type sourceCoffset struct {
	Data []struct {
		Product_id   string `json:"product_id"`
		Title  string `json:"product_name"`
		Price string `json:"price"`
		Type string `json:"type"`
	} `json:"data"`
	
	Next_Offset *int `json:"next_offset"`
	Max_Page_Size int `json:"max_page_size"`
}

func fetchSourceC() ([]Product, error) {
	var out []Product
	for pageOffset := 0; ; {
		url := fmt.Sprintf("%s/source-c/products?offset=%d&limit=2", baseURL, pageOffset)

		var p sourceCoffset
		if err := getJSON(url, &p); err != nil {
			return nil, err
		}
		
		for _, it := range p.Data {
			priceFloat, err := strconv.ParseFloat(it.Price, 64)
			if err != nil || it.Product_id == "" {
				dropped["source_c"]++
				log.Printf("source_c: dropping malformed record id=%q price=%q", it.Product_id, it.Price)
				continue
			}

			out = append(out, Product{
				ID:       it.Product_id,
				Name:     it.Title,
				Source:   "source_c",
				Price:    priceFloat,
				Category: it.Type,
			})
		}
		if p.Next_Offset == nil {
			break
		}
		pageOffset = *p.Next_Offset
	} 
	return out, nil
}

// for i in 1 2 3 4; do
//   curl -s -o /dev/null -w "%{http_code} " "http://localhost:8080/source-c/products?offset=0&limit=2"
// done; echo
