package auth

import (
	"database/sql"
	"time"

	"github.com/arlinggacr/BtechDevCases/internal/module/auth/controller"
	"github.com/arlinggacr/BtechDevCases/internal/module/auth/repository"
	"github.com/arlinggacr/BtechDevCases/internal/module/auth/service"
	"github.com/gofiber/fiber/v2"
)

type Module struct {
	controller *controller.Controller
}

func NewModule(db *sql.DB, jwtSecret string, tokenTTL time.Duration) *Module {
	repository.ConfigureAuthRepository(db)
	service.ConfigureAuthService(jwtSecret, tokenTTL)
	return &Module{
		controller: controller.NewController(),
	}
}

func (m *Module) RegisterRoutes(app *fiber.App) {
	auth := app.Group("/api/auth")
	auth.Post("/register", m.controller.RegisterAccount)
	auth.Post("/login", m.controller.LoginAccount)
	auth.Get("/me", service.AuthenticateRequest, m.controller.Me)
}
