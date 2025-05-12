package logging

import (
	"go.uber.org/zap"
	"os"
)

func NewLogger() (*zap.Logger, error) {
	logPath := os.Getenv("LOG_FILE_PATH")
	if logPath == "" {
		logPath = "logs/app.log"
	}

	cfg := zap.NewProductionConfig()
	cfg.OutputPaths = []string{logPath}
	cfg.ErrorOutputPaths = []string{logPath}

	return cfg.Build()
}
