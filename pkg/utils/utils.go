package utils

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"

	"github.com/google/uuid"
)

var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
)

func generateKey(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func NewUUID() string {
	u := uuid.New()
	return hex.EncodeToString(u[:])
}

type contextKey string

const (
	ContextKeyRequestID contextKey = "request_id"
)

func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ContextKeyRequestID, id)
}

func GetRequestID(ctx context.Context) string {
	val, ok := ctx.Value(ContextKeyRequestID).(string)
	if !ok {
		return ""
	}
	return val
}

func WithUserAgent(ctx context.Context, agent string) context.Context {
	return context.WithValue(ctx, "user_agent", agent)
}

func GetUserAgent(ctx context.Context) string {
	val, ok := ctx.Value("user_agent").(string)
	if !ok {
		return ""
	}
	return val
}

func WithRemoteIP(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, "remote_ip", id)
}

func GetRemoteIP(ctx context.Context) string {
	val, ok := ctx.Value("remote_ip").(string)
	if !ok {
		return ""
	}
	return val
}
func StringIndex(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func StringsContains(s, sub string) bool {
	return len(sub) > 0 && len(s) >= len(sub) && (func() bool { return (StringIndex(s, sub) >= 0) })()
}
