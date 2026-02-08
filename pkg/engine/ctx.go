package engine

import (
	"asum/pkg/utils"
	"context"
	"fmt"

	"github.com/gofiber/fiber/v3"
)

func (c *Ctx) OK(data any) error {
	return c.JSON(Response{
		Code: CodeOK,
		Msg:  "ok",
		Data: data,
	})
}

func (c *Ctx) Fail(code int, msg any) error {
	var message string
	switch v := msg.(type) {
	case string:
		message = v
	case error:
		message = v.Error()
	default:
		message = fmt.Sprintf("%v", v)
	}
	return c.Status(code).JSON(Response{
		Code: CodeFail,
		Msg:  message,
		Data: nil,
	})
}

type Ctx struct {
	fiber.Ctx
	StdCtx context.Context
}

type HandlerFunc func(c *Ctx) error

func H(h HandlerFunc) fiber.Handler {
	return func(c fiber.Ctx) error {
		stdCtx := c.Context()
		var rid string
		if v, ok := c.Locals("requestid").(string); ok {
			rid = v
		} else {
			rid = c.Get("X-Request-ID")
		}

		if rid != "" {
			stdCtx = utils.WithRequestID(stdCtx, rid)
		}

		stdCtx = utils.WithRemoteIP(stdCtx, c.IP())
		stdCtx = utils.WithUserAgent(stdCtx, c.UserAgent())
		customCtx := &Ctx{
			Ctx:    c,
			StdCtx: stdCtx,
		}

		return h(customCtx)
	}
}

// func (c *Ctx) Context() context.Context {
// 	return c.StdCtx
// }
