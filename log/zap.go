package log

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	Level string
}

func NewZapLogger(config Config) *zap.SugaredLogger {
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderCfg.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderCfg.TimeKey = "time"

	level, _ := zapcore.ParseLevel(config.Level)

	loggerCfg := zap.Config{
		Level:             zap.NewAtomicLevelAt(level),
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
