package main

import (
	"fmt"

	"mail2calendar/config"
	"mail2calendar/database"
	db "mail2calendar/third_party/database"
)

func main() {
	cfg := config.New()
	store := db.NewSqlx(cfg.Database)

	seeder := database.Seeder(store.DB)
	seeder.SeedUsers()
	fmt.Println("seeding completed.")
}
