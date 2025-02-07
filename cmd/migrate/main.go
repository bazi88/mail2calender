package main

import (
	"log"

	"mail2calendar/config"
	"mail2calendar/database"
	db "mail2calendar/third_party/database"
)

// Version is injected using ldflags during build time
var Version string

func main() {
	log.Printf("Version: %s\n", Version)

	cfg := config.New()
	store := db.NewSqlx(cfg.Database)
	migrator := database.Migrator(store.DB)

	// todo: accept cli flag for other operations
	migrator.Up()
}
