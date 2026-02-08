package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"asum/internal/user"
	"asum/pkg/errorx"
	"asum/pkg/logx"
	"asum/pkg/models"
	"asum/pkg/queue"
	"asum/pkg/rdb"
	"asum/pkg/token"
	"asum/pkg/utils"
)

const (
	codeExpiry    = 10 * time.Minute
	confirmExpiry = 24 * time.Hour
	resetExpiry   = 1 * time.Hour

	codeCachePrefix    = "verify:code:"
	tokenCachePrefix   = "verify:token:"
	confirmCachePrefix = "verify:confirm:"
	resetCachePrefix   = "reset:token:"
	rateLimitPrefix    = "rate:email:"
)

type Service interface {
	Login(ctx context.Context, req *LoginReq) (*LoginResp, error)
	Register(ctx context.Context, req *RegisterReq) (*RegisterResp, error)
	Verify(ctx context.Context, req *VerifyReq) (*VerifyResp, error)
	ConfirmCode(ctx context.Context, email, code string) (*ConfirmResp, error)
	ConfirmURL(ctx context.Context, req *ConfirmReq) (*ConfirmResp, error)
	RefreshToken(ctx context.Context, req *RefreshReq) (*RefreshResp, error)
	ResetPassword(ctx context.Context, req *ResetPasswordReq) (*VerifyResp, error)
	ResetPasswordConfirm(ctx context.Context, req *ResetPasswordConfirmReq) (*ConfirmResp, error)
}

type service struct {
	userRepo user.Repository
	jwt      *token.Manager
	cache    *rdb.Client
	baseURL  string
	q        *queue.RedisQueue[*EmailJob]
}

func NewService(
	userRepo user.Repository,
	emailQueue *queue.RedisQueue[*EmailJob],
	jwtMgr *token.Manager,
	cache *rdb.Client,
	baseURL string,
) Service {
	return &service{
		userRepo: userRepo,
		jwt:      jwtMgr,
		cache:    cache,
		baseURL:  baseURL,
		q:        emailQueue,
	}
}

func (s *service) Login(ctx context.Context, req *LoginReq) (*LoginResp, error) {
	email := utils.SanitizeEmail(req.Email)
	if err := utils.ValidateEmail(email); err != nil {
		return nil, err
	}

	u, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			return nil, errorx.ErrInvalidCredentials
		}
		return nil, err
	}

	if u.Status == models.StatusInactive {
		return nil, errorx.ErrUserNotActive
	}
	if u.Status == models.StatusBanned {
		return nil, errorx.ErrUserBanned
	}

	if !utils.CheckPassword(req.Password, u.Password) {
		return nil, errorx.ErrInvalidCredentials
	}

	accessToken, err := s.jwt.GenerateAccessToken(u.ID, u.Email, int(u.Level))
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, err := s.jwt.GenerateRefreshToken(u.ID, u.Email, int(u.Level))
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	_ = s.userRepo.UpdateLoginTime(ctx, u.ID)

	_ = s.userRepo.AddLog(ctx, &models.UserLog{
		UserID:    u.ID,
		Type:      models.LogTypeLogin,
		IP:        utils.GetRemoteIP(ctx),
		UserAgent: utils.GetUserAgent(ctx),
	})

	// 	cache, _ := json.Marshal(models.ApiCache{UserLevel: lev, Quota: 0})
	// if err := r.rdb.SetNX(ctx, fmt.Sprintf("apiKey:%s", taskKey), string(cache), 0).Err(); err != nil {
	// 	return err
	// }
	return &LoginResp{
		Token:        accessToken,
		RefreshToken: refreshToken,
		User: &UserInfo{
			ID:    u.ID,
			Name:  u.Name,
			Email: u.Email,
			Level: int(u.Level),
		},
	}, nil
}

