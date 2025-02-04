package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAudit(t *testing.T) {
	tests := []struct {
		name          string
		setupContext  bool
		setupHeaders  map[string]string
		expectedEvent Event
	}{
		{
			name:         "with user ID in context",
			setupContext: true,
			setupHeaders: map[string]string{
				"X-Real-Ip":  "1.2.3.4",
				"User-Agent": "test-agent",
			},
			expectedEvent: Event{
				ActorID:    123,
				HTTPMethod: "GET",
				URL:        "/test",
				IPAddress:  "1.2.3.4",
				UserAgent:  "test-agent",
			},
		},
		{
			name:         "without user ID in context",
			setupContext: false,
			setupHeaders: map[string]string{
				"X-Real-Ip":  "1.2.3.4",
				"User-Agent": "test-agent",
			},
			expectedEvent: Event{
				ActorID:    0,
				HTTPMethod: "GET",
				URL:        "/test",
				IPAddress:  "1.2.3.4",
				UserAgent:  "test-agent",
			},
		},
		{
			name:         "with X-Forwarded-For header",
			setupContext: false,
			setupHeaders: map[string]string{
				"X-Forwarded-For": "5.6.7.8",
				"User-Agent":      "test-agent",
			},
			expectedEvent: Event{
				ActorID:    0,
				HTTPMethod: "GET",
				URL:        "/test",
				IPAddress:  "5.6.7.8",
				UserAgent:  "test-agent",
			},
		},
		{
			name:         "with remote addr only",
			setupContext: false,
			setupHeaders: map[string]string{
				"User-Agent": "test-agent",
			},
			expectedEvent: Event{
				ActorID:    0,
				HTTPMethod: "GET",
				URL:        "/test",
				IPAddress:  "192.0.2.1:1234",
				UserAgent:  "test-agent",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a handler that will check the audit event in context
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Get the audit event from context
				ev, ok := r.Context().Value(KeyAuditID).(Event)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedEvent.ActorID, ev.ActorID)
				assert.Equal(t, tt.expectedEvent.HTTPMethod, ev.HTTPMethod)
				assert.Equal(t, tt.expectedEvent.URL, ev.URL)
				assert.Equal(t, tt.expectedEvent.IPAddress, ev.IPAddress)
				assert.Equal(t, tt.expectedEvent.UserAgent, ev.UserAgent)

				w.WriteHeader(http.StatusOK)
			})

			// Create the middleware handler
			handler := Audit(nextHandler)

			// Create test request
			req := httptest.NewRequest(http.MethodGet, "/test", nil)

			// Set up context if needed
			if tt.setupContext {
				ctx := context.WithValue(req.Context(), KeySession, uint64(123))
				req = req.WithContext(ctx)
			}

			// Set up headers
			for key, value := range tt.setupHeaders {
				req.Header.Set(key, value)
			}

			// Set remote addr for testing
			req.RemoteAddr = "192.0.2.1:1234"

			rec := httptest.NewRecorder()

			// Call the handler
			handler.ServeHTTP(rec, req)

			// Check response
			assert.Equal(t, http.StatusOK, rec.Code)
		})
	}
}

func TestGetUserID(t *testing.T) {
	tests := []struct {
		name         string
		setupContext bool
		expectedID   uint64
	}{
		{
			name:         "with user ID in context",
			setupContext: true,
			expectedID:   123,
		},
		{
			name:         "without user ID in context",
			setupContext: false,
			expectedID:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)

			if tt.setupContext {
				ctx := context.WithValue(req.Context(), KeySession, uint64(123))
				req = req.WithContext(ctx)
			}

			userID := getUserID(req)
			assert.Equal(t, tt.expectedID, userID)
		})
	}
}

func TestReadUserIP(t *testing.T) {
	tests := []struct {
		name       string
		headers    map[string]string
		remoteAddr string
		expectedIP string
	}{
		{
			name: "with X-Real-Ip",
			headers: map[string]string{
				"X-Real-Ip": "1.2.3.4",
			},
			remoteAddr: "192.0.2.1:1234",
			expectedIP: "1.2.3.4",
		},
		{
			name: "with X-Forwarded-For",
			headers: map[string]string{
				"X-Forwarded-For": "5.6.7.8",
			},
			remoteAddr: "192.0.2.1:1234",
			expectedIP: "5.6.7.8",
		},
		{
			name:       "with remote addr only",
			headers:    map[string]string{},
			remoteAddr: "192.0.2.1:1234",
			expectedIP: "192.0.2.1:1234",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)

			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}
			req.RemoteAddr = tt.remoteAddr

			ip := readUserIP(req)
			assert.Equal(t, tt.expectedIP, ip)
		})
	}
}
