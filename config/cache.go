package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/kelseyhightower/envconfig"
)

type Cache struct {
	Enable    bool   `default:"false"`
	Host      string `default:"0.0.0.0"`
	Port      string `default:"6379"`
	Hosts     []string
	Name      int `default:"1"`
	User      string
	Pass      string
	CacheTime time.Duration `split_words:"true" default:"5s"`
}

func NewCache() Cache {
	var cache Cache
	envconfig.MustProcess("REDIS", &cache)

	if strings.Contains(cache.Host, ",") {
		cache.Hosts = strings.Split(cache.Host, ",")
	}

	return cache
}

// NewRedisClient tạo một Redis client mới
func (c *Cache) NewRedisClient() (*redis.Client, error) {
	if !c.Enable {
		return nil, fmt.Errorf("redis is not enabled")
	}

	options := &redis.Options{
		Addr:     fmt.Sprintf("%s:%s", c.Host, c.Port),
		Password: c.Pass,
		DB:       c.Name,
	}

	if c.User != "" {
		options.Username = c.User
	}

	client := redis.NewClient(options)
	return client, nil
}
