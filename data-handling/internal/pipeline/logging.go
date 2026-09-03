package pipeline

import (
	"log/slog"
	"os"
)

// logger emits structured events to stderr, keeping stdout clean for the
// product JSON. Every operational decision (retry, drop, source start/finish,
// give-up) is logged here as key=value pairs.
var logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
	Level: slog.LevelInfo,
}))
