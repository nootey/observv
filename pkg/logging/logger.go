package logging

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var newEncoder = map[bool]func() zapcore.Encoder{
	true: func() zapcore.Encoder {
		return zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	},
	false: func() zapcore.Encoder {
		return zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	},
}

func InitLogger(release bool, logLevel string) *zap.Logger {
	level := zap.NewAtomicLevelAt(parseLogLevel(logLevel))

	core := zapcore.NewCore(newEncoder[release](), zapcore.AddSync(os.Stdout), level)
	return zap.New(core, zap.AddCaller())
}

func parseLogLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel // default to info if invalid
	}
}
