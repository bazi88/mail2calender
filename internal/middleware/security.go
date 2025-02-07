package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"mail2calendar/internal/pkg/security"

	"github.com/gin-gonic/gin"
)

// setSecurityHeaders sets security headers based on the provided map
func setSecurityHeaders(w http.ResponseWriter, headers map[string]string) {
	for key, value := range headers {
		if value != "" {
			w.Header().Set(key, value)
		}
	}
	log.Printf("Security headers set: %v", headers)
}

// SecurityHeaders middleware adds security headers with default configuration
func SecurityHeaders(next http.Handler) http.Handler {
	return SecurityHeadersWithConfig(nil)(next)
}

// SecurityHeadersWithConfig creates middleware that adds security headers with custom configuration
func SecurityHeadersWithConfig(config *security.SecurityConfig) func(http.Handler) http.Handler {
	// If config is nil, use default configuration
	if config == nil {
		config = security.DefaultSecurityConfig()
	}

	// Validate config before creating middleware
	if err := security.ValidateConfig(config); err != nil {
		log.Printf("Security config validation failed: %v", err)
		config = security.DefaultSecurityConfig()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hstsValue := fmt.Sprintf("max-age=%d", config.HSTSMaxAge)
			if config.HSTSIncludeSubdomains {
				hstsValue += "; includeSubDomains"
			}

			headers := map[string]string{
				"Strict-Transport-Security": hstsValue,
				"Content-Security-Policy":   security.BuildCSP(config.CSPDirectives),
				"X-Frame-Options":           config.FrameOptions,
				"X-Content-Type-Options":    config.XContentTypeOptions,
				"Referrer-Policy":           config.ReferrerPolicy,
				"Permissions-Policy":        security.BuildFeaturePolicy(config.FeaturePolicy),
			}

			// Add custom headers
			for key, value := range config.CustomHeaders {
				headers[key] = value
			}

			setSecurityHeaders(w, headers)
			next.ServeHTTP(w, r)
		})
	}
}

// SetSecurityHeaders thiết lập các tiêu đề bảo mật cho phản hồi
func SetSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", "default-src 'self'")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		next.ServeHTTP(w, r)
	})
}

// SecurityMiddleware adds timeout for security checks
func SecurityMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
	}
}
