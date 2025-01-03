package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSecurityHeaders(t *testing.T) {
	// Test với cấu hình mặc định
	t.Run("Default Config", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()

		SecurityHeaders(handler).ServeHTTP(rr, req)

		headers := rr.Header()
		assert.Equal(t, "nosniff", headers.Get("X-Content-Type-Options"))
		assert.Equal(t, "DENY", headers.Get("X-Frame-Options"))
		assert.Equal(t, "1; mode=block", headers.Get("X-XSS-Protection"))
		assert.Contains(t, headers.Get("Content-Security-Policy"), "default-src 'self'")
	})

	// Test với cấu hình tùy chỉnh
	t.Run("Custom Config", func(t *testing.T) {
		config := &SecurityConfig{
			CSPDirectives: map[string][]string{
				"default-src": {"'self'", "https://api.example.com"},
				"script-src":  {"'self'", "https://cdn.example.com"},
			},
			HSTSMaxAge:            3600,
			HSTSIncludeSubdomains: false,
			FrameOptions:          "SAMEORIGIN",
			CustomHeaders: map[string]string{
				"X-Custom-Header": "test-value",
			},
		}

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()

		SecurityHeadersWithConfig(config)(handler).ServeHTTP(rr, req)

		headers := rr.Header()

		// Kiểm tra CSP tùy chỉnh
		csp := headers.Get("Content-Security-Policy")
		assert.Contains(t, csp, "default-src 'self' https://api.example.com")
		assert.Contains(t, csp, "script-src 'self' https://cdn.example.com")

		// Kiểm tra HSTS tùy chỉnh
		hsts := headers.Get("Strict-Transport-Security")
		assert.Contains(t, hsts, "max-age=3600")
		assert.NotContains(t, hsts, "includeSubDomains")

		// Kiểm tra Frame Options tùy chỉnh
		assert.Equal(t, "SAMEORIGIN", headers.Get("X-Frame-Options"))

		// Kiểm tra Custom Header
		assert.Equal(t, "test-value", headers.Get("X-Custom-Header"))
	})

	// Test với CSP cho API endpoints
	t.Run("API Config", func(t *testing.T) {
		config := &SecurityConfig{
			CSPDirectives: map[string][]string{
				"default-src": {"'none'"},
				"connect-src": {"'self'"},
			},
			FrameOptions: "DENY",
			CustomHeaders: map[string]string{
				"X-API-Version": "1.0",
			},
		}

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest("GET", "/api/test", nil)
		rr := httptest.NewRecorder()

		SecurityHeadersWithConfig(config)(handler).ServeHTTP(rr, req)

		headers := rr.Header()
		assert.Contains(t, headers.Get("Content-Security-Policy"), "default-src 'none'")
		assert.Contains(t, headers.Get("Content-Security-Policy"), "connect-src 'self'")
		assert.Equal(t, "1.0", headers.Get("X-API-Version"))
	})

	// Test với Feature Policy tùy chỉnh
	t.Run("Feature Policy Config", func(t *testing.T) {
		config := &SecurityConfig{
			FeaturePolicy: map[string]string{
				"camera":      "self https://meet.example.com",
				"microphone":  "self https://meet.example.com",
				"geolocation": "self",
			},
		}

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()

		SecurityHeadersWithConfig(config)(handler).ServeHTTP(rr, req)

		pp := rr.Header().Get("Permissions-Policy")
		assert.Contains(t, pp, "camera=self https://meet.example.com")
		assert.Contains(t, pp, "microphone=self https://meet.example.com")
		assert.Contains(t, pp, "geolocation=self")
	})
}

func TestSecurityHeadersNilConfig(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	SecurityHeadersWithConfig(nil)(handler).ServeHTTP(rr, req)

	// Kiểm tra xem có sử dụng cấu hình mặc định không
	headers := rr.Header()
	assert.Equal(t, "nosniff", headers.Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", headers.Get("X-Frame-Options"))
}
