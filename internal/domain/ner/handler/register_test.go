package handler

import (
	"net/http"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRateLimiter is a mock implementation of middleware.RateLimiter
type MockRateLimiter struct {
	mock.Mock
}

func (m *MockRateLimiter) Limit(next http.Handler) http.Handler {
	args := m.Called(next)
	if handler, ok := args.Get(0).(http.Handler); ok {
		return handler
	}
	return next
}

func TestRegister(t *testing.T) {
	mockUseCase := new(MockNERUseCase)
	mockRateLimiter := new(MockRateLimiter)
	router := chi.NewRouter()

	// Setup rate limiter mock
	mockRateLimiter.On("Limit", mock.AnythingOfType("http.HandlerFunc")).Return(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	Register(router, mockUseCase, mockRateLimiter)

	walkFunc := func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		if route == "/api/v1/ner/extract" {
			assert.Equal(t, http.MethodPost, method)
			assert.NotNil(t, handler)
		}
		return nil
	}

	err := chi.Walk(router, walkFunc)
	assert.NoError(t, err)
	mockRateLimiter.AssertExpectations(t)
}
