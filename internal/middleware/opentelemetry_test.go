package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace"
)

func TestOtlp(t *testing.T) {
	tests := []struct {
		name      string
		enable    bool
		setupPath bool
	}{
		{
			name:      "enabled with route pattern",
			enable:    true,
			setupPath: true,
		},
		{
			name:      "enabled without route pattern",
			enable:    true,
			setupPath: false,
		},
		{
			name:      "disabled",
			enable:    false,
			setupPath: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a handler that will be wrapped
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Get the span from context
				span := trace.SpanFromContext(r.Context())
				assert.NotNil(t, span)

				w.WriteHeader(http.StatusOK)
			})

			// Create the middleware handler
			handler := Otlp(tt.enable)(nextHandler)

			// Create test request
			req := httptest.NewRequest(http.MethodGet, "/test", nil)

			// If using router, set up the route context
			if tt.setupPath {
				rctx := chi.NewRouteContext()
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			}

			rec := httptest.NewRecorder()

			// Call the handler
			handler.ServeHTTP(rec, req)

			// Check response
			assert.Equal(t, http.StatusOK, rec.Code)
		})
	}
}
