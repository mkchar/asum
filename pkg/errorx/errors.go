package errorx

import "errors"

// 通用

var (
	ErrInvalidPayload     = errors.New("无效的请求参数")
	ErrTooManyRequests    = errors.New("请求过于频繁")
	ErrInternal           = errors.New("服务器内部错误")
	ErrInvalidRequestBody = errors.New("无效的请求参数")
	ErrUnauthorized       = errors.New("未授权")
	ErrTokenExpired       = errors.New("令牌已过期")
)

// task
var (
	ErrTaskNotFound      = errors.New("任务不存在")
	ErrTaskAlreadyExists = errors.New("任务已存在")
	ErrInvalidTaskKey    = errors.New("无效的API")
)

// auth
var (
	ErrInvalidCredentials   = errors.New("非法帐号信息")
	ErrUserNotActive        = errors.New("此用户未激活")
	ErrUserBanned           = errors.New("此用户被封号")
	ErrCodeExpired          = errors.New("验证码过期")
	ErrCodeInvalid          = errors.New("无效的验证码")
	ErrTokenInvalid         = errors.New("无效的Token")
	ErrUrlORCodeInvalid     = errors.New("该链接已过期或者失效")
	ErrMailerNotInitialized = errors.New("客户端未初始化")
	ErrUnknownEmailType     = errors.New("未识别的邮件")

	AccountAlreadyExists = errors.New("用户已存在，请勿重复注册。")
)

// user
var (
	ErrQuota = errors.New("余额不足")
)
var (
	ErrInvalidIP = errors.New("无效的IP")
)

var (
	ErrRateLimited       = errors.New("限制访问")
	ErrRateLimitServeice = errors.New("限流计数错误")
)
