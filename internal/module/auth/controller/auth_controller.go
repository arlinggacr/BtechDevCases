package controller

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"github.com/arlinggacr/BtechDevCases/internal/module/auth/model"
	"github.com/arlinggacr/BtechDevCases/internal/module/auth/service"
)

type Controller struct{}

func NewController() *Controller { return &Controller{} }

func (h *Controller) RegisterAccount(c *fiber.Ctx) error {
	var request model.RegisterRequest
	if err := c.BodyParser(&request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	user, err := service.RegisterAccount(request)
	if err != nil {
		if errors.Is(err, service.ErrEmailExists) {
			return fiber.NewError(fiber.StatusConflict, err.Error())
		}
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"user": user})
}

func (h *Controller) LoginAccount(c *fiber.Ctx) error {
	var request model.LoginRequest
	if err := c.BodyParser(&request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	response, err := service.LoginAccount(request)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, service.ErrInvalidCredentials.Error())
	}
	return c.JSON(response)
}

func (h *Controller) Me(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Hello " + c.Locals("email").(string) + ", welcome back"})
}
