package logger

import (
	"fmt"

	"go.uber.org/zap/zapcore"
)

func Info(msg string, fields ...zapcore.Field) {
	Logger.Info(msg, fields...)
}

func Error(err error, fields ...zapcore.Field) {
	fmt.Println(err)
	fmt.Println(fields)
	Logger.Error(err.Error(), fields...)
}
