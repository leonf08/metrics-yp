package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

// NewLogger creates a new logger.
func NewLogger() zerolog.Logger {
	zerolog.TimeFieldFormat = time.DateTime
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	log := zerolog.New(os.Stderr).With().Timestamp().Logger()

	return log
}