func (s *service) Register(ctx context.Context, req *RegisterReq) (*RegisterResp, error) {
	name := utils.SanitizeName(req.Name)
	email := utils.SanitizeEmail(req.Email)
	if err := utils.ValidateName(name); err != nil {
		return nil, err
	}
	if err := utils.ValidateEmail(email); err != nil {
		return nil, err
	}
	if err := utils.ValidatePassword(req.Password); err != nil {
		return nil, err
	}

	exists, err := s.userRepo.ExistsByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, user.ErrUserAlreadyExists
	}

	if s.isRateLimited(ctx, email) {
		return nil, errorx.ErrTooManyRequests
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	u := &models.User{
		Name:     name,
		Email:    email,
		Password: hashedPassword,
		Level:    models.LevelBasic,
		Status:   models.StatusInactive,
	}

	if err := s.userRepo.Create(ctx, u); err != nil {
		return nil, err
	}

	token := utils.GenerateConfirmToken()
	cacheTokenKey := confirmCachePrefix + token

	if err := s.cache.Set(ctx, cacheTokenKey, email, codeExpiry).Err(); err != nil {
		return nil, fmt.Errorf("cache token: %w", err)
	}

	code := utils.GenerateVerificationCode(6)
	cacheKey := codeCachePrefix + email

	if err := s.cache.Set(ctx, cacheKey, code, codeExpiry).Err(); err != nil {
		return nil, fmt.Errorf("cache code: %w", err)
	}

	link := fmt.Sprintf("%s/v1/auth/confirm?token=%s", s.baseURL, token)
	go s.sendEmailAsync(ctx, email, name, code, link, TypeRegister)
	s.setRateLimit(ctx, email)

	return &RegisterResp{
		Message: "注册成功！请查看你的邮箱激活。",
		UserID:  u.ID,
	}, nil
}

func (s *service) sendEmailAsync(ctx context.Context, email, name, code, link string, emailType EmailType) {
	payload := &ConfirmPayload{
		Code: code,
		Link: link,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		logx.Errorf("email marshal err: %v", err)
		return
	}

	job := &EmailJob{
		RequestID: utils.GetRequestID(ctx),
		EmailType: emailType,
		To:        email,
		Name:      name,
		Data:      payloadBytes,
	}
	if err := s.q.Push(ctx, job); err != nil {
		logx.Errorf("push email job err: %s: %v", email, err)
		return
	}
}

func (s *service) Verify(ctx context.Context, req *VerifyReq) (*VerifyResp, error) {
	email := utils.SanitizeEmail(req.Email)
	u, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if u.Status == models.StatusActive {
		return nil, errorx.AccountAlreadyExists
	}

	if s.isRateLimited(ctx, email) {
		return nil, errorx.ErrTooManyRequests
	}

	code := utils.GenerateVerificationCode(6)
	cacheKey := codeCachePrefix + email
	if err := s.cache.Set(ctx, cacheKey, code, codeExpiry).Err(); err != nil {
		return nil, fmt.Errorf("cache code: %w", err)
	}

	go s.sendEmailAsync(ctx, email, u.Name, code, "", TypeVerifyCode)

	s.setRateLimit(ctx, email)

	return &VerifyResp{
		Message:   "验证码已发送,注意查收",
		ExpiresIn: int(codeExpiry.Seconds()),
	}, nil
}

func (s *service) ConfirmCode(ctx context.Context, email, code string) (*ConfirmResp, error) {
	email = utils.SanitizeEmail(email)
	cacheKey := codeCachePrefix + email
	storedCode, err := s.cache.Get(ctx, cacheKey).Result()
	if err != nil {
		return nil, errorx.ErrUrlORCodeInvalid
	}

	if storedCode != code {
		return nil, errorx.ErrCodeInvalid
	}

	u, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	u.Status = models.StatusActive
	if err := s.userRepo.UserActiveAndInit(ctx, u.ID, u.Level); err != nil {
		return nil, err
	}

	_ = s.cache.Del(ctx, cacheKey)
	return &ConfirmResp{
		Message: "成功激活帐号",
		User: &UserInfo{
			ID:    u.ID,
			Name:  u.Name,
			Email: u.Email,
			Level: int(u.Level),
		},
	}, nil
}

