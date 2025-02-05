package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJson(t *testing.T) {
	// Create a mock handler that will be wrapped
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(`{"message": "test"}`))
		if err != nil {
			t.Fatal(err)
		}
	})

	// Create the middleware handler
	handler := Json(nextHandler)

	// Create a test request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	// Call the handler
	handler.ServeHTTP(rec, req)

	// Check the response
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	assert.Equal(t, `{"message": "test"}`, rec.Body.String())
}
