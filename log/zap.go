package log

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	Level zapcore.Level
}

func NewZapLogger(config Config) *zap.SugaredLogger {
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderCfg.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderCfg.TimeKey = "time"

	loggerCfg := zap.Config{
		Level:             zap.NewAtomicLevelAt(config.Level),
		DisableCaller:     true,
		DisableStacktrace: true,
		Encoding:          "json",
		EncoderConfig:     encoderCfg,
		OutputPaths:       []string{"stdout"},
	}

	zapLog, err := loggerCfg.Build()
	if err != nil {
		panic(fmt.Sprintf("failed to init logger: %s", err))
	}

	return zapLog.Sugar()
}
