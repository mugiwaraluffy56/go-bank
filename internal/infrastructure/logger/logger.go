package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

type Logger struct {
	*zerolog.Logger
}

func New(environment string) *Logger {
	var output io.Writer = os.Stdout

	if environment == "development" {
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
	}

	zerolog.TimeFieldFormat = time.RFC3339
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	if environment == "development" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	logger := zerolog.New(output).
		With().
		Timestamp().
		Caller().
		Str("service", "gobank").
		Logger()

	return &Logger{&logger}
}

func (l *Logger) WithRequestID(requestID string) *Logger {
	logger := l.Logger.With().Str("request_id", requestID).Logger()
	return &Logger{&logger}
}

func (l *Logger) WithUserID(userID string) *Logger {
	logger := l.Logger.With().Str("user_id", userID).Logger()
	return &Logger{&logger}
}

func (l *Logger) WithField(key string, value interface{}) *Logger {
	logger := l.Logger.With().Interface(key, value).Logger()
	return &Logger{&logger}
}

func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	ctx := l.Logger.With()
	for k, v := range fields {
		ctx = ctx.Interface(k, v)
	}
	logger := ctx.Logger()
	return &Logger{&logger}
}

func (l *Logger) WithError(err error) *Logger {
	logger := l.Logger.With().Err(err).Logger()
	return &Logger{&logger}
}
