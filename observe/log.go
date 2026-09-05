package observe

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/rs/zerolog"
)

type Logger = zerolog.Logger

type Level = zerolog.Level

const (
	LevelTrace    Level = -1
	LevelDebug    Level = 0
	LevelInfo     Level = 1
	LevelWarn     Level = 2
	LevelError    Level = 3
	LevelFatal    Level = 4
	LevelPanic    Level = 5
	LevelNone     Level = 6
	LevelDisabled Level = 7
)

func ParseLevel(level string) Level {
	switch strings.ToLower(level) {
	case "trace":
		return LevelTrace
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "warn", "warning":
		return LevelWarn
	case "error":
		return LevelError
	case "fatal":
		return LevelFatal
	case "panic":
		return LevelPanic
	case "none":
		return LevelNone
	case "disabled":
		return LevelDisabled
	default:
		return LevelInfo
	}
}

func NewLogger(level Level, pretty bool) *Logger {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	var logger Logger
	if pretty {
		logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout})
	} else {
		logger = zerolog.New(os.Stdout)
	}

	logger = logger.With().Timestamp().Logger()

	if !pretty {
		logger = logger.With().Caller().Logger()
		logger = logger.With().Stack().Logger()
	}

	logger = logger.Level(level)
	return &logger
}

func TestLogger(t *testing.T) *Logger {
	logger := zerolog.New(t.Output())
	logger = logger.Level(LevelTrace)
	return &logger
}

var (
	LogMetadataErrKey   = "error"
	LogMetadataPanicKey = "panic"
)

type LoggingMetadataKey struct{}
type LoggingMetadata map[string]any

func LoggingContext(ctx context.Context) context.Context {
	metadata := LoggingMetadata{}
	return context.WithValue(ctx, LoggingMetadataKey{}, metadata)
}

func GetLoggingMetadata(ctx context.Context) LoggingMetadata {
	return ctx.Value(LoggingMetadataKey{}).(LoggingMetadata)
}

func SetLoggingMetadata(ctx context.Context, key string, value any) {
	GetLoggingMetadata(ctx)[key] = value
}
