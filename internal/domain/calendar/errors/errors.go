package errors

import (
	"fmt"
	"time"
)

// Error types
const (
	InvalidEmail       = "INVALID_EMAIL"
	InvalidToken       = "INVALID_TOKEN"
	InvalidTime        = "INVALID_TIME"
	ConflictDetected   = "CONFLICT_DETECTED"
	ServiceUnavailable = "SERVICE_UNAVAILABLE"
	ParseError         = "PARSE_ERROR"
	ValidationError    = "VALIDATION_ERROR"
)

// CalendarError represents a domain-specific error
type CalendarError struct {
	Type       string
	Message    string
	Details    map[string]interface{}
	Time       time.Time
	RetryAfter *time.Duration
	WrappedErr error
}

func (e *CalendarError) Error() string {
	if e.WrappedErr != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Type, e.Message, e.WrappedErr)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Is implements error interface for error comparison
func (e *CalendarError) Is(target error) bool {
	if t, ok := target.(*CalendarError); ok {
		return e.Type == t.Type
	}
	return false
}

// NewError creates a new CalendarError
func NewError(errType string, message string) *CalendarError {
	return &CalendarError{
		Type:    errType,
		Message: message,
		Time:    time.Now(),
		Details: make(map[string]interface{}),
	}
}

// WithDetails adds additional error details
func (e *CalendarError) WithDetails(details map[string]interface{}) *CalendarError {
	for k, v := range details {
		e.Details[k] = v
	}
	return e
}

// WithRetry adds retry information
func (e *CalendarError) WithRetry(duration time.Duration) *CalendarError {
	e.RetryAfter = &duration
	return e
}

// WithWrappedError adds an underlying error
func (e *CalendarError) WithWrappedError(err error) *CalendarError {
	e.WrappedErr = err
	return e
}

// Common error constructors
func NewInvalidEmailError(message string) *CalendarError {
	return NewError(InvalidEmail, message)
}

func NewInvalidTokenError(message string) *CalendarError {
	return NewError(InvalidToken, message)
}

func NewInvalidTimeError(message string) *CalendarError {
	return NewError(InvalidTime, message)
}

func NewConflictError(message string) *CalendarError {
	return NewError(ConflictDetected, message)
}

func NewServiceUnavailableError(message string) *CalendarError {
	return NewError(ServiceUnavailable, message)
}

func NewParseError(message string) *CalendarError {
	return NewError(ParseError, message)
}

func NewValidationError(message string) *CalendarError {
	return NewError(ValidationError, message)
}

// Error utility functions
func IsInvalidEmail(err error) bool {
	if cerr, ok := err.(*CalendarError); ok {
		return cerr.Type == InvalidEmail
	}
	return false
}

func IsInvalidToken(err error) bool {
	if cerr, ok := err.(*CalendarError); ok {
		return cerr.Type == InvalidToken
	}
	return false
}

func IsInvalidTime(err error) bool {
	if cerr, ok := err.(*CalendarError); ok {
		return cerr.Type == InvalidTime
	}
	return false
}

func IsConflict(err error) bool {
	if cerr, ok := err.(*CalendarError); ok {
		return cerr.Type == ConflictDetected
	}
	return false
}

func IsServiceUnavailable(err error) bool {
	if cerr, ok := err.(*CalendarError); ok {
		return cerr.Type == ServiceUnavailable
	}
	return false
}

func IsParseError(err error) bool {
	if cerr, ok := err.(*CalendarError); ok {
		return cerr.Type == ParseError
	}
	return false
}

func IsValidationError(err error) bool {
	if cerr, ok := err.(*CalendarError); ok {
		return cerr.Type == ValidationError
	}
	return false
}

// ShouldRetry determines if the error is retryable
func ShouldRetry(err error) bool {
	if cerr, ok := err.(*CalendarError); ok {
		return cerr.Type == ServiceUnavailable || cerr.RetryAfter != nil
	}
	return false
}

// GetRetryAfter returns the suggested retry duration
func GetRetryAfter(err error) *time.Duration {
	if cerr, ok := err.(*CalendarError); ok {
		return cerr.RetryAfter
	}
	return nil
}

// GetErrorDetails returns the error details map
func GetErrorDetails(err error) map[string]interface{} {
	if cerr, ok := err.(*CalendarError); ok {
		return cerr.Details
	}
	return nil
}

// GetErrorTime returns when the error occurred
func GetErrorTime(err error) time.Time {
	if cerr, ok := err.(*CalendarError); ok {
		return cerr.Time
	}
	return time.Time{}
}
