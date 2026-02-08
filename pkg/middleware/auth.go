package middleware

import (
	"asum/pkg/engine"
	"asum/pkg/errorx"
	"asum/pkg/token"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

func Auth(secret string) fiber.Handler {
	return func(c fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"code": engine.CodeFail,
				"msg":  errorx.ErrUnauthorized.Error(),
			})
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"code": engine.CodeFail,
				"msg":  errorx.ErrUnauthorized.Error(),
			})
		}
		tokenString := parts[1]

		claims := &token.Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errorx.ErrUnauthorized
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			if strings.Contains(err.Error(), "expired") {
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

		// c = utils.WithRequestID(stdCtx, rid)
		c.Locals("userID", claims.UserID)
		c.Locals("email", claims.Email)
		c.Locals("userLevel", claims.Level)

		return c.Next()
	}
}
