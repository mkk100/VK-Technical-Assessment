package pipeline

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// MaxRetries is the number of retries per request (so MaxRetries+1 attempts).
const MaxRetries = 4

var (
	httpClient = &http.Client{Timeout: 5 * time.Second}
	// BaseBackoff is the initial retry delay, doubled on each retry. Exposed so
	// tests can shrink it; production keeps the default.
	BaseBackoff = 300 * time.Millisecond
)

// getJSON fetches url and decodes the body into v. It retries 429 and 5xx
// (honoring Retry-After) with exponential backoff, and fails fast on 4xx.
func getJSON(url string, v any) error {
	backoff := BaseBackoff

	for attempt := 0; attempt <= MaxRetries; attempt++ {
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
			if attempt == MaxRetries {
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
