package main

import (
	"log"

	"github.com/arlinggacr/BtechDevCases/internal/config"
	"github.com/arlinggacr/BtechDevCases/internal/module/auth"
	"github.com/arlinggacr/BtechDevCases/internal/server"
)

func main() {
	cfg := config.Load()
	authModule := auth.NewModule(cfg.JWTSecret, cfg.TokenTTL)
	app := server.New(authModule)

	log.Printf("server listening on %s", cfg.Address)
	if err := app.Listen(cfg.Address); err != nil {
		log.Fatal(err)
	}
}
