package logx

import (
	"log/slog"
	"os"
)

type Logger = slog.Logger

func New() *Logger {
	h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})
	return slog.New(h)
}
