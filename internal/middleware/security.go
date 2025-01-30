package middleware

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

// SecurityConfig holds configuration for security headers
type SecurityConfig struct {
	HSTSMaxAge            int
	HSTSIncludeSubdomains bool
	CSPDirectives         map[string][]string
	FrameOptions          string
	XContentTypeOptions   string
	ReferrerPolicy        string
	PermissionsPolicy     string
	FeaturePolicy         map[string]string
	CustomHeaders         map[string]string
}

// DefaultSecurityConfig returns the default security configuration
func DefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		HSTSMaxAge:            31536000, // 1 year
		HSTSIncludeSubdomains: true,
		CSPDirectives: map[string][]string{
			"default-src": {"'self'"},
			"script-src":  {"'self'"},
			"style-src":   {"'self'"},
		},
		FrameOptions:        "DENY",
		XContentTypeOptions: "nosniff",
		ReferrerPolicy:      "strict-origin-when-cross-origin",
		PermissionsPolicy:   "accelerometer=(), camera=(), geolocation=(), gyroscope=(), magnetometer=(), microphone=(), payment=(), usb=()",
		FeaturePolicy:       make(map[string]string),
		CustomHeaders: map[string]string{
			"X-XSS-Protection": "1; mode=block",
		},
	}
}

// ValidateConfig checks if the security configuration is valid
func ValidateConfig(config *SecurityConfig) error {
	if config == nil {
		return fmt.Errorf("security config cannot be nil")
	}
	if config.HSTSMaxAge < 0 {
		return fmt.Errorf("HSTS max age cannot be negative: %d", config.HSTSMaxAge)
	}
	return nil
}

// buildCSP constructs the Content-Security-Policy header value
func buildCSP(directives map[string][]string) string {
	if len(directives) == 0 {
		return ""
	}

	var policies []string
	for directive, sources := range directives {
		if len(sources) > 0 {
			policies = append(policies, fmt.Sprintf("%s %s", directive, strings.Join(sources, " ")))
		}
	}
	return strings.Join(policies, "; ")
}

// buildFeaturePolicy constructs the Feature-Policy header value
func buildFeaturePolicy(policies map[string]string) string {
	if len(policies) == 0 {
		return ""
	}

	var featurePolicies []string
	for feature, value := range policies {
		featurePolicies = append(featurePolicies, fmt.Sprintf("%s=%s", feature, value))
	}
	return strings.Join(featurePolicies, "; ")
}

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
func SecurityHeadersWithConfig(config *SecurityConfig) func(http.Handler) http.Handler {
	// If config is nil, use default configuration
	if config == nil {
		config = DefaultSecurityConfig()
	}

	// Validate config before creating middleware
	if err := ValidateConfig(config); err != nil {
		log.Printf("Security config validation failed: %v", err)
		config = DefaultSecurityConfig()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hstsValue := fmt.Sprintf("max-age=%d", config.HSTSMaxAge)
			if config.HSTSIncludeSubdomains {
				hstsValue += "; includeSubDomains"
			}

			headers := map[string]string{
				"Strict-Transport-Security": hstsValue,
				"Content-Security-Policy":   buildCSP(config.CSPDirectives),
				"X-Frame-Options":           config.FrameOptions,
				"X-Content-Type-Options":    config.XContentTypeOptions,
				"Referrer-Policy":           config.ReferrerPolicy,
				"Permissions-Policy":        config.PermissionsPolicy,
				"Feature-Policy":            buildFeaturePolicy(config.FeaturePolicy),
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
