package pipeline

import (
	"io"
	"log/slog"
	"os"
)

// logger emits structured events to stderr, keeping stdout clean for the
// product JSON. Every operational decision (retry, drop, source start/finish,
// give-up) is logged here as key=value pairs.
var logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
	Level: slog.LevelInfo,
}))

// SetLogOutput redirects structured logs to w (e.g. a file, or io.Discard in
// tests). Not safe to call concurrently with a run.
func SetLogOutput(w io.Writer) {
	logger = slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{Level: slog.LevelInfo}))
}
