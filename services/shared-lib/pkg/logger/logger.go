package logger

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Log is the global structured logger. Must call Init before use.
var Log *zap.Logger

// Init initialises the global structured JSON logger for production use.
//
// Environment variables consumed:
//   - ENV             – deployment environment (dev/staging/prod)
//   - LOG_LEVEL       – minimum log level: debug, info, warn, error (default: info)
//   - SERVICE_VERSION – semantic version injected at build time (e.g. "2.1.0")
//   - BUILD_SHA       – git commit SHA injected at build time
func Init(serviceName string) {
	level := parseLevel(os.Getenv("LOG_LEVEL"))

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderCfg.EncodeLevel = zapcore.LowercaseLevelEncoder
	encoderCfg.CallerKey = "caller"
	encoderCfg.EncodeCaller = zapcore.ShortCallerEncoder

	jsonEncoder := zapcore.NewJSONEncoder(encoderCfg)

	// Split: INFO+ → stdout, ERROR+ → stderr (allows log aggregators to triage by stream)
	stdoutSyncer := zapcore.Lock(os.Stdout)
	stderrSyncer := zapcore.Lock(os.Stderr)

	core := zapcore.NewTee(
		// All logs at configured level → stdout
		zapcore.NewCore(jsonEncoder, stdoutSyncer, level),
		// Error and above → stderr as well
		zapcore.NewCore(jsonEncoder, stderrSyncer, zapcore.ErrorLevel),
	)

	Log = zap.New(core,
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
		zap.Fields(
			zap.String("service", serviceName),
			zap.String("env", getEnv("ENV", "production")),
			zap.String("version", getEnv("SERVICE_VERSION", "unknown")),
			zap.String("build_sha", getEnv("BUILD_SHA", "unknown")),
		),
	)
}

// parseLevel converts a string log level to a zapcore.Level, defaulting to Info.
func parseLevel(lvl string) zapcore.Level {
	switch strings.ToLower(lvl) {
	case "debug":
		return zapcore.DebugLevel
	case "warn", "warning":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
