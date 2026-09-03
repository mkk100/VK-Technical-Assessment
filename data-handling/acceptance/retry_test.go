package acceptance

import (
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	"data-handling/internal/pipeline"
)

// 3. Retry behavior.
//
// Retry logic is timing-dependent and impossible to verify by eyeballing.
// Wrong backoff, retrying a 4xx, ignoring Retry-After, an off-by-one on the
// attempt cap — all invisible until production. Drive it through a fetcher
// against an httptest.Server with a call counter.

func TestRetry_TransientThenSuccess(t *testing.T) {
	var calls atomic.Int32
	serve(t, func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) < 3 {
			write(w, http.StatusTooManyRequests, `{"error":"rate_limited"}`)
			return
		}
		write(w, http.StatusOK, `{"page":1,"total_pages":1,"products":[
			{"id":"a-1","name":"X","price":1,"category":"c"}]}`)
	})

	got, err := pipeline.FetchSourceA()
	if err != nil {
		t.Fatalf("expected success after retries, got %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 product, got %d", len(got))
	}
	if n := calls.Load(); n != 3 {
		t.Fatalf("expected 3 requests (2x429 + 1x200), got %d", n)
	}
}

func TestRetry_GivesUpAfterMaxRetries(t *testing.T) {
	var calls atomic.Int32
	serve(t, func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		write(w, http.StatusServiceUnavailable, `{"error":"upstream"}`)
	})

	if _, err := pipeline.FetchSourceA(); err == nil {
		t.Fatal("expected an error when upstream never recovers")
	}
	if n := calls.Load(); n != int32(pipeline.MaxRetries+1) {
		t.Fatalf("expected %d attempts, got %d", pipeline.MaxRetries+1, n)
	}
}

func TestRetry_NoRetryOn4xx(t *testing.T) {
	var calls atomic.Int32
	serve(t, func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		write(w, http.StatusBadRequest, `{"error":"bad_request"}`)
	})

	if _, err := pipeline.FetchSourceA(); err == nil {
		t.Fatal("expected an error for 400")
	}
	if n := calls.Load(); n != 1 {
		t.Fatalf("4xx must not be retried: got %d requests", n)
	}
}

func TestRetry_HonorsRetryAfterHeader(t *testing.T) {
	var calls atomic.Int32
	serve(t, func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			w.Header().Set("Retry-After", "1")
			write(w, http.StatusTooManyRequests, `{}`)
			return
		}
		write(w, http.StatusOK, `{"page":1,"total_pages":1,"products":[]}`)
	})

	start := time.Now()
	if _, err := pipeline.FetchSourceA(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if waited := time.Since(start); waited < 900*time.Millisecond {
		t.Fatalf("expected to wait ~1s for Retry-After, waited %s", waited)
	}
}
