package pipeline

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

const maxRetries = 4

var httpClient = &http.Client{Timeout: 5 * time.Second}

// getJSON fetches url and decodes the body into v. It retries 429 and 5xx
// (honoring Retry-After) with exponential backoff, and fails fast on 4xx.
func getJSON(url string, v any) error {
	backoff := 300 * time.Millisecond

	for attempt := 0; attempt <= maxRetries; attempt++ {
		resp, err := httpClient.Get(url)
		if err != nil {
			logger.Error("request failed", "url", url, "err", err)
			return err // transport error
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			return json.Unmarshal(body, v)
		}

		if resp.StatusCode == 429 || resp.StatusCode >= 500 {
			if attempt == maxRetries {
				logger.Error("upstream give up", "url", url, "status", resp.StatusCode, "attempts", attempt+1)
				return fmt.Errorf("GET %s: gave up, last status %d", url, resp.StatusCode)
			}
			wait := backoff
			if ra := resp.Header.Get("Retry-After"); ra != "" {
				if secs, e := strconv.Atoi(ra); e == nil {
					wait = time.Duration(secs) * time.Second
				}
			}
			logger.Warn("retrying upstream", "url", url, "status", resp.StatusCode, "attempt", attempt+1, "wait", wait.String())
			time.Sleep(wait)
			backoff *= 2
			continue
		}

		logger.Error("upstream permanent error", "url", url, "status", resp.StatusCode)
		return fmt.Errorf("GET %s: status %d: %s", url, resp.StatusCode, body)
	}
	return fmt.Errorf("GET %s: retries exhausted", url)
}
