package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSecurityConfig(t *testing.T) {
	t.Run("JWT Config", func(t *testing.T) {
		cfg := SecurityConfig{
			JWT: JWTConfig{
				Secret:          "test-secret",
				ExpirationHours: 1,
			},
			Password: PasswordConfig{
				Salt:      "test-salt",
				MinLength: 8,
			},
		}

		assert.Equal(t, "test-secret", cfg.JWT.Secret)
		assert.Equal(t, 1, cfg.JWT.ExpirationHours)
		assert.Equal(t, "test-salt", cfg.Password.Salt)
		assert.Equal(t, 8, cfg.Password.MinLength)
	})
}
