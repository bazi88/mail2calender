package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger provides structured logging with tracing integration
type Logger struct {
	zap    *zap.Logger
	tracer trace.Tracer
	mu     sync.RWMutex
}

// Fields represents logging fields
type Fields map[string]interface{}

// LogLevel represents logging level
type LogLevel string

const (
	DebugLevel LogLevel = "debug"
	InfoLevel  LogLevel = "info"
	WarnLevel  LogLevel = "warn"
	ErrorLevel LogLevel = "error"
)

// New creates a new logger instance
func New(tracer trace.Tracer) (*Logger, error) {
	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// Disable stderr syncing to avoid "invalid argument" errors in tests
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stdout"}

	zapLogger, err := config.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %v", err)
	}

	return &Logger{
		zap:    zapLogger,
		tracer: tracer,
	}, nil
}

// WithContext adds trace context to log entries
func (l *Logger) WithContext(ctx context.Context) *Logger {
	l.mu.RLock()
	defer l.mu.RUnlock()

	spanCtx := trace.SpanContextFromContext(ctx)
	if !spanCtx.IsValid() {
		return l
	}

	logger := l.zap.With(
		zap.String("trace_id", spanCtx.TraceID().String()),
		zap.String("span_id", spanCtx.SpanID().String()),
	)

	return &Logger{
		zap:    logger,
		tracer: l.tracer,
	}
}

// WithFields adds fields to log entries
func (l *Logger) WithFields(fields Fields) *Logger {
	l.mu.RLock()
	defer l.mu.RUnlock()

	zapFields := make([]zap.Field, 0, len(fields))
	var fieldsMu sync.Mutex

	var wg sync.WaitGroup
	for k, v := range fields {
		wg.Add(1)
		go func(key string, value interface{}) {
			defer wg.Done()
			fieldsMu.Lock()
			zapFields = append(zapFields, zap.Any(key, value))
			fieldsMu.Unlock()
		}(k, v)
	}
	wg.Wait()

	return &Logger{
		zap:    l.zap.With(zapFields...),
		tracer: l.tracer,
	}
}

// Debug logs debug level message
func (l *Logger) Debug(msg string, fields ...Fields) {
	l.log(DebugLevel, msg, fields...)
}

// Info logs info level message
func (l *Logger) Info(msg string, fields ...Fields) {
	l.log(InfoLevel, msg, fields...)
}

// Warn logs warning level message
func (l *Logger) Warn(msg string, fields ...Fields) {
	l.log(WarnLevel, msg, fields...)
}

// Error logs error level message
func (l *Logger) Error(msg string, fields ...Fields) {
	l.log(ErrorLevel, msg, fields...)
}

// ErrorWithContext logs error with trace context
func (l *Logger) ErrorWithContext(ctx context.Context, msg string, err error, fields ...Fields) {
	mergedFields := mergeFields(fields...)
	if err != nil {
		mergedFields["error"] = err.Error()
	}

	l.WithContext(ctx).log(ErrorLevel, msg, mergedFields)

	// Record error in trace span if available
	if span := trace.SpanFromContext(ctx); span != nil {
		span.RecordError(err)
	}
}

func (l *Logger) log(level LogLevel, msg string, fields ...Fields) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	mergedFields := mergeFields(fields...)

	zapFields := make([]zap.Field, 0, len(mergedFields))
	var fieldsMu sync.Mutex

	var wg sync.WaitGroup
	for k, v := range mergedFields {
		wg.Add(1)
		go func(key string, value interface{}) {
			defer wg.Done()
			fieldsMu.Lock()
			zapFields = append(zapFields, zap.Any(key, value))
			fieldsMu.Unlock()
		}(k, v)
	}
	wg.Wait()

	switch level {
	case DebugLevel:
		l.zap.Debug(msg, zapFields...)
	case InfoLevel:
		l.zap.Info(msg, zapFields...)
	case WarnLevel:
		l.zap.Warn(msg, zapFields...)
	case ErrorLevel:
		l.zap.Error(msg, zapFields...)
	}
}

var mergeFieldsMu sync.Mutex

func mergeFields(fields ...Fields) Fields {
	mergeFieldsMu.Lock()
	defer mergeFieldsMu.Unlock()

	merged := Fields{}
	for _, f := range fields {
		for k, v := range f {
			merged[k] = v
		}
	}
	return merged
}

// LogRequestResponse logs API request/response details
func (l *Logger) LogRequestResponse(ctx context.Context, requestID string, method, path string, requestBody, responseBody interface{}, duration time.Duration, err error) {
	fields := Fields{
		"request_id":   requestID,
		"method":       method,
		"path":         path,
		"duration_ms":  duration.Milliseconds(),
		"request_time": time.Now().Format(time.RFC3339),
	}

	if requestBody != nil {
		if reqJSON, err := json.Marshal(requestBody); err == nil {
			fields["request_body"] = string(reqJSON)
		}
	}

	if responseBody != nil {
		if respJSON, err := json.Marshal(responseBody); err == nil {
			fields["response_body"] = string(respJSON)
		}
	}

	if err != nil {
		fields["error"] = err.Error()
		l.WithContext(ctx).Error("API request failed", fields)
		return
	}

	l.WithContext(ctx).Info("API request completed", fields)
}

// LogMetric logs metric data
func (l *Logger) LogMetric(ctx context.Context, name string, value interface{}, fields ...Fields) {
	mergedFields := mergeFields(fields...)
	mergedFields["metric_name"] = name
	mergedFields["metric_value"] = value
	mergedFields["metric_time"] = time.Now().Format(time.RFC3339)

	l.WithContext(ctx).Info("Metric recorded", mergedFields)
}

// Close flushes any buffered log entries
func (l *Logger) Close() error {
	l.mu.RLock()
	defer l.mu.RUnlock()
	_ = l.zap.Sync() // Ignore sync errors as they're not relevant for cleanup
	return nil
}

func (l *Logger) LogRequest(requestBody interface{}, responseBody interface{}, err error) {
	fields := Fields{
		"timestamp": time.Now().Format(time.RFC3339),
	}

	if requestBody != nil {
		reqJSON, jsonErr := json.Marshal(requestBody)
		if jsonErr == nil {
			fields["request"] = string(reqJSON)
		}
	}

	if responseBody != nil {
		respJSON, jsonErr := json.Marshal(responseBody)
		if jsonErr == nil {
			fields["response"] = string(respJSON)
		}
	}

	if err != nil {
		fields["error"] = err.Error()
		l.WithFields(fields).Error("Request failed")
	} else {
		l.WithFields(fields).Info("Request completed")
	}
}
