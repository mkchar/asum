package user

import "github.com/gofiber/fiber/v3"

type Handler struct {
	service Service
}

func NewHandler(userSvc Service) *Handler {
	return &Handler{service: userSvc}
}

func (h *Handler) CreateUser(c fiber.Ctx) error {
	// return
	return nil
}

func (h *Handler) UpdateUser(c fiber.Ctx) error {
	// return
	return nil
}

func (h *Handler) GetUser(c fiber.Ctx) error {
	// return
	return nil
}

func (h *Handler) DeleteUser(c fiber.Ctx) error {
	// return
	return nil
}

func (h *Handler) ListUser(c fiber.Ctx) error {
	return nil
}
