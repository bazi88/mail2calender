package config

type Config struct {
	Security SecurityConfig `mapstructure:"security"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
}

type SecurityConfig struct {
	JWT      JWTConfig      `mapstructure:"jwt"`
	Password PasswordConfig `mapstructure:"password"`
}

type JWTConfig struct {
	Secret          string `mapstructure:"secret"`
	ExpirationHours int    `mapstructure:"expiration_hours"`
}

type PasswordConfig struct {
	Salt      string `mapstructure:"salt"`
	MinLength int    `mapstructure:"min_length"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}
