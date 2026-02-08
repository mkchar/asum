package middleware

import (
	"asum/pkg/engine"
	"asum/pkg/errorx"
	"asum/pkg/models"
	"asum/pkg/rdb"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
)

type RequestPayload struct {
	ApiKey string `json:"apiKey"`
}

type LimitConfig struct {
	Max        int
	Expiration time.Duration
}

var levelRules = map[models.Level]LimitConfig{
	models.LevelBasic:   {Max: 1, Expiration: 1 * time.Second},
	models.LevelPlus:    {Max: 100, Expiration: 1 * time.Second},
	models.LevelPremium: {Max: 1000, Expiration: 1 * time.Second},
	models.LevelTop:     {Max: 5000, Expiration: 1 * time.Second},
}

func RateLimitAndAuthMiddleware(ctx context.Context, rdb *rdb.Client) fiber.Handler {
	return func(c fiber.Ctx) error {
		var payload RequestPayload
		if len(c.Body()) > 0 {
			if err := json.Unmarshal(c.Body(), &payload); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"code": engine.CodeFail,
					"msg":  errorx.ErrInvalidPayload.Error(),
				})
			}
		}

		currentLevel := models.LevelBasic
		limitConfig := levelRules[models.LevelBasic]
		limiterKey := "ratelimit:ip:" + c.IP()

		if payload.ApiKey != "" {
			redisKey := fmt.Sprintf("apiKey:%s", payload.ApiKey)
			levelStr, err := rdb.Get(ctx, redisKey).Result()
			if err == redis.Nil {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"code": engine.CodeFail,
					"msg":  errorx.ErrInvalidTaskKey.Error(),
				})
			} else if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"code": engine.CodeFail,
					"msg":  errorx.ErrRateLimitServeice.Error(),
				})
			}

			apiCache := models.ApiCache{}
			err = json.Unmarshal([]byte(levelStr), &apiCache)
			if err != nil {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"code": engine.CodeFail,
					"msg":  errorx.ErrInvalidTaskKey.Error(),
				})
			}
			userLevel := apiCache.UserLevel
			if rule, ok := levelRules[userLevel]; ok {
				currentLevel = userLevel
				limitConfig = rule
				limiterKey = "ratelimit:apikey:" + payload.ApiKey
			} else {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"code": engine.CodeFail,
					"msg":  errorx.ErrRateLimitServeice.Error(),
				})
			}
		}

		count, err := rdb.Incr(ctx, limiterKey).Result()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": engine.CodeFail,
				"msg":  errorx.ErrRateLimitServeice.Error(),
			})
		}

		if count == 1 {
			rdb.Expire(ctx, limiterKey, limitConfig.Expiration)
		}

		if count > int64(limitConfig.Max) {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"code": engine.CodeFail,
				"msg":  errorx.ErrRateLimited.Error(),
			})
		}
		c.Locals("userLevel", currentLevel)
		c.Locals("apiKey", payload.ApiKey)

		return c.Next()
	}
}

type RateLimit struct {
	Err   string `json:"err"`
	Level string `json:"level"`
}
