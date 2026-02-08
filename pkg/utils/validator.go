package utils

import (
	"errors"
	"regexp"
	"strings"
	"unicode/utf8"
)

var (
	ErrInvalidEmail    = errors.New("invalid email format")
	ErrInvalidPassword = errors.New("password must be at least 8 characters")
	ErrInvalidName     = errors.New("name must be 2-50 characters")
	ErrEmptyField      = errors.New("field cannot be empty")
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// ValidateEmail 验证邮箱格式
func ValidateEmail(email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return ErrEmptyField
	}
	if !emailRegex.MatchString(email) {
		return ErrInvalidEmail
	}
	return nil
}

// ValidatePassword 验证密码强度
func ValidatePassword(password string) error {
	if password == "" {
		return ErrEmptyField
	}
	if utf8.RuneCountInString(password) < 8 {
		return ErrInvalidPassword
	}
	return nil
}

// ValidateName 验证用户名
func ValidateName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return ErrEmptyField
	}
	length := utf8.RuneCountInString(name)
	if length < 2 || length > 50 {
		return ErrInvalidName
	}
	return nil
}

func SanitizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func SanitizeName(name string) string {
	return strings.TrimSpace(name)
}
