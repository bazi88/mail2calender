package logger

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "default production config",
			config: Config{
				Environment: "production",
				Level:       "info",
				OutputPaths: []string{"stdout"},
			},
			wantErr: false,
		},
		{
			name: "development config",
			config: Config{
				Environment: "development",
				Level:       "debug",
				OutputPaths: []string{"stdout"},
			},
			wantErr: false,
		},
		{
			name: "invalid level",
			config: Config{
				Environment: "production",
				Level:       "invalid",
				OutputPaths: []string{"stdout"},
			},
			wantErr: false, // Should not error, defaults to info
		},
		{
			name: "invalid output path",
			config: Config{
				Environment: "production",
				Level:       "info",
				OutputPaths: []string{"/invalid/path/that/does/not/exist"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := NewLogger(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, logger)
		})
	}
}

func TestLogger_Levels(t *testing.T) {
	core, logs := observer.New(zapcore.DebugLevel)
	testLogger := &loggerImpl{logger: zap.New(core)}

	ctx := context.Background()
	testCases := []struct {
		level   string
		logFunc func(string, ...zapcore.Field)
	}{
		{
			level: "debug",
			logFunc: func(msg string, fields ...zapcore.Field) {
				testLogger.Debug(ctx, msg, fields...)
			},
		},
		{
			level: "info",
			logFunc: func(msg string, fields ...zapcore.Field) {
				testLogger.Info(ctx, msg, fields...)
			},
		},
		{
			level: "warn",
			logFunc: func(msg string, fields ...zapcore.Field) {
				testLogger.Warn(ctx, msg, fields...)
			},
		},
		{
			level: "error",
			logFunc: func(msg string, fields ...zapcore.Field) {
				testLogger.Error(ctx, msg, fields...)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.level, func(t *testing.T) {
			msg := "test message"
			tc.logFunc(msg)
			require.NotZero(t, logs.Len())
			lastLog := logs.All()[logs.Len()-1]
			assert.Equal(t, msg, lastLog.Message)
		})
	}
}

func TestLogger_WithFields(t *testing.T) {
	core, logs := observer.New(zapcore.DebugLevel)
	testLogger := &loggerImpl{logger: zap.New(core)}

	ctx := context.Background()
	fields := []zapcore.Field{
		zap.String("key1", "value1"),
		zap.Int("key2", 42),
	}

	// Create a logger with fields
	loggerWithFields := testLogger.With(fields...)

	// Log a message
	msg := "test with fields"
	loggerWithFields.Info(ctx, msg)

	// Verify the log entry
	require.NotZero(t, logs.Len())
	lastLog := logs.All()[logs.Len()-1]
	assert.Equal(t, msg, lastLog.Message)

	// Verify fields are present
	contextMap := make(map[string]interface{})
	for _, field := range lastLog.Context {
		contextMap[field.Key] = field.Interface
	}

	assert.Equal(t, "value1", contextMap["key1"])
	assert.Equal(t, 42, contextMap["key2"])
}

func TestLogger_TraceContext(t *testing.T) {
	core, logs := observer.New(zapcore.DebugLevel)
	testLogger := &loggerImpl{logger: zap.New(core)}

	// Create a traced context
	tracer := otel.Tracer("test-tracer")
	ctx, span := tracer.Start(context.Background(), "test-span")
	defer span.End()

	// Log with traced context
	msg := "test with trace"
	testLogger.Info(ctx, msg)

	// Verify trace information is included
	require.NotZero(t, logs.Len())
	lastLog := logs.All()[logs.Len()-1]
	assert.Equal(t, msg, lastLog.Message)

	// Get the span context
	spanCtx := trace.SpanFromContext(ctx).SpanContext()
	if spanCtx.IsValid() {
		contextMap := make(map[string]interface{})
		for _, field := range lastLog.Context {
			contextMap[field.Key] = field.Interface
		}

		assert.Contains(t, contextMap, "trace_id")
		assert.Contains(t, contextMap, "span_id")
	}
}

func TestLogger_Fatal(t *testing.T) {
	core, logs := observer.New(zapcore.DebugLevel)
	testLogger := &loggerImpl{logger: zap.New(core)}

	ctx := context.Background()
	msg := "fatal error"

	// Use a defer to prevent the test from actually exiting
	defer func() {
		if r := recover(); r != nil {
			require.NotZero(t, logs.Len())
			lastLog := logs.All()[logs.Len()-1]
			assert.Equal(t, msg, lastLog.Message)
		}
	}()

	testLogger.Fatal(ctx, msg)
}
