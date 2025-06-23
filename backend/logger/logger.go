package logger

import (
	"sync"

	"go.uber.org/zap"
)

var (
	logger *zap.SugaredLogger
	once   sync.Once
)

func GetLogger() *zap.SugaredLogger {
	once.Do(func() {
		rawLogger, _ := zap.NewDevelopment()
		logger = rawLogger.Sugar()
	})
	return logger
}