func (s *service) ConfirmURL(ctx context.Context, req *ConfirmReq) (*ConfirmResp, error) {
	tokenKey := confirmCachePrefix + req.Token
	email, err := s.cache.Get(ctx, tokenKey).Result()
	if err != nil {
		return nil, errorx.ErrUrlORCodeInvalid
	}

	u, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if u.Status == models.StatusActive {
		return &ConfirmResp{
			Message: "此帐号已被激活。",
		}, nil
	}

	if err := s.userRepo.UserActiveAndInit(ctx, u.ID, u.Level); err != nil {
		return nil, err
	}

	_ = s.cache.Del(ctx, tokenKey)

	return &ConfirmResp{
		Message: "成功激活帐号",
		User: &UserInfo{
			ID:    u.ID,
			Name:  u.Name,
			Email: u.Email,
			Level: int(u.Level),
		},
	}, nil
}

func (s *service) RefreshToken(ctx context.Context, req *RefreshReq) (*RefreshResp, error) {
	accessToken, refreshToken, err := s.jwt.RefreshToken(req.RefreshToken)
	if err != nil {
		return nil, errorx.ErrTokenInvalid
	}

	return &RefreshResp{
		Token:        accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *service) ResetPassword(ctx context.Context, req *ResetPasswordReq) (*VerifyResp, error) {
	email := utils.SanitizeEmail(req.Email)
	u, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			return &VerifyResp{
				Message:   "已发送重置邮件，请注意查收",
				ExpiresIn: int(resetExpiry.Seconds()),
			}, nil
		}
		return nil, err
	}

	if s.isRateLimited(ctx, email) {
		return nil, errorx.ErrTooManyRequests
	}

	token := utils.GenerateConfirmToken()
	tokenKey := resetCachePrefix + token
	if err := s.cache.Set(ctx, tokenKey, email, resetExpiry).Err(); err != nil {
		return nil, fmt.Errorf("cache token: %w", err)
	}
	link := fmt.Sprintf("%s/v1/auth/reset-password?token=%s", s.baseURL, token)
	go s.sendEmailAsync(ctx, email, u.Name, "", link, TypePasswordReset)

	s.setRateLimit(ctx, email)

	return &VerifyResp{
		Message:   "已发送重置邮件，请注意查收",
		ExpiresIn: int(resetExpiry.Seconds()),
	}, nil
}

func (s *service) ResetPasswordConfirm(ctx context.Context, req *ResetPasswordConfirmReq) (*ConfirmResp, error) {
	if err := utils.ValidatePassword(req.NewPassword); err != nil {
		return nil, err
	}

	tokenKey := resetCachePrefix + req.Token
	email, err := s.cache.Get(ctx, tokenKey).Result()
	if err != nil {
		if errors.Is(err, rdb.ErrNotFound) || errors.Is(err, rdb.ErrExpired) {
			return nil, errorx.ErrTokenInvalid
		}
		return nil, err
	}

	u, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return nil, err
	}

	u.Password = hashedPassword
	if err := s.userRepo.Update(ctx, u); err != nil {
		return nil, err
	}

	_ = s.cache.Del(ctx, tokenKey)
	_ = s.userRepo.AddLog(ctx, &models.UserLog{
		UserID:    u.ID,
		Type:      models.LogTypeResetPassword,
		IP:        utils.GetRemoteIP(ctx),
		UserAgent: utils.GetUserAgent(ctx),
	})
	return &ConfirmResp{
		Message: "已成功修改密码",
	}, nil
}

func (s *service) isRateLimited(ctx context.Context, email string) bool {
	key := rateLimitPrefix + email
	return s.cache.Exists(ctx, key).Val() > 0
}

func (s *service) setRateLimit(ctx context.Context, email string) {
	key := rateLimitPrefix + email
	_ = s.cache.Set(ctx, key, "1", 1*time.Minute)
}
