package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSecurityConfig(t *testing.T) {
	t.Run("JWT Config", func(t *testing.T) {
		cfg := &SecurityConfig{
			JWT: JWTConfig{
				Secret:          "test-secret",
				ExpirationHours: 24,
			},
		}
		assert.Equal(t, "test-secret", cfg.JWT.Secret)
		assert.Equal(t, 24, cfg.JWT.ExpirationHours)
	})

	t.Run("Password Config", func(t *testing.T) {
		cfg := &SecurityConfig{
			Password: PasswordConfig{
				Salt:      "test-salt",
				MinLength: 8,
			},
		}
		assert.Equal(t, "test-salt", cfg.Password.Salt)
		assert.Equal(t, 8, cfg.Password.MinLength)
	})
}

func TestDatabaseConfig(t *testing.T) {
	cfg := &DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "test-user",
		Password: "test-pass",
		DBName:   "test-db",
		SSLMode:  "disable",
	}

	assert.Equal(t, "localhost", cfg.Host)
	assert.Equal(t, 5432, cfg.Port)
	assert.Equal(t, "test-user", cfg.User)
	assert.Equal(t, "test-pass", cfg.Password)
	assert.Equal(t, "test-db", cfg.DBName)
	assert.Equal(t, "disable", cfg.SSLMode)
}

func TestRedisConfig(t *testing.T) {
	cfg := &RedisConfig{
		Host:     "localhost",
		Port:     6379,
		Password: "test-pass",
		DB:       0,
	}

	assert.Equal(t, "localhost", cfg.Host)
	assert.Equal(t, 6379, cfg.Port)
	assert.Equal(t, "test-pass", cfg.Password)
	assert.Equal(t, 0, cfg.DB)
}
