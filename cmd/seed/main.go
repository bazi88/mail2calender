package main

import (
	"fmt"
	"mono-golang/config"
	"mono-golang/database"
	db "mono-golang/third_party/database"
)

func main() {
	cfg := config.New()
	store := db.NewSqlx(cfg.Database)

	seeder := database.Seeder(store.DB)
	seeder.SeedUsers()
	fmt.Println("seeding completed.")
}
