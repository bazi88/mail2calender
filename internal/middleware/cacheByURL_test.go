package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCacheByURL(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		expectedStatus int
		expectCache    bool
	}{
		{
			name:           "valid URL with params",
			url:            "/test?param=value",
			expectedStatus: http.StatusOK,
			expectCache:    true,
		},
		{
			name:           "valid URL without params",
			url:            "/test",
			expectedStatus: http.StatusOK,
			expectCache:    true,
		},
		{
			name:           "root URL",
			url:            "/",
			expectedStatus: http.StatusOK,
			expectCache:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a handler that checks for cache key in context
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Get cache key from context
				cacheKey, ok := r.Context().Value(CacheURL).(string)
				if tt.expectCache {
					assert.True(t, ok)
					assert.NotEmpty(t, cacheKey)
				}
				w.WriteHeader(tt.expectedStatus)
			})

			// Create the middleware handler
			handler := CacheByURL(nextHandler)

			// Create test request
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			rec := httptest.NewRecorder()

			// Call the handler
			handler.ServeHTTP(rec, req)

			// Check response
			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestSum(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		expectError bool
	}{
		{
			name:        "valid URL with params",
			url:         "/test?param=value",
			expectError: false,
		},
		{
			name:        "valid URL without params",
			url:         "/test",
			expectError: false,
		},
		{
			name:        "empty URL",
			url:         "",
			expectError: false,
		},
		{
			name:        "root URL",
			url:         "/",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := sum(tt.url)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, hash)

				// Verify hash is consistent
				hash2, err := sum(tt.url)
				assert.NoError(t, err)
				assert.Equal(t, hash, hash2)
			}
		})
	}
}
