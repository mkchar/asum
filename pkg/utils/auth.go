package utils

import "github.com/gofiber/fiber/v3"

func GetUserID(c fiber.Ctx) uint64 {
	if uid, ok := c.Locals("userID").(uint64); ok {
		return uid
	}
	return 0
}

func GetUserLevel(c fiber.Ctx) int {
	if level, ok := c.Locals("userLevel").(int); ok {
		return level
	}
	return 0
}
