package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContentType(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		expectedType string
	}{
		{
			name:         "png file",
			path:         "/image.png",
			expectedType: "image/png",
		},
		{
			name:         "css file",
			path:         "/style.css",
			expectedType: "text/css",
		},
		{
			name:         "js file",
			path:         "/script.js",
			expectedType: "application/javascript",
		},
		{
			name:         "json file",
			path:         "/data.json",
			expectedType: "application/json",
		},
		{
			name:         "ico file",
			path:         "/favicon.ico",
			expectedType: "image/icon",
		},
		{
			name:         "html file",
			path:         "/index.html",
			expectedType: "text/html",
		},
		{
			name:         "unknown extension",
			path:         "/file.unknown",
			expectedType: "text/html",
		},
		{
			name:         "no extension",
			path:         "/path",
			expectedType: "text/html",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock handler that will be wrapped
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

			// Create the middleware handler
			handler := ContentType(nextHandler)

			// Create a test request
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			// Call the handler
			handler.ServeHTTP(rec, req)

			// Check the response
			assert.Equal(t, tt.expectedType, rec.Header().Get("Content-Type"))
		})
	}
}
