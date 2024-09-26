package logger

import (
	"github.com/rs/zerolog"
)

const (
	uuidField    = "uuid"
	packageField = "package"
	funcField    = "func"
)

func (l *ZeroLogger) WithUUID(uuid string) *ZeroLogger {
	return l.WithField(uuidField, uuid)
}

func (l *ZeroLogger) WithFuncName(funcName string) *ZeroLogger {
	return l.WithField(funcField, funcName)
}

func (l *ZeroLogger) WithPackage(text string) *ZeroLogger {
	return l.WithField(packageField, text)
}

type Meta struct {
	key, val string
}

func NewMeta(key, val string) Meta {
	return Meta{key: key, val: val}
}

func (l *ZeroLogger) WithField(key, val string) *ZeroLogger {
	out := l.logger.With().Str(key, val).Logger().Level(l.logger.GetLevel())

	return &ZeroLogger{
		logger:          &out,
		writers:         l.writers,
		developmentMode: l.developmentMode,
	}
}

func (l *ZeroLogger) WithFields(meta []Meta) *ZeroLogger {
	if len(meta) == 0 {
		return l
	}

	lCtx := l.logger.With()
	for _, m := range meta {
		lCtx = lCtx.Str(m.key, m.val)
	}

	out := lCtx.Logger()

	return &ZeroLogger{
		logger:          &out,
		writers:         l.writers,
		developmentMode: l.developmentMode,
	}
}

func (l *ZeroLogger) logWithMetadata(severity zerolog.Level, message string, metadata map[string]any) {
	event := l.logger.WithLevel(severity)
	flatten("", metadata, func(key string, value any) {
		event = event.Interface(key, value)
	})
	event.Msg(message)
}

func flatten(prefix string, metadata map[string]any, visit func(key string, value any)) {
	for k, v := range metadata {
		if nested, ok := v.(map[string]any); ok {
			flatten(prefix+k+".", nested, visit)
		} else {
			visit(prefix+k, v)
		}
	}
}
