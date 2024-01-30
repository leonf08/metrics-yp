package logger

import (
	"os"

	"github.com/rs/zerolog"
)

// NewLogger creates a new logger.
func NewLogger() zerolog.Logger {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	log := zerolog.New(os.Stderr).With().Timestamp().Logger()

	return log
}
