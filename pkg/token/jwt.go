package token

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
)

type Config struct {
	Secret        string        `yaml:"secret"`
	Issuer        string        `yaml:"issuer"`
	AccessExpiry  time.Duration `yaml:"access_expiry"`
	RefreshExpiry time.Duration `yaml:"refresh_expiry"`
}

type Claims struct {
	UserID uint64 `json:"userId"`
	Email  string `json:"email"`
	Level  int    `json:"level"`
	jwt.RegisteredClaims
}

type Manager struct {
	cfg Config
}

func NewManager(cfg Config) *Manager {
	if cfg.AccessExpiry == 0 {
		cfg.AccessExpiry = 24 * time.Hour
	}
	if cfg.RefreshExpiry == 0 {
		cfg.RefreshExpiry = 7 * 24 * time.Hour
	}
	return &Manager{cfg: cfg}
}

func (m *Manager) GenerateAccessToken(userID uint64, email string, level int) (string, error) {
	return m.generateToken(userID, email, level, m.cfg.AccessExpiry)
}

func (m *Manager) GenerateRefreshToken(userID uint64, email string, level int) (string, error) {
	return m.generateToken(userID, email, level, m.cfg.RefreshExpiry)
}

func (m *Manager) generateToken(userID uint64, email string, level int, expiry time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		Email:  email,
		Level:  level,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.cfg.Issuer,
			Subject:   email,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.cfg.Secret))
}

func (m *Manager) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(m.cfg.Secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func (m *Manager) RefreshToken(refreshToken string) (string, string, error) {
	claims, err := m.ParseToken(refreshToken)
	if err != nil {
		return "", "", err
	}

	accessToken, err := m.GenerateAccessToken(claims.UserID, claims.Email, claims.Level)
	if err != nil {
		return "", "", err
	}

	newRefreshToken, err := m.GenerateRefreshToken(claims.UserID, claims.Email, claims.Level)
	if err != nil {
		return "", "", err
	}

	return accessToken, newRefreshToken, nil
}
