package config

type Config struct {
	Server struct {
		Port int
	}
}

func Load() *Config {
	cfg := &Config{}
	cfg.Server.Port = 8080
	return cfg
}
