package errors

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCalendarError_Error(t *testing.T) {
	tests := []struct {
		name        string
		err         *CalendarError
		expectedMsg string
	}{
		{
			name:        "basic error",
			err:         NewError("TEST_ERROR", "test message"),
			expectedMsg: "TEST_ERROR: test message",
		},
		{
			name: "error with wrapped error",
			err: NewError("TEST_ERROR", "test message").
				WithWrappedError(errors.New("inner error")),
			expectedMsg: "TEST_ERROR: test message (caused by: inner error)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedMsg, tt.err.Error())
		})
	}
}

func TestCalendarError_WithDetails(t *testing.T) {
	err := NewError("TEST_ERROR", "test message").
		WithDetails(map[string]interface{}{
			"key1": "value1",
			"key2": 123,
		})

	assert.Equal(t, "value1", err.Details["key1"])
	assert.Equal(t, 123, err.Details["key2"])
}

func TestCalendarError_WithRetry(t *testing.T) {
	retryDuration := 5 * time.Second
	err := NewError("TEST_ERROR", "test message").
		WithRetry(retryDuration)

	assert.Equal(t, &retryDuration, err.RetryAfter)
}

// ErrorConstructor is a type for error constructors
type ErrorConstructor = func(string) *CalendarError

func TestErrorConstructors(t *testing.T) {
	tests := []struct {
		name        string
		constructor ErrorConstructor
		errType     string
	}{
		{
			name:        "invalid email error",
			constructor: NewInvalidEmailError,
			errType:     InvalidEmail,
		},
		{
			name:        "invalid token error",
			constructor: NewInvalidTokenError,
			errType:     InvalidToken,
		},
		{
			name:        "invalid time error",
			constructor: NewInvalidTimeError,
			errType:     InvalidTime,
		},
		{
			name:        "conflict error",
			constructor: NewConflictError,
			errType:     ConflictDetected,
		},
		{
			name:        "service unavailable error",
			constructor: NewServiceUnavailableError,
			errType:     ServiceUnavailable,
		},
		{
			name:        "parse error",
			constructor: NewParseError,
			errType:     ParseError,
		},
		{
			name:        "validation error",
			constructor: NewValidationError,
			errType:     ValidationError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.constructor("test message")
			assert.Equal(t, tt.errType, err.Type)
			assert.Equal(t, "test message", err.Message)
		})
	}
}

func TestErrorTypeChecks(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		checker  func(error) bool
		expected bool
	}{
		{
			name:     "is invalid email",
			err:      NewInvalidEmailError("test"),
			checker:  IsInvalidEmail,
			expected: true,
		},
		{
			name:     "is invalid token",
			err:      NewInvalidTokenError("test"),
			checker:  IsInvalidToken,
			expected: true,
		},
		{
			name:     "is invalid time",
			err:      NewInvalidTimeError("test"),
			checker:  IsInvalidTime,
			expected: true,
		},
		{
			name:     "is conflict",
			err:      NewConflictError("test"),
			checker:  IsConflict,
			expected: true,
		},
		{
			name:     "is service unavailable",
			err:      NewServiceUnavailableError("test"),
			checker:  IsServiceUnavailable,
			expected: true,
		},
		{
			name:     "is parse error",
			err:      NewParseError("test"),
			checker:  IsParseError,
			expected: true,
		},
		{
			name:     "is validation error",
			err:      NewValidationError("test"),
			checker:  IsValidationError,
			expected: true,
		},
		{
			name:     "wrong error type",
			err:      errors.New("test"),
			checker:  IsValidationError,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.checker(tt.err))
		})
	}
}

func TestShouldRetry(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "service unavailable error",
			err:      NewServiceUnavailableError("test"),
			expected: true,
		},
		{
			name: "error with retry duration",
			err: NewError("TEST_ERROR", "test").
				WithRetry(5 * time.Second),
			expected: true,
		},
		{
			name:     "non-retryable error",
			err:      NewValidationError("test"),
			expected: false,
		},
		{
			name:     "non-calendar error",
			err:      errors.New("test"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, ShouldRetry(tt.err))
		})
	}
}

func TestGetRetryAfter(t *testing.T) {
	retryDuration := 5 * time.Second
	tests := []struct {
		name     string
		err      error
		expected *time.Duration
	}{
		{
			name: "error with retry duration",
			err: NewError("TEST_ERROR", "test").
				WithRetry(retryDuration),
			expected: &retryDuration,
		},
		{
			name:     "error without retry duration",
			err:      NewValidationError("test"),
			expected: nil,
		},
		{
			name:     "non-calendar error",
			err:      errors.New("test"),
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, GetRetryAfter(tt.err))
		})
	}
}

func TestGetErrorDetails(t *testing.T) {
	details := map[string]interface{}{
		"key": "value",
	}

	tests := []struct {
		name     string
		err      error
		expected map[string]interface{}
	}{
		{
			name: "error with details",
			err: NewError("TEST_ERROR", "test").
				WithDetails(details),
			expected: details,
		},
		{
			name:     "error without details",
			err:      NewValidationError("test"),
			expected: map[string]interface{}{},
		},
		{
			name:     "non-calendar error",
			err:      errors.New("test"),
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetErrorDetails(tt.err)
			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestGetErrorTime(t *testing.T) {
	now := time.Now()
	err := NewError("TEST_ERROR", "test")

	// The error time should be close to now
	assert.WithinDuration(t, now, GetErrorTime(err), time.Second)

	// Non-calendar error should return zero time
	assert.True(t, GetErrorTime(errors.New("test")).IsZero())
}
