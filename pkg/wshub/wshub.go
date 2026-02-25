package wshub

import (
	"encoding/json"
	"sync"

	"github.com/gofiber/contrib/v3/websocket"
)

type UID = uint64

type Hub struct {
	mu    sync.RWMutex
	conns map[UID]map[*websocket.Conn]struct{}
}

func New() *Hub {
	return &Hub{conns: make(map[UID]map[*websocket.Conn]struct{})}
}

func (h *Hub) Add(uid UID, c *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.conns[uid] == nil {
		h.conns[uid] = make(map[*websocket.Conn]struct{})
	}
	h.conns[uid][c] = struct{}{}
}

func (h *Hub) Del(uid UID, c *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if m := h.conns[uid]; m != nil {
		delete(m, c)
		if len(m) == 0 {
			delete(h.conns, uid)
		}
	}
}

func (h *Hub) Online(uid UID) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.conns[uid]) > 0
}

func (h *Hub) OnlineUIDs() []UID {
	h.mu.RLock()
	defer h.mu.RUnlock()
	uids := make([]UID, 0, len(h.conns))
	for uid := range h.conns {
		uids = append(uids, uid)
	}
	return uids
}

func (h *Hub) PushToUID(uid UID, payload []byte) {
	h.mu.RLock()
	m := h.conns[uid]
	h.mu.RUnlock()
	for c := range m {
		_ = c.WriteMessage(websocket.TextMessage, payload)
	}
}

func (h *Hub) Broadcast(payload []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, m := range h.conns {
		for c := range m {
			_ = c.WriteMessage(websocket.TextMessage, payload)
		}
	}
}

func MustJSON(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}
