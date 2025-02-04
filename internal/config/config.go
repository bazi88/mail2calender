package config

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	API struct {
		Name       string `envconfig:"API_NAME" default:"go8"`
		Host       string `envconfig:"API_HOST" default:"0.0.0.0"`
		Port       int    `envconfig:"API_PORT" default:"8080"`
		RequestLog bool   `envconfig:"API_REQUEST_LOG" default:"false"`
		RunSwagger bool   `envconfig:"API_RUN_SWAGGER" default:"false"`
	}

	CORS struct {
		AllowedOrigins string `envconfig:"CORS_ALLOWED_ORIGINS" default:"http://localhost:3000"`
	}

	DB struct {
		Driver             string        `envconfig:"DB_DRIVER" default:"postgres"`
		Host               string        `envconfig:"DB_HOST" default:"postgres"`
		Port               int           `envconfig:"DB_PORT" default:"5432"`
		User               string        `envconfig:"DB_USER" default:"postgres"`
		Pass               string        `envconfig:"DB_PASS" default:"password"`
		Name               string        `envconfig:"DB_NAME" default:"go8_db"`
		SSLMode            string        `envconfig:"DB_SSL_MODE" default:"disable"`
		MaxConnectionPool  int           `envconfig:"DB_MAX_CONNECTION_POOL" default:"4"`
		MaxIdleConnections int           `envconfig:"DB_MAX_IDLE_CONNECTIONS" default:"4"`
		ConnectionLifetime time.Duration `envconfig:"DB_CONNECTIONS_MAX_LIFETIME" default:"300s"`
		TestName           string        `envconfig:"DB_TEST_NAME" default:"go8_e2e_db"`
	}

	Redis struct {
		Hosts     string        `envconfig:"REDIS_HOSTS"`
		Host      string        `envconfig:"REDIS_HOST" default:"redis"`
		Port      int           `envconfig:"REDIS_PORT" default:"6379"`
		Name      int           `envconfig:"REDIS_NAME" default:"0"`
		User      string        `envconfig:"REDIS_USER"`
		Pass      string        `envconfig:"REDIS_PASS"`
		CacheTime time.Duration `envconfig:"REDIS_CACHE_TIME" default:"5s"`
		Enable    bool          `envconfig:"REDIS_ENABLE" default:"true"`
	}

	Session struct {
		Name     string        `envconfig:"SESSION_NAME" default:"session"`
		Path     string        `envconfig:"SESSION_PATH" default:"/"`
		Domain   string        `envconfig:"SESSION_DOMAIN"`
		Duration time.Duration `envconfig:"SESSION_DURATION" default:"24h"`
		HTTPOnly bool          `envconfig:"SESSION_HTTP_ONLY" default:"true"`
		Secure   bool          `envconfig:"SESSION_SECURE" default:"true"`
	}

	OTEL struct {
		Enable         bool    `envconfig:"OTEL_ENABLE" default:"false"`
		OTLPEndpoint   string  `envconfig:"OTEL_OTLP_ENDPOINT" default:"otel-collector:4317"`
		ServiceName    string  `envconfig:"OTEL_OTLP_SERVICE_NAME" default:"go8"`
		ServiceVersion string  `envconfig:"OTEL_OTLP_SERVICE_VERSION" default:"0.1.0"`
		MeterName      string  `envconfig:"OTEL_OTLP_METER_NAME" default:"demo"`
		SamplerRatio   float64 `envconfig:"OTEL_OTLP_SAMPLER_RATIO" default:"0.1"`
	}

	JWT struct {
		Secret     string        `envconfig:"API_JWT_SECRET" required:"true"`
		Expiration time.Duration `envconfig:"API_JWT_EXPIRATION" default:"24h"`
	}

	Docker struct {
		Image string `envconfig:"DOCKER_IMAGE" default:"go8/server"`
		Tag   string `envconfig:"DOCKER_TAG" default:"latest"`
	}

	NER struct {
		Host string `envconfig:"NER_SERVICE_HOST" default:"ner-service"`
		Port int    `envconfig:"NER_SERVICE_PORT" default:"50051"`
	}
}

func Load() *Config {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	cfg := &Config{}
	if err := envconfig.Process("", cfg); err != nil {
		log.Fatal("Error processing environment variables:", err)
	}

	return cfg
}
