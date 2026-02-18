package observability

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var GlobalLogger *zap.Logger

// InitLogger initializes the global logger based on environment
func InitLogger(environment string) error {
	var config zap.Config

	if environment == "production" {
		config = zap.NewProductionConfig()
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		config.EncoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}

	logger, err := config.Build(
		zap.AddCaller(),
		zap.AddCallerSkip(0),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)

	if err != nil {
		return err
	}

	GlobalLogger = logger
	zap.ReplaceGlobals(logger)

	return nil
}

// Sync flushes any buffered log entries
func Sync() {
	if GlobalLogger != nil {
		_ = GlobalLogger.Sync()
	}
}

// GetLogger returns the global logger or creates a default one
func GetLogger() *zap.Logger {
	if GlobalLogger == nil {
		logger, _ := zap.NewProduction()
		GlobalLogger = logger
	}
	return GlobalLogger
}

// SetOutput sets the output for the logger
func SetOutput(paths []string) {
	if GlobalLogger != nil {
		config := zap.NewProductionConfig()
		config.OutputPaths = paths
		logger, _ := config.Build()
		GlobalLogger = logger
		zap.ReplaceGlobals(logger)
	}
}

// IsProduction checks if running in production mode
func IsProduction() bool {
	env := os.Getenv("ENVIRONMENT")
	return env == "production"
}
