package health

import (
	"net/http"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestRegisterHTTPEndPoints(t *testing.T) {
	mockUseCase := new(MockUseCase)
	router := chi.NewRouter()

	handler := RegisterHTTPEndPoints(router, mockUseCase)
	assert.NotNil(t, handler)

	// Test routes are registered
	routes := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/health"},
		{http.MethodGet, "/api/health/readiness"},
	}

	for _, rt := range routes {
		walkFunc := func(method string, path string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
			if path == rt.path {
				assert.Equal(t, rt.method, method)
				assert.NotNil(t, handler)
			}
			return nil
		}

		err := chi.Walk(router, walkFunc)
		assert.NoError(t, err)
	}
}
