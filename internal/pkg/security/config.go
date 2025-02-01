package security

import (
	"fmt"
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

// BuildCSP constructs the Content-Security-Policy header value
func BuildCSP(directives map[string][]string) string {
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

// BuildFeaturePolicy constructs the Feature-Policy header value
func BuildFeaturePolicy(policies map[string]string) string {
	if len(policies) == 0 {
		return ""
	}

	var featurePolicies []string
	for feature, value := range policies {
		featurePolicies = append(featurePolicies, fmt.Sprintf("%s=%s", feature, value))
	}
	return strings.Join(featurePolicies, "; ")
}
