package logger

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

type mockTracer struct {
	trace.Tracer
}

type mockSpan struct {
	trace.Span
	recordedError error
	mu            sync.RWMutex
}

func (s *mockSpan) RecordError(err error, opts ...trace.EventOption) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.recordedError = err
}

func (s *mockSpan) SpanContext() trace.SpanContext {
	return trace.SpanContext{}
}

// createTestLogger creates a logger with an observer for testing
func createTestLogger() (*Logger, *observer.ObservedLogs) {
	core, recorded := observer.New(zapcore.InfoLevel)
	logger := &Logger{
		zap:    zap.New(core),
		tracer: &mockTracer{},
	}
	return logger, recorded
}

func TestLogger_WithContext(t *testing.T) {
	// Create a test logger
	logger, err := New(&mockTracer{})
	assert.NoError(t, err)

	// Test with invalid context
	invalidCtx := context.Background()
	loggerWithCtx := logger.WithContext(invalidCtx)
	assert.NotNil(t, loggerWithCtx)

	// TODO: Add test with valid trace context when needed
}

func TestLogger_WithFields(t *testing.T) {
	logger, err := New(&mockTracer{})
	assert.NoError(t, err)

	fields := Fields{
		"key1": "value1",
		"key2": 123,
	}

	loggerWithFields := logger.WithFields(fields)
	assert.NotNil(t, loggerWithFields)
}

func TestLogger_LogLevels(t *testing.T) {
	logger, logs := createTestLogger()

	tests := []struct {
		name     string
		logFunc  func(string, ...Fields)
		message  string
		fields   Fields
		expected zapcore.Level
	}{
		{
			name:     "info level",
			logFunc:  logger.Info,
			message:  "info message",
			fields:   Fields{"test": "info"},
			expected: zapcore.InfoLevel,
		},
		{
			name:     "warn level",
			logFunc:  logger.Warn,
			message:  "warn message",
			fields:   Fields{"test": "warn"},
			expected: zapcore.WarnLevel,
		},
		{
			name:     "error level",
			logFunc:  logger.Error,
			message:  "error message",
			fields:   Fields{"test": "error"},
			expected: zapcore.ErrorLevel,
		},
	}

	for _, tt := range tests {
		tt := tt // Capture range variable
		t.Run(tt.name, func(t *testing.T) {
			// Clear logs before each test
			logs.TakeAll()

			// Execute log function
			tt.logFunc(tt.message, tt.fields)

			// Get logs after logging
			allLogs := logs.All()
			if assert.Len(t, allLogs, 1) {
				entry := allLogs[0]
				assert.Equal(t, tt.message, entry.Message)
				assert.Equal(t, tt.fields["test"], entry.ContextMap()["test"])
				assert.Equal(t, tt.expected, entry.Level)
			}
		})
	}
}

func TestLogger_ErrorWithContext(t *testing.T) {
	// Create test logger with mock span
	mockSpan := &mockSpan{}
	ctx := trace.ContextWithSpan(context.Background(), mockSpan)

	logger, err := New(&mockTracer{})
	assert.NoError(t, err)

	testError := errors.New("test error")
	logger.ErrorWithContext(ctx, "error occurred", testError)

	// Verify error was recorded in span
	mockSpan.mu.RLock()
	assert.Equal(t, testError, mockSpan.recordedError)
	mockSpan.mu.RUnlock()
}

func TestLogger_LogRequestResponse(t *testing.T) {
	logger, logs := createTestLogger()
	var mu sync.Mutex

	ctx := context.Background()
	requestBody := map[string]string{"key": "value"}
	responseBody := map[string]int{"count": 42}
	duration := 100 * time.Millisecond

	// Test successful request
	logger.LogRequestResponse(ctx, "req123", "GET", "/test", requestBody, responseBody, duration, nil)

	mu.Lock()
	allLogs := logs.All()
	assert.True(t, len(allLogs) > 0)
	lastLog := allLogs[len(allLogs)-1]
	assert.Equal(t, "API request completed", lastLog.Message)
	assert.Equal(t, "req123", lastLog.ContextMap()["request_id"])
	assert.Equal(t, "GET", lastLog.ContextMap()["method"])
	assert.Equal(t, "/test", lastLog.ContextMap()["path"])
	mu.Unlock()

	// Test failed request
	testError := errors.New("request failed")
	logger.LogRequestResponse(ctx, "req124", "POST", "/test", requestBody, nil, duration, testError)

	mu.Lock()
	allLogs = logs.All()
	lastLog = allLogs[len(allLogs)-1]
	assert.Equal(t, "API request failed", lastLog.Message)
	assert.Contains(t, lastLog.ContextMap()["error"], "request failed")
	mu.Unlock()
}

func TestLogger_LogMetric(t *testing.T) {
	logger, logs := createTestLogger()
	var mu sync.Mutex

	ctx := context.Background()
	logger.LogMetric(ctx, "test_metric", 42.0, Fields{"dimension": "test"})

	mu.Lock()
	allLogs := logs.All()
	assert.True(t, len(allLogs) > 0)
	lastLog := allLogs[len(allLogs)-1]
	assert.Equal(t, "Metric recorded", lastLog.Message)
	assert.Equal(t, "test_metric", lastLog.ContextMap()["metric_name"])
	assert.Equal(t, 42.0, lastLog.ContextMap()["metric_value"])
	assert.Equal(t, "test", lastLog.ContextMap()["dimension"])
	mu.Unlock()
}

func TestMergeFields(t *testing.T) {
	fields1 := Fields{"key1": "value1", "key2": 2}
	fields2 := Fields{"key2": "overwritten", "key3": true}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		merged := mergeFields(fields1, fields2)
		assert.Equal(t, "value1", merged["key1"])
		assert.Equal(t, "overwritten", merged["key2"])
		assert.Equal(t, true, merged["key3"])
	}()
	wg.Wait()
}

func TestLogger_Close(t *testing.T) {
	logger, err := New(&mockTracer{})
	assert.NoError(t, err)

	err = logger.Close()
	assert.NoError(t, err)
}
