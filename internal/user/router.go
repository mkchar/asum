package user

import (
	"github.com/gofiber/fiber/v3"
)

func RegisterRoutes(r fiber.Router, h *Handler) {
	user := r.Group("/user")
	{
		user.Get("/", h.ListUser)
		user.Post("/", h.CreateUser)
		user.Patch("/:id", h.UpdateUser)
		user.Delete("/:id", h.DeleteUser)
		user.Get("/:id", h.GetUser)
	}
}
