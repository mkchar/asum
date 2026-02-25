package notify

type Out struct {
	Type   string      `json:"type"`             // "notify" | "badge"
	Data   interface{} `json:"data,omitempty"`   // 通知内容
	Unread int64       `json:"unread,omitempty"` // 红点数
	Ts     int64       `json:"ts"`
}
