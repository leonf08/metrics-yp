package logger

import (
	"log/slog"
	"os"
	"path/filepath"
)

func NewLogger() *slog.Logger {
	replacer := func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.SourceKey {
			source := a.Value.Any().(*slog.Source)
			source.File = filepath.Base(source.File)
		}
		return a
	}

	l := slog.New(slog.NewJSONHandler(
		os.Stdout,
		&slog.HandlerOptions{AddSource: true, Level: slog.LevelInfo, ReplaceAttr: replacer}),
	)

	return l
}
