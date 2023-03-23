package logging

import (
	"log"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LogPath struct {
	Value string
}

var (
	instance *zap.Logger
	once     sync.Once
)

func YLSLogger(opts ...interface{}) *zap.Logger {
	once.Do(func() {
		zapConfig := zap.NewProductionConfig()
		for _, opt := range opts {
			if val, ok := opt.(zapcore.Level); ok {
				zapConfig.Level.SetLevel(val)
			}
			if val, ok := opt.(LogPath); ok {
				zapConfig.OutputPaths = append(zapConfig.OutputPaths, val.Value)
			}
		}

		logger, err := zapConfig.Build()
		if err != nil {
			log.Fatalf("cannot initialize zap logger: %v", err)
		}

		instance = logger
	})

	return instance
}
