package engine

const (
	CodeOK   = 0
	CodeFail = 1
)

// Response 仅用于 Swagger 文档生成
type Response struct {
	Code int    `json:"code" example:"0"`
	Msg  string `json:"msg" example:"ok"`
	Data any    `json:"data,omitempty"`
}
