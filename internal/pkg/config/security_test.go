package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type SecurityConfig struct {
	JWT      JWTConfig
	Password PasswordConfig
}

type JWTConfig struct {
	Secret string
	TTL    int64
}

type PasswordConfig struct {
	MinLength      int
	RequireNumber  bool
	RequireSpecial bool
}

func TestSecurityConfig(t *testing.T) {
	t.Run("JWT Config", func(t *testing.T) {
		cfg := SecurityConfig{
			JWT: JWTConfig{
				Secret: "test-secret",
				TTL:    3600,
			},
		}

		assert.Equal(t, "test-secret", cfg.JWT.Secret)
		assert.Equal(t, int64(3600), cfg.JWT.TTL)
	})

	t.Run("Password Config", func(t *testing.T) {
		cfg := SecurityConfig{
			Password: PasswordConfig{
				MinLength:      8,
				RequireNumber:  true,
				RequireSpecial: true,
			},
		}

		assert.Equal(t, 8, cfg.Password.MinLength)
		assert.True(t, cfg.Password.RequireNumber)
		assert.True(t, cfg.Password.RequireSpecial)
	})
}
