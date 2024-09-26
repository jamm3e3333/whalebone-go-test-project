package logger

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

const (
	timestampFieldName = "@timestamp"
	timeFieldFormat    = time.RFC3339Nano
)

var once sync.Once

type Logger interface {
	Trace(message any, args ...any)
	Debug(message any, args ...any)
	Info(message any, args ...any)
	Warn(message any, args ...any)
	Error(message any, args ...any)
	Fatal(message any, args ...any)
	WithFields(meta []Meta) *ZeroLogger
	DebugWithMetadata(message string, metadata map[string]any)
	InfoWithMetadata(message string, metadata map[string]any)
	WarnWithMetadata(message string, metadata map[string]any)
	ErrorWithMetadata(message string, metadata map[string]any)
	FatalWithMetadata(message string, metadata map[string]any)
}

type ZeroLogger struct {
	logger          *zerolog.Logger
	writers         []io.Writer
	developmentMode bool
}

var (
	_ Logger = (*ZeroLogger)(nil)
)

func New(lvl Level, develFlag bool, additionalWriters ...io.Writer) *ZeroLogger {
	var multiWriter zerolog.LevelWriter
	var writers []io.Writer

	writers = additionalWriters

	var consoleWriter zerolog.ConsoleWriter
	if develFlag {
		consoleWriter = zerolog.ConsoleWriter{Out: os.Stdout, NoColor: false, TimeFormat: "15:04:05"}
		writers = append(writers, consoleWriter)
	} else {
		writers = append(writers, os.Stdout)
	}

	zerolog.SetGlobalLevel(zerolog.Level(lvl))

	multiWriter = zerolog.MultiLevelWriter(writers...)
	innerLog := zerolog.New(multiWriter).With().Timestamp().Logger().Level(zerolog.TraceLevel)

	zerolog.TimestampFieldName = timestampFieldName
	zerolog.TimeFieldFormat = timeFieldFormat

	return &ZeroLogger{
		logger:          &innerLog,
		writers:         writers,
		developmentMode: develFlag,
	}
}

func (l *ZeroLogger) Trace(message any, args ...any) {
	l.msg(TraceLevel, message, args...)
}

func (l *ZeroLogger) Debug(message any, args ...any) {
	l.msg(DebugLevel, message, args...)
}

func (l *ZeroLogger) Info(message any, args ...any) {
	l.msg(InfoLevel, message, args...)
}

func (l *ZeroLogger) Warn(message any, args ...any) {
	l.msg(WarnLevel, message, args...)
}

func (l *ZeroLogger) Error(message any, args ...any) {
	l.msg(ErrorLevel, message, args...)
}

func (l *ZeroLogger) Fatal(message any, args ...any) {
	l.msg(FatalLevel, message, args...)
	os.Exit(1)
}

func (l *ZeroLogger) DebugWithMetadata(message string, metadata map[string]any) {
	l.logWithMetadata(zerolog.DebugLevel, message, metadata)
}

func (l *ZeroLogger) InfoWithMetadata(message string, metadata map[string]any) {
	l.logWithMetadata(zerolog.InfoLevel, message, metadata)
}

func (l *ZeroLogger) WarnWithMetadata(message string, metadata map[string]any) {
	l.logWithMetadata(zerolog.WarnLevel, message, metadata)
}

func (l *ZeroLogger) ErrorWithMetadata(message string, metadata map[string]any) {
	l.logWithMetadata(zerolog.ErrorLevel, message, metadata)
}

func (l *ZeroLogger) FatalWithMetadata(message string, metadata map[string]any) {
	l.logWithMetadata(zerolog.FatalLevel, message, metadata)
}

func (l *ZeroLogger) IsDevelopmentMode() bool {
	return l.developmentMode
}

func (l *ZeroLogger) getEventAtLevel(level Level) *zerolog.Event {
	var e *zerolog.Event

	switch level {
	case FatalLevel:
		e = l.logger.Fatal()
	case ErrorLevel:
		e = l.logger.Error()
	case WarnLevel:
		e = l.logger.Warn()
	case DebugLevel:
		e = l.logger.Debug()
	case TraceLevel:
		e = l.logger.Trace()
	default: // default covers the info level
		e = l.logger.Info()
	}

	return e
}

func (l *ZeroLogger) log(level Level, message string, args ...any) {
	if len(args) == 0 {
		l.getEventAtLevel(level).Msg(message)
		return
	}

	l.getEventAtLevel(level).Msgf(message, args...)
}

func (l *ZeroLogger) msg(level Level, message any, args ...any) {
	switch msgType := message.(type) {
	case error:
		l.log(level, msgType.Error(), args...)
	case string:
		l.log(level, msgType, args...)
	default:
		if len(args) == 0 {
			l.log(level, fmt.Sprintf("%v", message))
			return
		}
		l.log(level, fmt.Sprintf("%v %v", message, args))
	}
}
