/**
 * @Author: dingQingHui
 * @Description:
 * @File: logger
 * @Version: 1.0.0
 * @Date: 2024/11/25 16:42
 */

package api

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type IZLogger interface {
	IModule
	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Panic(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
	SetLogLevel(logLevel zapcore.Level)
	GetLogLevel() zapcore.Level
}
