package middleware

import (
	"fmt"
	"net/http"
	"strings"
)

// SecurityConfig chứa cấu hình cho security headers
type SecurityConfig struct {
	// CSP Configuration
	CSPDirectives map[string][]string

	// HSTS Configuration
	HSTSMaxAge            int
	HSTSIncludeSubdomains bool
	HSTSPreload           bool

	// Frame Options
	FrameOptions string // DENY, SAMEORIGIN, or ALLOW-FROM

	// Referrer Policy
	ReferrerPolicy string

	// Feature Policy
	FeaturePolicy map[string]string

	// Additional Security Headers
	CustomHeaders map[string]string

	// Content Type Options
	ContentTypeOptions string

	// XSS Protection
	XSSProtection string

	// Cross-Origin Policies
	CrossOriginEmbedderPolicy string
	CrossOriginOpenerPolicy   string
	CrossOriginResourcePolicy string
}

// DefaultSecurityConfig trả về cấu hình mặc định
func DefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		CSPDirectives: map[string][]string{
			"default-src": {"'self'"},
			"script-src":  {"'self'", "'unsafe-inline'", "'unsafe-eval'"},
			"style-src":   {"'self'", "'unsafe-inline'"},
			"img-src":     {"'self'", "data:", "https:"},
			"font-src":    {"'self'"},
			"connect-src": {"'self'"},
			"frame-src":   {"'none'"},
			"object-src":  {"'none'"},
			"base-uri":    {"'self'"},
			"form-action": {"'self'"},
		},
		HSTSMaxAge:            31536000,
		HSTSIncludeSubdomains: true,
		HSTSPreload:           true,
		FrameOptions:          "DENY",
		ReferrerPolicy:        "strict-origin-when-cross-origin",
		FeaturePolicy: map[string]string{
			"accelerometer":   "()",
			"camera":          "()",
			"geolocation":     "()",
			"gyroscope":       "()",
			"magnetometer":    "()",
			"microphone":      "()",
			"payment":         "()",
			"usb":             "()",
			"interest-cohort": "()",
			"autoplay":        "()",
			"fullscreen":      "self",
		},
		ContentTypeOptions:        "nosniff",
		XSSProtection:             "1; mode=block",
		CrossOriginEmbedderPolicy: "require-corp",
		CrossOriginOpenerPolicy:   "same-origin",
		CrossOriginResourcePolicy: "same-origin",
		CustomHeaders: map[string]string{
			"X-Permitted-Cross-Domain-Policies": "none",
			"X-Download-Options":                "noopen",
			"X-DNS-Prefetch-Control":            "off",
		},
	}
}

// SecurityHeadersWithConfig tạo middleware với cấu hình tùy chỉnh
func SecurityHeadersWithConfig(config *SecurityConfig) func(http.Handler) http.Handler {
	if config == nil {
		config = DefaultSecurityConfig()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Content-Type Options
			w.Header().Set("X-Content-Type-Options", config.ContentTypeOptions)

			// Frame Options
			w.Header().Set("X-Frame-Options", config.FrameOptions)

			// XSS Protection
			w.Header().Set("X-XSS-Protection", config.XSSProtection)

			// HSTS
			if config.HSTSMaxAge > 0 {
				hsts := []string{
					fmt.Sprintf("max-age=%d", config.HSTSMaxAge),
				}
				if config.HSTSIncludeSubdomains {
					hsts = append(hsts, "includeSubDomains")
				}
				if config.HSTSPreload {
					hsts = append(hsts, "preload")
				}
				w.Header().Set("Strict-Transport-Security", strings.Join(hsts, "; "))
			}

			// Content Security Policy
			if len(config.CSPDirectives) > 0 {
				var csp []string
				for directive, values := range config.CSPDirectives {
					csp = append(csp, directive+" "+strings.Join(values, " "))
				}
				w.Header().Set("Content-Security-Policy", strings.Join(csp, "; "))
			}

			// Referrer Policy
			w.Header().Set("Referrer-Policy", config.ReferrerPolicy)

			// Feature Policy
			if len(config.FeaturePolicy) > 0 {
				var fp []string
				for feature, value := range config.FeaturePolicy {
					fp = append(fp, feature+"="+value)
				}
				w.Header().Set("Permissions-Policy", strings.Join(fp, ", "))
			}

			// Cross-Origin Policies
			w.Header().Set("Cross-Origin-Embedder-Policy", config.CrossOriginEmbedderPolicy)
			w.Header().Set("Cross-Origin-Opener-Policy", config.CrossOriginOpenerPolicy)
			w.Header().Set("Cross-Origin-Resource-Policy", config.CrossOriginResourcePolicy)

			// Custom Headers
			for header, value := range config.CustomHeaders {
				w.Header().Set(header, value)
			}

			next.ServeHTTP(w, r)
		})
	}
}

// SecurityHeaders sử dụng cấu hình mặc định
func SecurityHeaders(next http.Handler) http.Handler {
	return SecurityHeadersWithConfig(DefaultSecurityConfig())(next)
}
