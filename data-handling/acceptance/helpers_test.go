// Package acceptance holds black-box behavior tests for the pipeline. It only
// touches the exported API of data-handling/internal/pipeline and drives the
// fetchers against httptest servers, so no mock service or network is needed.
package acceptance

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"data-handling/internal/pipeline"
)

// TestMain silences pipeline logs and shrinks the retry backoff for the suite.
func TestMain(m *testing.M) {
	pipeline.SetLogOutput(io.Discard)
	pipeline.BaseBackoff = time.Millisecond
	os.Exit(m.Run())
}

// serve starts a test server, points pipeline.BaseURL at it, and restores the
// previous value on cleanup. Also resets the malformed-record counters.
func serve(t *testing.T, h http.HandlerFunc) {
	t.Helper()
	srv := httptest.NewServer(h)
	prev := pipeline.BaseURL
	pipeline.BaseURL = srv.URL
	pipeline.ResetDropped()
	t.Cleanup(func() {
		pipeline.BaseURL = prev
		srv.Close()
	})
}

func write(w http.ResponseWriter, status int, body string) {
	w.WriteHeader(status)
	io.WriteString(w, body)
}
