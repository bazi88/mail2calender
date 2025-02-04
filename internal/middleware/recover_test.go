package middleware

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRecovery(t *testing.T) {
	tests := []struct {
		name         string
		handler      func(w http.ResponseWriter, r *http.Request)
		expectedCode int
		expectPanic  bool
	}{
		{
			name: "no panic",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			expectedCode: http.StatusOK,
			expectPanic:  false,
		},
		{
			name: "with panic",
			handler: func(w http.ResponseWriter, r *http.Request) {
				panic("test panic")
			},
			expectedCode: http.StatusInternalServerError,
			expectPanic:  true,
		},
		{
			name: "with http.ErrAbortHandler",
			handler: func(w http.ResponseWriter, r *http.Request) {
				panic(http.ErrAbortHandler)
			},
			expectedCode: http.StatusOK, // The panic will be ignored
			expectPanic:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture log output
			var buf bytes.Buffer
			log.SetOutput(&buf)
			defer log.SetOutput(os.Stderr)

			// Create the handler
			nextHandler := http.HandlerFunc(tt.handler)
			handler := Recovery(nextHandler)

			// Create test request
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()

			// Call the handler
			handler.ServeHTTP(rec, req)

			// Check response code
			assert.Equal(t, tt.expectedCode, rec.Code)

			// Check log output for panic cases
			if tt.expectPanic {
				assert.True(t, strings.Contains(buf.String(), "PANIC:"))
				assert.True(t, strings.Contains(buf.String(), "request:"))
				assert.True(t, strings.Contains(buf.String(), "host:"))
			} else {
				assert.Empty(t, buf.String())
			}
		})
	}
}
