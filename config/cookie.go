package config

import (
	"net/http"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Session struct {
	Name     string          `envconfig:"SESSION_NAME" default:"session"`
	Path     string          `envconfig:"SESSION_PATH" default:"/"`
	Domain   string          `envconfig:"SESSION_DOMAIN"`
	Secret   string          `required:"false"`
	Duration time.Duration   `envconfig:"SESSION_DURATION" default:"24h"`
	HTTPOnly bool            `envconfig:"SESSION_HTTP_ONLY" default:"true"`
	Secure   bool            `envconfig:"SESSION_SECURE" default:"true"`
	SameSite SameSiteDecoder `split_words:"true" default:"lax"`
}

func NewSession() Session {
	var a Session
	envconfig.MustProcess("SESSION", &a)

	return a
}

type SameSiteDecoder http.SameSite

func (sd *SameSiteDecoder) Decode(value string) error {
	switch value {
	case "default":
		*sd = SameSiteDecoder(http.SameSiteDefaultMode)
	case "lax":
		*sd = SameSiteDecoder(http.SameSiteLaxMode)
	case "strict":
		*sd = SameSiteDecoder(http.SameSiteStrictMode)
	case "none":
		*sd = SameSiteDecoder(http.SameSiteNoneMode)
	default:
		*sd = SameSiteDecoder(http.SameSiteLaxMode)
	}

	return nil
}
