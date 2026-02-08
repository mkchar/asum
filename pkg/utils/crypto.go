package utils

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	mrand "math/rand"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword 密码哈希
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(bytes), nil
}

// CheckPassword 验证密码
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateRandomKey 生成随机密钥（十六进制）
func GenerateRandomKey(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// GenerateRandomToken 生成随机令牌（Base64 URL 安全）
func GenerateRandomToken(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)
}

// GenerateVerificationCode 生成数字验证码
func GenerateVerificationCode(length int) string {
	if length <= 0 {
		length = 6
	}

	r := mrand.New(mrand.NewSource(time.Now().UnixNano()))
	code := ""
	for i := 0; i < length; i++ {
		code += fmt.Sprintf("%d", r.Intn(10))
	}
	return code
}

// GenerateConfirmToken 生成确认令牌
func GenerateConfirmToken() string {
	return GenerateRandomToken(32)
}
