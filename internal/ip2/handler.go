package ip2

import (
	"asum/pkg/engine"
	"asum/pkg/errorx"

	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	service Service
}

func NewHandler(ip2Svc Service) *Handler {
	return &Handler{
		service: ip2Svc,
	}
}

// GetIP 查询单个 IP 信息
// @Summary 查询单个 IP 信息
// @Description 获取指定 IP 的地理位置、ASN、运营商等详细信息
// @Tags IP
// @Accept json
// @Produce json
// @Param ip path string true "IP 地址 (例如: 1.1.1.1)"
// @Param lang query string false "语言代码 (默认: en)" Enums(en, zh-CN) default(en)
// @Success 200 {object} engine.Response{data=GetIP} "查询成功"
// @Failure 400 {object} engine.Response "参数错误"
// @Failure 500 {object} engine.Response "服务器内部错误"
// @Router /ip/{ip} [get]
func (h *Handler) GetIP(c *engine.Ctx) error {
	ip := c.Params("ip")
	lang := c.Query("lang", "en")
	data, err := h.service.GetIP(c.StdCtx, ip, lang)
	if err != nil {
		return c.Fail(fiber.StatusForbidden, err.Error())
	}
	return c.OK(data)
}

type BatchIps struct {
	IPs    []string `json:"ips"`
	Lang   string   `json:"lang"`
	ApiKey string   `json:"apiKey"`
}

// BatchIP 批量查询 IP 信息
// @Summary 批量查询 IP 信息
// @Description 一次性查询多个 IP 的详细信息
// @Tags IP
// @Accept json
// @Produce json
// @Param request body BatchIps true "批量查询参数"
// @Success 200 {object} engine.Response{data=BatchIPResp} "查询成功"
// @Failure 400 {object} engine.Response "参数错误"
// @Router /ip/batch [post]
func (h *Handler) BatchIP(c *engine.Ctx) error {
	var req BatchIps
	if err := c.Bind().Body(&req); err != nil {
		return c.Fail(fiber.StatusBadRequest, errorx.ErrInvalidRequestBody)
	}

	if req.ApiKey == "" {
		return c.Fail(fiber.StatusBadRequest, errorx.ErrInvalidRequestBody)
	}

	if req.Lang == "" {
		req.Lang = "en"
	}

	data, err := h.service.BatchIP(c.StdCtx, req.IPs, req.ApiKey, req.Lang)
	if err != nil {
		return c.Fail(fiber.StatusForbidden, err.Error())
	}
	return c.OK(data)
}
