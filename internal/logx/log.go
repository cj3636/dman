package logx

import (
	"log/slog"
	"os"
	"strings"
)

type Logger = slog.Logger

func New() *Logger { return NewWithLevel("info") }

func NewWithLevel(level string) *Logger {
	lvl := slog.LevelInfo
	switch strings.ToLower(level) {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	case "info":
	default:
	}
	h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: lvl})
	return slog.New(h)
}
