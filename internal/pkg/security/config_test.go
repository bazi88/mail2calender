package security

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultSecurityConfig(t *testing.T) {
	config := DefaultSecurityConfig()

	assert.NotNil(t, config)
	assert.Equal(t, 31536000, config.HSTSMaxAge)
	assert.True(t, config.HSTSIncludeSubdomains)
	assert.Equal(t, "DENY", config.FrameOptions)
	assert.Equal(t, "nosniff", config.XContentTypeOptions)
	assert.Equal(t, "strict-origin-when-cross-origin", config.ReferrerPolicy)
	assert.Equal(t, "accelerometer=(), camera=(), geolocation=(), gyroscope=(), magnetometer=(), microphone=(), payment=(), usb=()", config.PermissionsPolicy)
	assert.Equal(t, "1; mode=block", config.CustomHeaders["X-XSS-Protection"])

	// Check CSP directives
	assert.Contains(t, config.CSPDirectives, "default-src")
	assert.Contains(t, config.CSPDirectives, "script-src")
	assert.Contains(t, config.CSPDirectives, "style-src")
	assert.Equal(t, []string{"'self'"}, config.CSPDirectives["default-src"])
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name          string
		config        *SecurityConfig
		expectedError string
	}{
		{
			name:          "nil config",
			config:        nil,
			expectedError: "security config cannot be nil",
		},
		{
			name: "negative HSTS max age",
			config: &SecurityConfig{
				HSTSMaxAge: -1,
			},
			expectedError: "HSTS max age cannot be negative: -1",
		},
		{
			name:          "valid config",
			config:        DefaultSecurityConfig(),
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.config)
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBuildCSP(t *testing.T) {
	tests := []struct {
		name       string
		directives map[string][]string
		expected   []string
	}{
		{
			name:       "empty directives",
			directives: map[string][]string{},
			expected:   nil,
		},
		{
			name: "single directive",
			directives: map[string][]string{
				"default-src": {"'self'"},
			},
			expected: []string{"default-src 'self'"},
		},
		{
			name: "multiple directives",
			directives: map[string][]string{
				"default-src": {"'self'"},
				"script-src":  {"'self'", "'unsafe-inline'"},
				"style-src":   {"'self'", "https://fonts.googleapis.com"},
			},
			expected: []string{
				"default-src 'self'",
				"script-src 'self' 'unsafe-inline'",
				"style-src 'self' https://fonts.googleapis.com",
			},
		},
		{
			name: "empty sources",
			directives: map[string][]string{
				"default-src": {},
				"script-src":  {"'self'"},
			},
			expected: []string{"script-src 'self'"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildCSP(tt.directives)
			if tt.expected == nil {
				assert.Empty(t, result)
			} else {
				// Split result into individual directives and verify each one exists
				directives := strings.Split(result, "; ")
				assert.Equal(t, len(tt.expected), len(directives))
				for _, expectedDirective := range tt.expected {
					assert.Contains(t, directives, expectedDirective)
				}
			}
		})
	}
}

func TestBuildFeaturePolicy(t *testing.T) {
	tests := []struct {
		name     string
		policies map[string]string
		expected []string
	}{
		{
			name:     "empty policies",
			policies: map[string]string{},
			expected: nil,
		},
		{
			name: "single policy",
			policies: map[string]string{
				"camera": "none",
			},
			expected: []string{"camera=none"},
		},
		{
			name: "multiple policies",
			policies: map[string]string{
				"camera":      "none",
				"microphone":  "self",
				"geolocation": "*",
			},
			expected: []string{
				"camera=none",
				"microphone=self",
				"geolocation=*",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildFeaturePolicy(tt.policies)
			if tt.expected == nil {
				assert.Empty(t, result)
			} else {
				// Split result into individual policies and verify each one exists
				policies := strings.Split(result, "; ")
				assert.Equal(t, len(tt.expected), len(policies))
				for _, expectedPolicy := range tt.expected {
					assert.Contains(t, policies, expectedPolicy)
				}
			}
		})
	}
}
