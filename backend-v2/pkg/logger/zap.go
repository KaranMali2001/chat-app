package logger

import (
	"go.uber.org/zap"
	"sync"
)

var (
	once   sync.Once
	Logger *zap.Logger
	err    error
)

func Init(isProduction bool) error {
	once.Do(func() {
		if isProduction {
			Logger, err = zap.NewProduction()
		} else {
			Logger, err = zap.NewDevelopment()
		}
		if err == nil {
			zap.ReplaceGlobals(Logger)
		}
	})
	return err
}
