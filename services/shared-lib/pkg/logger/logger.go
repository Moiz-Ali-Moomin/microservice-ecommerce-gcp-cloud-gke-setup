package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger

func Init(serviceName string) {
	config := zap.NewProductionEncoderConfig()
	config.TimeKey = "timestamp"
	config.EncodeTime = zapcore.ISO8601TimeEncoder

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(config),
		zapcore.Lock(os.Stdout),
		zapcore.InfoLevel,
	)

	Log = zap.New(core, zap.Fields(
		zap.String("service", serviceName),
		zap.String("env", os.Getenv("ENV")),
	))
}
