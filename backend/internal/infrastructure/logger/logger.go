package logger

import (
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

func New(level string) zerolog.Logger {
	parsed, err := zerolog.ParseLevel(strings.ToLower(level))
	if err != nil {
		parsed = zerolog.DebugLevel
	}
	zerolog.SetGlobalLevel(parsed)

	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	if strings.EqualFold(os.Getenv("APP_ENV"), "production") {
		return zerolog.New(os.Stdout).With().Timestamp().Logger()
	}
	return zerolog.New(output).With().Timestamp().Logger()
}
