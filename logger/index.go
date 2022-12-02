package logger

import (
	"go.uber.org/zap/zapcore"
)

func Info(msg string, fields ...zapcore.Field) {
	Logger.Info(msg, fields...)
}

func Error(err error, fields ...zapcore.Field) {
	Logger.Error(err.Error(), fields...)
}
