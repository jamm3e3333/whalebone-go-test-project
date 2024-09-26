package logger

import (
	"strings"

	"github.com/rs/zerolog"
)

// Level defines log levels.
type Level int8

const (
	// DebugLevel defines debug log level.
	DebugLevel Level = iota
	// InfoLevel defines info log level.
	InfoLevel
	// WarnLevel defines warn log level.
	WarnLevel
	// ErrorLevel defines error log level.
	ErrorLevel
	// FatalLevel defines fatal log level.
	FatalLevel

	// TraceLevel defines trace log level.
	TraceLevel Level = -1
	// Values less than TraceLevel are handled as numbers.
)

const (
	traceLevelValue = "trace"
	debugLevelValue = "debug"
	infoLevelValue  = "info"
	warnLevelValue  = "warn"
	errorLevelValue = "error"
	fatalLevelValue = "fatal"
)

func (l Level) String() string {
	switch l {
	case TraceLevel:
		return traceLevelValue
	case DebugLevel:
		return debugLevelValue
	case WarnLevel:
		return warnLevelValue
	case ErrorLevel:
		return errorLevelValue
	case FatalLevel:
		return fatalLevelValue
	default:
		return infoLevelValue
	}
}

// ParseLevel will parse a string to the Level with InfoLevel as default and safe fallback value
func ParseLevel(level string) Level {
	var l Level

	switch strings.ToLower(level) {
	case fatalLevelValue:
		l = FatalLevel
	case errorLevelValue:
		l = ErrorLevel
	case warnLevelValue:
		l = WarnLevel
	case infoLevelValue:
		l = InfoLevel
	case debugLevelValue:
		l = DebugLevel
	case traceLevelValue:
		l = TraceLevel
	default:
		l = InfoLevel
	}
	return l
}

func (l *ZeroLogger) GetLevel() Level {
	return Level(zerolog.GlobalLevel())
}

func (l *ZeroLogger) SetLevel(lvl Level) {
	zerolog.SetGlobalLevel(zerolog.Level(lvl))
}
