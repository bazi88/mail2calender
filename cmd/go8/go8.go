package main

import (
	"flag"

	"github.com/go-chi/httplog"

	"eight/internal/api"
	"eight/internal/configs"
	"eight/internal/datastore"
	"eight/internal/server/http"
	"eight/internal/service/authors"
	"eight/internal/service/books"
	"eight/pkg/redis"
)

const Version = "v0.1.0"

var flagConfig = flag.String("config", "./config/dev.yml", "path to the config file")

func main() {
	logger := httplog.NewLogger("go8", httplog.Options{
		JSON: true, // false
		Concise: true,
		Tags: map[string]string{"version": Version},
	})
	logger = logger.With().Caller().Logger()

	//cfg, err := configs.NewService("dev")
	cfg, err := configs.NewService(*flagConfig)
	if err != nil {
		logger.Error().Err(err)
		return
	}

	dataStoreCfg, err := cfg.DataStore()
	if err != nil {
		logger.Error().Err(err)
		return
	}

	db, err := datastore.NewService(dataStoreCfg)
	if err != nil {
		logger.Error().Err(err)
		return
	}

	cacheCfg, err := cfg.CacheStore()
	if err != nil {
		logger.Error().Err(err)
		return
	}

	rdb, err := redis.NewClient(cacheCfg)
	if err != nil {
		logger.Error().Err(err)
		return
	}

	bookService, err := books.NewService(db, rdb)
	if err != nil {
		logger.Error().Err(err)
		return
	}

	authorService, err := authors.NewService(db, rdb)
	if err != nil {
		logger.Error().Err(err)
		return
	}

	// additional microservice added here
	a, err := api.NewService(bookService, authorService)
	if err != nil {
		logger.Error().Err(err)
		return
	}

	httpCfg, err := cfg.HTTP()
	if err != nil {
		logger.Error().Err(err)
		return
	}

	h, err := http.NewService(httpCfg, a, logger, cfg.Time)
	if err != nil {
		logger.Error().Err(err)
		return
	}

	h.Start(logger)
}
