package middleware

import (
	"asum/pkg/logx"
	"bytes"
	"encoding/json"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gofiber/fiber/v3"
)

func AccessLog() fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()
		reqBody := safeTrimBody(c.Body())
		err := c.Next()
		latency := time.Since(start)
		status := c.Response().StatusCode()
		respContentType := string(c.Response().Header.ContentType())
		var respBody string
		if isLoggableContentType(respContentType) {
			respBody = safeTrimBody(c.Response().Body())
		} else {
			respBody = "[BINARY/STREAM DATA]"
		}
		keysAndValues := []interface{}{
			"status", status,
			"method", c.Method(),
			"path", c.Path(),
			"ip", c.IP(),
			"latency", latency.String(),
			"request_id", c.RequestID(),
			"req_body", reqBody,
			"resp_body", respBody,
		}
		if err != nil {
			keysAndValues = append(keysAndValues, "error", err.Error())
		}
		if status >= 500 || err != nil {
			logx.Errorw("access_error", keysAndValues...)
		} else if status >= 400 {
			logx.Infow("access_warn", keysAndValues...)
		} else {
			logx.Infow("access", keysAndValues...)
		}
		return err
	}
}

const MaxLogBodySize = 1024

func safeTrimBody(body []byte) string {
	if len(body) == 0 {
		return ""
	}

	if !utf8.Valid(body) {
		return "[NON-UTF8 DATA]"
	}

	var str string
	dst := &bytes.Buffer{}
	if err := json.Compact(dst, body); err == nil {
		str = dst.String()
	} else {
		str = string(body)
		str = strings.ReplaceAll(str, "\n", "")
		str = strings.ReplaceAll(str, "\r", "")
		str = strings.Join(strings.Fields(str), " ")
	}

	if len(str) > MaxLogBodySize {
		return str[:MaxLogBodySize] + "...(truncated)"
	}
	return str
}

func isLoggableContentType(contentType string) bool {
	if contentType == "" {
		return true
	}
	return strings.Contains(contentType, "application/json") ||
		strings.Contains(contentType, "text/") ||
		strings.Contains(contentType, "application/xml")
}
