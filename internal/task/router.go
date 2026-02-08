package task

import (
	"asum/pkg/engine"

	"github.com/gofiber/fiber/v3"
)

func RegisterRoutes(r fiber.Router, h *Handler) {

	task := r.Group("/task")
	{
		task.Get("/", engine.H(h.ListTask))
		task.Post("/", engine.H(h.CreateTask))
		task.Patch("/:id", engine.H(h.UpdateTask))
		task.Delete("/:id", engine.H(h.DeleteTask))
		task.Get("/:id", engine.H(h.GetTask))
	}
}
