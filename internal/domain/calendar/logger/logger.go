package logger

import (
	"context"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger represents the logging interface
type Logger interface {
	Debug(ctx context.Context, msg string, fields ...zapcore.Field)
	Info(ctx context.Context, msg string, fields ...zapcore.Field)
	Warn(ctx context.Context, msg string, fields ...zapcore.Field)
	Error(ctx context.Context, msg string, fields ...zapcore.Field)
	Fatal(ctx context.Context, msg string, fields ...zapcore.Field)
	With(fields ...zapcore.Field) Logger
}

type loggerImpl struct {
	logger *zap.Logger
}

// Config holds logger configuration
type Config struct {
	Environment string // "development" or "production"
	Level       string // "debug", "info", "warn", "error"
	OutputPaths []string
}

// NewLogger creates a new configured logger
func NewLogger(config Config) (Logger, error) {
	// Base logger configuration
	zapConfig := zap.NewProductionConfig()
	if config.Environment == "development" {
		zapConfig = zap.NewDevelopmentConfig()
	}

	// Configure log level
	level, err := zapcore.ParseLevel(config.Level)
	if err != nil {
		level = zapcore.InfoLevel
	}
	zapConfig.Level = zap.NewAtomicLevelAt(level)

	// Configure output paths
	if len(config.OutputPaths) > 0 {
		zapConfig.OutputPaths = config.OutputPaths
	}

	// Create custom encoder config for JSON formatting
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	zapConfig.EncoderConfig = encoderConfig

	logger, err := zapConfig.Build(
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	if err != nil {
		return nil, err
	}

	return &loggerImpl{
		logger: logger,
	}, nil
}

// extractTraceInfo extracts OpenTelemetry trace information from context
func extractTraceInfo(ctx context.Context) []zapcore.Field {
	if ctx == nil {
		return nil
	}

	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return nil
	}

	spanCtx := span.SpanContext()
	if !spanCtx.IsValid() {
		return nil
	}

	return []zapcore.Field{
		zap.String("trace_id", spanCtx.TraceID().String()),
		zap.String("span_id", spanCtx.SpanID().String()),
	}
}

func (l *loggerImpl) Debug(ctx context.Context, msg string, fields ...zapcore.Field) {
	if traceFields := extractTraceInfo(ctx); len(traceFields) > 0 {
		fields = append(fields, traceFields...)
	}
	l.logger.Debug(msg, fields...)
}

func (l *loggerImpl) Info(ctx context.Context, msg string, fields ...zapcore.Field) {
	if traceFields := extractTraceInfo(ctx); len(traceFields) > 0 {
		fields = append(fields, traceFields...)
	}
	l.logger.Info(msg, fields...)
}

func (l *loggerImpl) Warn(ctx context.Context, msg string, fields ...zapcore.Field) {
	if traceFields := extractTraceInfo(ctx); len(traceFields) > 0 {
		fields = append(fields, traceFields...)
	}
	l.logger.Warn(msg, fields...)
}

func (l *loggerImpl) Error(ctx context.Context, msg string, fields ...zapcore.Field) {
	if traceFields := extractTraceInfo(ctx); len(traceFields) > 0 {
		fields = append(fields, traceFields...)
	}
	l.logger.Error(msg, fields...)
}

func (l *loggerImpl) Fatal(ctx context.Context, msg string, fields ...zapcore.Field) {
	if traceFields := extractTraceInfo(ctx); len(traceFields) > 0 {
		fields = append(fields, traceFields...)
	}
	l.logger.Fatal(msg, fields...)
}

func (l *loggerImpl) With(fields ...zapcore.Field) Logger {
	return &loggerImpl{
		logger: l.logger.With(fields...),
	}
}

// Example usage:
/*
config := logger.Config{
    Environment: "production",
    Level:       "info",
    OutputPaths: []string{
        "stdout",
        "/var/log/calendar-service.log",
        "elasticsearch://localhost:9200/logs-calendar/doc",
    },
}

log, err := logger.NewLogger(config)
if err != nil {
    panic(err)
}

// Usage with context and tracing
log.Info(ctx, "Processing email",
    zap.String("user_id", userID),
    zap.String("email_id", emailID),
    zap.Int("retry_count", retryCount),
)

// Create scoped logger
userLogger := log.With(
    zap.String("user_id", userID),
    zap.String("service", "calendar"),
)

userLogger.Error(ctx, "Failed to create calendar event",
    zap.Error(err),
    zap.String("event_id", eventID),
)
*/
