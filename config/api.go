// Package config cung cấp các cấu hình cho ứng dụng
package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

// API chứa cấu hình cho API server
type API struct {
	Name              string        `default:"go8_api"`
	Host              string        `default:"0.0.0.0"`
	Port              string        `default:"3080"`
	ReadHeaderTimeout time.Duration `split_words:"true" default:"60s"`
	GracefulTimeout   time.Duration `split_words:"true" default:"8s"`

	RequestLog bool `split_words:"true" default:"false"`
	RunSwagger bool `split_words:"true" default:"true"`
}

// API trả về cấu hình API mặc định
func API() API {
	var api API
	envconfig.MustProcess("API", &api)

	return api
}
