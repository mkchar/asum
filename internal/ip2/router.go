package ip2

import (
	"asum/pkg/engine"

	"github.com/gofiber/fiber/v3"
)

func RegisterRoutes(r fiber.Router, h *Handler) {
	r.Get("/:ip", engine.H(h.GetIP))      // GET 查询单个IP
	r.Post("/batch", engine.H(h.BatchIP)) // POST 批量查询IP
}
