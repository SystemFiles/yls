package logging

import (
	"log"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	instance *zap.Logger
	once     sync.Once
)

func YLSLogger(level ...zapcore.Level) *zap.Logger {
	once.Do(func() {
		zapConfig := zap.NewProductionConfig()
		if len(level) > 0 {
			zapConfig.Level.SetLevel(level[0])
		}

		logger, err := zapConfig.Build()
		if err != nil {
			log.Fatalf("cannot initialize zap logger: %v", err)
		}

		instance = logger
	})

	return instance
}
