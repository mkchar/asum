package auth

import (
	"asum/pkg/engine"
	"asum/pkg/errorx"

	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	service Service
}

func NewHandler(authSvc Service) *Handler {
	return &Handler{
		service: authSvc,
	}
}

type LoginReq struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}
type LoginResp struct {
	Token        string    `json:"token"`
	RefreshToken string    `json:"refreshToken,omitempty"`
	User         *UserInfo `json:"user"`
}

// Login 用户登录
// @Summary 用户登录
// @Description 使用邮箱和密码登录
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body LoginReq true "登录参数"
// @Success 200 {object} engine.Response{data=LoginResp}
// @Failure 400 {object} engine.Response
// @Failure 401 {object} engine.Response
// @Router /auth/login [post]
func (h *Handler) Login(c *engine.Ctx) error {
	var req LoginReq
	if err := c.Bind().Body(&req); err != nil {
		return c.Fail(fiber.StatusBadRequest, errorx.ErrInvalidRequestBody)
	}

	data, err := h.service.Login(c.StdCtx, &req)
	if err != nil {
		return c.Fail(fiber.StatusForbidden, err.Error())
	}

	return c.OK(data)
}

type RegisterReq struct {
	Name     string `json:"name" validate:"required,min=2,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}
type RegisterResp struct {
	Message string `json:"message"`
	UserID  uint64 `json:"userId,omitempty"`
}

// Register 用户注册
// @Summary 用户注册
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body RegisterReq true "注册参数"
// @Success 200 {object} engine.Response{data=RegisterResp}
// @Failure 400 {object} engine.Response
// @Router /auth/register [post]
func (h *Handler) Register(c *engine.Ctx) error {
	var req RegisterReq
	if err := c.Bind().Body(&req); err != nil {
		return c.Fail(fiber.StatusBadRequest, errorx.ErrInvalidRequestBody)
	}

	data, err := h.service.Register(c.StdCtx, &req)
	if err != nil {
		return c.Fail(fiber.StatusForbidden, err.Error())
	}

	return c.OK(data)
}

type VerifyReq struct {
	Email string `json:"email" validate:"required,email"`
}
type VerifyResp struct {
	Message   string `json:"message"`
	ExpiresIn int    `json:"expiresIn"`
}

// Verify 发送验证邮件
// @Summary 发送验证邮件
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body VerifyReq true "验证参数"
// @Success 200 {object} engine.Response{data=VerifyResp}
// @Failure 400 {object} engine.Response
// @Router /auth/verify [post]
func (h *Handler) Verify(c *engine.Ctx) error {
	var req VerifyReq
	if err := c.Bind().Body(&req); err != nil {
		return c.Fail(fiber.StatusBadRequest, errorx.ErrInvalidRequestBody)
	}

	data, err := h.service.Verify(c, &req)
	if err != nil {
		return c.Fail(fiber.StatusForbidden, err.Error())
	}

	return c.OK(data)
}

type ConfirmCodeReq struct {
	Email string `json:"email" validate:"required,email"`
}
type ConfirmResp struct {
	Message string    `json:"message"`
	User    *UserInfo `json:"user,omitempty"`
}

// ConfirmCode 验证码确认
// @Summary 验证码确认
// @Description 需同时提供URL路径中的code和Body中的email
// @Tags Auth
// @Accept json
// @Produce json
// @Param code path string true "验证码"
// @Param request body ConfirmCodeReq true "确认参数"
// @Success 200 {object} engine.Response{data=ConfirmResp}
// @Failure 400 {object} engine.Response
// @Router /auth/confirm/{code} [post]
func (h *Handler) ConfirmCode(c *engine.Ctx) error {
	code := c.Params("code")
	var req ConfirmCodeReq
	if err := c.Bind().Body(&req); err != nil {
		return c.Fail(fiber.StatusBadRequest, errorx.ErrInvalidRequestBody)
	}

	data, err := h.service.ConfirmCode(c.StdCtx, req.Email, code)
	if err != nil {
		return c.Fail(fiber.StatusForbidden, err.Error())
	}

	return c.OK(data)
}

type ConfirmReq struct {
	Token string `json:"token" validate:"required"`
}

// ConfirmURL 注册链接确认
// @Summary 链接确认
// @Tags Auth
// @Accept json
// @Produce json
// @Param token query string true "Token参数"
// @Success 200 {object} engine.Response
// @Failure 400 {object} engine.Response
// @Router /auth/confirm [get]
func (h *Handler) ConfirmURL(c *engine.Ctx) error {
	token := c.Query("token")
	if token == "" {
		return c.Fail(400, "token is required")
	}

	data, err := h.service.ConfirmURL(c.StdCtx, &ConfirmReq{Token: token})
	if err != nil {
		return c.Fail(fiber.StatusForbidden, err.Error())
	}

	return c.OK(data)
}

type RefreshReq struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}

type RefreshResp struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refreshToken"`
}
type UserInfo struct {
	ID    uint64 `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Level int    `json:"level"`
}

// RefreshToken 刷新 Token
// @Summary 刷新 Access Token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body RefreshReq true "刷新参数"
// @Success 200 {object} engine.Response{data=RefreshResp}
// @Failure 400 {object} engine.Response
// @Failure 401 {object} engine.Response
// @Router /auth/refresh [post]
func (h *Handler) RefreshToken(c *engine.Ctx) error {
	var req RefreshReq
	if err := c.Bind().Body(&req); err != nil {
		return c.Fail(fiber.StatusBadRequest, errorx.ErrInvalidRequestBody)
	}

	data, err := h.service.RefreshToken(c, &req)
	if err != nil {
		return c.Fail(fiber.StatusForbidden, err.Error())
	}

	return c.OK(data)
}

type ResetPasswordReq struct {
	Email string `json:"email" validate:"required,email"`
}

// ResetPassword 重置密码请求
// @Summary 请求重置密码
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body ResetPasswordReq true "邮箱参数"
// @Success 200 {object} engine.Response{data=VerifyResp}
// @Failure 400 {object} engine.Response
// @Router /auth/reset-password [post]
func (h *Handler) ResetPassword(c *engine.Ctx) error {
	var req ResetPasswordReq
	if err := c.Bind().Body(&req); err != nil {
		return c.Fail(fiber.StatusBadRequest, errorx.ErrInvalidRequestBody)
	}

	data, err := h.service.ResetPassword(c, &req)
	if err != nil {
		return c.Fail(fiber.StatusForbidden, err.Error())
	}

	return c.OK(data)
}

type ResetPasswordConfirmReq struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"newPassword" validate:"required,min=8"`
}

// ResetPasswordConfirm 确认重置密码
// @Summary 确认重置密码
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body ResetPasswordConfirmReq true "重置参数"
// @Success 200 {object} engine.Response{data=ConfirmResp}
// @Failure 400 {object} engine.Response
// @Router /auth/reset-password/confirm [post]
func (h *Handler) ResetPasswordConfirm(c *engine.Ctx) error {
	var req ResetPasswordConfirmReq
	if err := c.Bind().Body(&req); err != nil {
		return c.Fail(fiber.StatusBadRequest, errorx.ErrInvalidRequestBody)
	}

	data, err := h.service.ResetPasswordConfirm(c.StdCtx, &req)
	if err != nil {
		return c.Fail(fiber.StatusForbidden, err.Error())
	}

	return c.OK(data)
}
