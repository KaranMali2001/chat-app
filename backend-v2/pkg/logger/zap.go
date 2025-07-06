package logger

import (
	"go.uber.org/zap"
	"sync"
)

var (
	once          sync.Once
	Logger        *zap.Logger
	SugaredLogger *zap.SugaredLogger
	err           error
)

func Init(isProduction bool) error {
	once.Do(func() {
		if isProduction {
			Logger, err = zap.NewProduction()
		} else {
			Logger, err = zap.NewDevelopment()
		}
		if err == nil {
			SugaredLogger = Logger.Sugar().WithOptions(zap.AddCallerSkip(1))
			Logger.WithOptions(zap.AddCallerSkip(1))
			zap.ReplaceGlobals(Logger)
		}
	})
	return err
}
func Info(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Info(msg, fields...)
	}
}

func Error(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Error(msg, fields...)
	}
}

func Warn(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Warn(msg, fields...)
	}
}

func Debug(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Debug(msg, fields...)
	}
}

func Fatal(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Fatal(msg, fields...)
	}
}
func Infof(template string, args ...interface{}) {
	if SugaredLogger != nil {
		SugaredLogger.Infof(template, args...)
	}
}
func Errorln(args ...interface{}) {
	if SugaredLogger != nil {
		SugaredLogger.Errorln(args...)
	}
}
