package logging

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path/filepath"
	"time"
)

func NewLogger(release bool) *zap.Logger {
	var cfg zap.Config

	if release {
		cfg = zap.NewProductionConfig()
	} else {
		cfg = zap.NewDevelopmentConfig()

		// Prettify output in dev mode
		//cfg.Encoding = "console"
		//cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// Build log file path
	logFile := getLogFilePath()

	cfg.OutputPaths = []string{
		"stdout", // keep console output
		logFile,
	}
	cfg.ErrorOutputPaths = []string{
		"stderr",
		logFile,
	}

	// Set log level
	cfg.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)

	logger, err := cfg.Build()
	if err != nil {
		panic(fmt.Sprintf("failed to build logger: %v", err))
	}

	return logger
}

func getLogFilePath() string {
	const logDir = "logs"

	// Ensure the directory exists
	if err := os.MkdirAll(logDir, 0755); err != nil {
		panic(fmt.Sprintf("failed to create log directory: %v", err))
	}

	filename := time.Now().Format("2006-01-02") + ".log"
	return filepath.Join(logDir, filename)
}
