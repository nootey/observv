package logging

import (
	"context"
	"os"
	"slices"

	"go.opentelemetry.io/contrib/bridges/otelzap"
	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const scopeName = "github.com/nootey/observv"

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
	return zap.New(traceCore{core}, zap.AddCaller())
}

// WithOTel also sends each entry to the OTel log bridge, at the stdout level.
func WithOTel(logger *zap.Logger, provider log.LoggerProvider) *zap.Logger {
	return logger.WithOptions(zap.WrapCore(func(stdout zapcore.Core) zapcore.Core {
		bridge := otelzap.NewCore(scopeName, otelzap.WithLoggerProvider(provider))
		return zapcore.NewTee(stdout, leveledCore{Core: bridge, level: stdout})
	}))
}

// Context links a log entry to the span in ctx.
func Context(ctx context.Context) zap.Field {
	return zap.Any("context", ctx)
}

// traceCore prints the trace and span IDs in place of a context field.
type traceCore struct {
	zapcore.Core
}

func (c traceCore) With(fields []zapcore.Field) zapcore.Core {
	return traceCore{c.Core.With(withTraceIDs(fields))}
}

func (c traceCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(ent.Level) {
		return ce.AddCore(ent, c)
	}
	return ce
}

func (c traceCore) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	return c.Core.Write(ent, withTraceIDs(fields))
}

func withTraceIDs(fields []zapcore.Field) []zapcore.Field {
	if !slices.ContainsFunc(fields, isContext) {
		return fields
	}

	out := make([]zapcore.Field, 0, len(fields)+1)
	for _, f := range fields {
		ctx, ok := f.Interface.(context.Context)
		if !ok {
			out = append(out, f)
			continue
		}
		if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
			out = append(out,
				zap.String("trace_id", sc.TraceID().String()),
				zap.String("span_id", sc.SpanID().String()),
			)
		}
	}
	return out
}

func isContext(f zapcore.Field) bool {
	_, ok := f.Interface.(context.Context)
	return ok
}

type leveledCore struct {
	zapcore.Core
	level zapcore.LevelEnabler
}

func (c leveledCore) Enabled(lvl zapcore.Level) bool {
	return c.level.Enabled(lvl) && c.Core.Enabled(lvl)
}

func (c leveledCore) With(fields []zapcore.Field) zapcore.Core {
	return leveledCore{Core: c.Core.With(fields), level: c.level}
}

func (c leveledCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if !c.level.Enabled(ent.Level) {
		return ce
	}
	return c.Core.Check(ent, ce)
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
