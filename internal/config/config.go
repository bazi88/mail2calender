package config

type Config struct {
	Server struct {
		Port int
	}
	Logger struct {
		Level string // debug, info, warn, error
	}
}

func Load() *Config {
	cfg := &Config{}
	cfg.Server.Port = 8080
	cfg.Logger.Level = "info" // default log level
	return cfg
}
