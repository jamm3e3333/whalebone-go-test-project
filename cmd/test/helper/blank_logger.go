package helper

import (
	"context"
	"fmt"

	"github.com/jamm3e3333/whalebone-go-test-project/pkg/logger"
)

type BlankLogger struct {
}

func NewBlankLogger() *BlankLogger {
	return &BlankLogger{}
}

func (b *BlankLogger) Trace(_ any, _ ...any) {
}
func (b *BlankLogger) Debug(_ any, _ ...any) {
}
func (b *BlankLogger) Info(_ any, _ ...any) {
}
func (b *BlankLogger) Warn(_ any, _ ...any) {
}
func (b *BlankLogger) Error(message any, args ...any) {
	fmt.Printf("[ERROR_TEST] %s, %v\n", message, args)
}
func (b *BlankLogger) Fatal(_ any, _ ...any) {
}
func (b *BlankLogger) WithFields(_ []logger.Meta) *logger.ZeroLogger {
	return nil
}
func (b *BlankLogger) WithAPM(_ context.Context) logger.Logger {
	return NewBlankLogger()
}

func (b *BlankLogger) TraceWithMetadata(_ string, _ map[string]any) {}
func (b *BlankLogger) DebugWithMetadata(_ string, _ map[string]any) {}
func (b *BlankLogger) InfoWithMetadata(_ string, _ map[string]any)  {}
func (b *BlankLogger) WarnWithMetadata(_ string, _ map[string]any)  {}
func (b *BlankLogger) ErrorWithMetadata(_ string, _ map[string]any) {}
func (b *BlankLogger) FatalWithMetadata(_ string, _ map[string]any) {}
