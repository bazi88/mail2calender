package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"mono-golang/internal/pkg/security"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
)

func TestSecurityHeaders(t *testing.T) {
	tests := []struct {
		name            string
		config          *security.SecurityConfig
		expectedHeaders map[string]string
	}{
		{
			name: "Default Config",
			config: &security.SecurityConfig{
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
			config: &security.SecurityConfig{
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

func TestSecurityRateLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})
	defer redisClient.Close()

	limiter := NewRedisRateLimiter(redisClient, 2, time.Second)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("Allow requests within limit", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()

		limiter.Limit(handler).ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)

		rr = httptest.NewRecorder()
		limiter.Limit(handler).ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Block requests over limit", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()

		limiter.Limit(handler).ServeHTTP(rr, req)
		assert.Equal(t, http.StatusTooManyRequests, rr.Code)
		assert.NotEmpty(t, rr.Header().Get("Retry-After"))
	})

	t.Run("Reset limit after window", func(t *testing.T) {
		time.Sleep(time.Second)

		req := httptest.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()

		limiter.Limit(handler).ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	})
}

func TestCORS(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(SecurityMiddleware())
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
	})

	t.Run("Normal Request", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "http://example.com")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
