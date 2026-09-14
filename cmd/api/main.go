package main

import (
	"log"

	"github.com/arlinggacr/BtechDevCases/internal/config"
	"github.com/arlinggacr/BtechDevCases/internal/database"
	"github.com/arlinggacr/BtechDevCases/internal/module/auth"
	"github.com/arlinggacr/BtechDevCases/internal/server"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf(".env not loaded: %v", err)
	}
	cfg := config.Load()
	db, err := database.Connect(cfg.Database)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err := database.InitializeSchema(db); err != nil {
		log.Fatal(err)
	}

	authModule := auth.NewModule(db, cfg.JWTSecret, cfg.TokenTTL)
	app := server.New(authModule)

	log.Printf("server listening on %s", cfg.Address)
	if err := app.Listen(cfg.Address); err != nil {
		log.Fatal(err)
	}
}
