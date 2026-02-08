package user

import "errors"

var (
	ErrUserNotFound      = errors.New("此用户不存在")
	ErrUserAlreadyExists = errors.New("此用户已存在")
	ErrInvalidPassword   = errors.New("无效的密码")
)
