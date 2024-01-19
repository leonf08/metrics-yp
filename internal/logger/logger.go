package logger

import (
	"github.com/rs/zerolog"
	"os"
)

func NewLogger() zerolog.Logger {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	log := zerolog.New(os.Stderr).With().Timestamp().Logger()

	return log
}
