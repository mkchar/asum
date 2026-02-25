package middleware

import (
	"strings"

	"asum/pkg/engine"
	"asum/pkg/errorx"
	"asum/pkg/token"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

func Auth(secret string) fiber.Handler {
	return func(c fiber.Ctx) error {
		tokenString := ""

		authHeader := c.Get("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && parts[0] == "Bearer" && parts[1] != "" {
				tokenString = parts[1]
			}
		}

		if tokenString == "" {
			tokenString = c.Query("token")
			if tokenString == "" {
				tokenString = c.Query("access_token")
			}
		}

		if tokenString == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"code": engine.CodeFail,
				"msg":  errorx.ErrUnauthorized.Error(),
			})
		}

		claims := &token.Claims{}
		tk, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errorx.ErrUnauthorized
			}
			return []byte(secret), nil
		})

		if err != nil || !tk.Valid {
			if err != nil && strings.Contains(err.Error(), "expired") {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"code": engine.CodeFail,
					"msg":  errorx.ErrTokenExpired.Error(),
				})
			}
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"code": engine.CodeFail,
				"msg":  errorx.ErrUnauthorized.Error(),
			})
		}

		c.Locals("userID", claims.UserID)
		c.Locals("email", claims.Email)
		c.Locals("userLevel", claims.Level)

		return c.Next()
	}
}
