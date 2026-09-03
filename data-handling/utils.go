package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

type Product struct {
	ID       string  `json:"id"`
	Name   string  `json:"name"`
	Price    float64 `json:"price"`
	Category string  `json:"category"`
}

const baseURL = "http://localhost:8080"
var httpClient = &http.Client{Timeout: 5 * time.Second}
const maxRetries = 4

func getJSON(url string, v any) error { // asking 
	backoff := 300 * time.Millisecond

	for attempt := 0; attempt <= maxRetries; attempt++ {
		resp, err := httpClient.Get(url)
		if err != nil {
			return err // transport error: just fail (or retry if you want)
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		// success
		if resp.StatusCode == http.StatusOK {
			return json.Unmarshal(body, v)
		}

		// retryable: 429 or any 5xx
		if resp.StatusCode == 429 || resp.StatusCode >= 500 {
			if attempt == maxRetries {
				return fmt.Errorf("GET %s: gave up, last status %d", url, resp.StatusCode)
			}
			wait := backoff
			if ra := resp.Header.Get("Retry-After"); ra != "" {
				if secs, e := strconv.Atoi(ra); e == nil {
					wait = time.Duration(secs) * time.Second
				}
			}
			time.Sleep(wait)
			backoff *= 2
			continue
		}

		// anything else (4xx): permanent, don't retry
		return fmt.Errorf("GET %s: status %d: %s", url, resp.StatusCode, body)
	}
	return fmt.Errorf("GET %s: retries exhausted", url)
}