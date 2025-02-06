package config

import (
	"time"
)

// WebConfig contains configuration for the web interface
type WebConfig struct {
	// Server settings
	Port            int           `env:"WEB_PORT" envDefault:"8080"`
	Host            string        `env:"WEB_HOST" envDefault:"0.0.0.0"`
	ReadTimeout     time.Duration `env:"WEB_READ_TIMEOUT" envDefault:"15s"`
	WriteTimeout    time.Duration `env:"WEB_WRITE_TIMEOUT" envDefault:"15s"`
	ShutdownTimeout time.Duration `env:"WEB_SHUTDOWN_TIMEOUT" envDefault:"30s"`

	// Session settings
	SessionSecret   string        `env:"SESSION_SECRET,required"`
	SessionDuration time.Duration `env:"SESSION_DURATION" envDefault:"24h"`

	// OAuth settings
	GoogleClientID      string `env:"GOOGLE_CLIENT_ID,required"`
	GoogleClientSecret  string `env:"GOOGLE_CLIENT_SECRET,required"`
	OutlookClientID     string `env:"OUTLOOK_CLIENT_ID,required"`
	OutlookClientSecret string `env:"OUTLOOK_CLIENT_SECRET,required"`
	OAuthRedirectURL    string `env:"OAUTH_REDIRECT_URL,required"`

	// UI settings
	TemplatesDir string `env:"TEMPLATES_DIR" envDefault:"web/templates"`
	StaticDir    string `env:"STATIC_DIR" envDefault:"web/static"`
	AssetsDir    string `env:"ASSETS_DIR" envDefault:"web/assets"`
}
