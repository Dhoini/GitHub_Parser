package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ZapToCustomAdapter адаптирует zap.Logger к вашему пользовательскому логгеру
type ZapToCustomAdapter struct {
	customLogger *Logger
}

// NewZapToCustomAdapter создает новый адаптер
func NewZapToCustomAdapter(customLogger *Logger) *zap.Logger {
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
		zapcore.AddSync(zapcore.WriteSyncer(zapcore.AddSync(&ZapToCustomAdapter{customLogger: customLogger}))),
		zapcore.DebugLevel,
	)
	return zap.New(core)
}

// CustomLogger возвращает пользовательский логгер для использования в проекте
func CustomLogger() *Logger {
	return New(DEBUG)
}

// Write реализует io.Writer для интеграции с zap
func (l *ZapToCustomAdapter) Write(p []byte) (n int, err error) {
	l.customLogger.Info(string(p))
	return len(p), nil
}
