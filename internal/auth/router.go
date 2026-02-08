package auth

import (
	"asum/pkg/engine"

	"github.com/gofiber/fiber/v3"
)

func RegisterRoutes(r fiber.Router, h *Handler) {
	r.Post("/login", engine.H(h.Login))
	r.Post("/register", engine.H(h.Register))
	r.Post("/verify", engine.H(h.Verify))
	r.Post("/confirm/:code", engine.H(h.ConfirmCode))
	r.Get("/confirm", engine.H(h.ConfirmURL))

	r.Post("/refresh", engine.H(h.RefreshToken))
	r.Post("/reset-password", engine.H(h.ResetPassword))
	r.Post("/reset-password/confirm", engine.H(h.ResetPasswordConfirm))
}
