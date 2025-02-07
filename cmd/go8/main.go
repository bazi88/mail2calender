package main

import (
	"mail2calendar/internal/config"
	"mail2calendar/internal/infrastructure/logger"
	"mail2calendar/internal/server"

	"github.com/sirupsen/logrus"
)

// Version is injected using ldflags during build time
var Version = "v0.1.0"

// @title Go8
// @version 0.1.0
// @description Go + Postgres + Chi router + sqlx + ent + Testing starter kit for API development.
// @contact.name User Name
// @contact.url https://github.com/gmhafiz/go8
// @contact.email email@example.com
// @host localhost:3080
// @BasePath /
func main() {
	// Load config
	cfg := config.Load()

	// Initialize logger
	log := logger.GetLogger()
	if cfg.API.RequestLog {
		log.SetLevel(logrus.DebugLevel)
	}
	log.Info("Starting application...")

	s := server.New(server.WithVersion(Version))
	s.Init()
	s.Run()
}
