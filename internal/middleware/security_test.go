package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSecurityHeaders(t *testing.T) {
	tests := []struct {
		name            string
		config          *SecurityConfig
		expectedHeaders map[string]string
	}{
		{
			name: "Default Config",
			config: &SecurityConfig{
				HSTSMaxAge:            31536000,
				HSTSIncludeSubdomains: true,
				CSPDirectives: map[string][]string{
					"default-src": {"'self'"},
					"script-src":  {"'self'"},
					"style-src":   {"'self'"},
				},
				FrameOptions:        "DENY",
				XContentTypeOptions: "nosniff",
				ReferrerPolicy:      "strict-origin-when-cross-origin",
				CustomHeaders: map[string]string{
					"X-XSS-Protection": "1; mode=block",
				},
			},
			expectedHeaders: map[string]string{
				"Content-Security-Policy":   "default-src 'self'; script-src 'self'; style-src 'self'",
				"Strict-Transport-Security": "max-age=31536000; includeSubDomains",
				"X-Frame-Options":           "DENY",
				"X-Content-Type-Options":    "nosniff",
				"Referrer-Policy":           "strict-origin-when-cross-origin",
				"X-XSS-Protection":          "1; mode=block",
			},
		},
		{
			name: "Custom Config",
			config: &SecurityConfig{
				HSTSMaxAge:            3600,
				HSTSIncludeSubdomains: false,
				CSPDirectives: map[string][]string{
					"default-src": {"'self'", "https://api.example.com"},
					"script-src":  {"'self'", "https://cdn.example.com"},
				},
				FrameOptions:        "SAMEORIGIN",
				XContentTypeOptions: "nosniff",
				CustomHeaders: map[string]string{
					"X-Custom-Header": "test-value",
				},
			},
			expectedHeaders: map[string]string{
				"Content-Security-Policy":   "default-src 'self' https://api.example.com; script-src 'self' https://cdn.example.com",
				"Strict-Transport-Security": "max-age=3600",
				"X-Frame-Options":           "SAMEORIGIN",
				"X-Content-Type-Options":    "nosniff",
				"X-Custom-Header":           "test-value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest("GET", "/test", nil)
			rr := httptest.NewRecorder()

			SecurityHeadersWithConfig(tt.config)(handler).ServeHTTP(rr, req)

			for key, expectedValue := range tt.expectedHeaders {
				assert.Equal(t, expectedValue, rr.Header().Get(key))
			}
		})
	}
}

func TestRateLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := &RateLimitConfig{
		Requests: 2,
		Window:   time.Second,
	}

	router := gin.New()
	router.Use(RateLimit(config))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "test")
	})

	// First request should succeed
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Second request should succeed
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Third request should be rate limited
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)

	// Wait for window to expire
	time.Sleep(time.Second)

	// Request should succeed again
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCORS(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := &CORSConfig{
		AllowOrigins: []string{"http://example.com"},
		AllowMethods: []string{"GET", "POST"},
		AllowHeaders: []string{"Content-Type"},
		MaxAge:       3600,
	}

	router := gin.New()
	router.Use(CORS(config))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "test")
	})

	t.Run("Preflight Request", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("OPTIONS", "/test", nil)
		req.Header.Set("Origin", "http://example.com")
		req.Header.Set("Access-Control-Request-Method", "POST")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "http://example.com", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "GET,POST", w.Header().Get("Access-Control-Allow-Methods"))
		assert.Equal(t, "Content-Type", w.Header().Get("Access-Control-Allow-Headers"))
		assert.Equal(t, "3600", w.Header().Get("Access-Control-Max-Age"))
	})

	t.Run("Normal Request", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "http://example.com")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "http://example.com", w.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("Invalid Origin", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "http://invalid.com")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}
