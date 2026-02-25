package notify

import (
	"asum/pkg/rdb"
	"asum/pkg/wshub"
	"context"
	"encoding/json"
	"time"

	"github.com/gofiber/contrib/v3/websocket"
)

type In struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`
}

type WSHandler struct {
	hub     *wshub.Hub
	redisDB *rdb.Client
	baseCtx context.Context
}

func NewWSHandler(baseCtx context.Context, hub *wshub.Hub, redisDB *rdb.Client) *WSHandler {
	return &WSHandler{baseCtx: baseCtx, hub: hub, redisDB: redisDB}
}

func (h *WSHandler) Handle(c *websocket.Conn) {
	uidAny := c.Locals("uid")
	uid, _ := uidAny.(uint64)
	if uid == 0 {
		_ = c.Close()
		return
	}

	h.hub.Add(uid, c)
	defer func() {
		h.hub.Del(uid, c)
		_ = c.Close()
	}()

	n, _ := GetUnread(h.baseCtx, h.redisDB, uid)
	h.hub.PushToUID(uid, wshub.MustJSON(Out{
		Type:   "badge",
		Unread: n,
		Ts:     time.Now().Unix(),
	}))

	for {
		_, payload, err := c.ReadMessage()
		if err != nil {
			return
		}
		var in In
		if err := json.Unmarshal(payload, &in); err != nil {
			continue
		}
		switch in.Type {
		case "clear_badge":
			_ = ClearUnread(h.baseCtx, h.redisDB, uid)
			h.hub.PushToUID(uid, wshub.MustJSON(Out{
				Type:   "badge",
				Unread: 0,
				Ts:     time.Now().Unix(),
			}))
		}
	}
}
