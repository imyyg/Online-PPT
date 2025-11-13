package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"net/mail"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"

	"online-ppt/internal/cache"
	"online-ppt/internal/captcha"
	mailpkg "online-ppt/internal/mail"
	"online-ppt/internal/storage"
)

var (
	// ErrInvalidCredentials is returned when email or password is incorrect.
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrEmailAlreadyRegistered indicates the email already exists.
	ErrEmailAlreadyRegistered = errors.New("email already registered")
	// ErrSessionNotFound indicates the refresh token does not exist or is revoked.
	ErrSessionNotFound = errors.New("session not found")
	// ErrInvalidCaptcha indicates the captcha is incorrect or expired.
	ErrInvalidCaptcha = errors.New("invalid captcha")
	// ErrRateLimited indicates too many requests.
	ErrRateLimited = errors.New("rate limited")
	// ErrInvalidVerificationCode indicates the email verification code is incorrect or expired.
	ErrInvalidVerificationCode = errors.New("invalid verification code")
	// ErrTooManyAttempts indicates too many verification attempts.
	ErrTooManyAttempts = errors.New("too many verification attempts")
)

// Service coordinates authentication workflows.
type Service struct {
	repo    *Repository
	tokens  *TokenManager
	audit   *storage.AuditLogger
	cache   cache.Service
	captcha captcha.Service
	mail    mailpkg.Service
	clockFn func() time.Time
}

// AuthResult represents the outcome of a login or refresh invocation.
type AuthResult struct {
	User             UserAccount
	AccessToken      string
	AccessExpiresAt  time.Time
	RefreshToken     string
	RefreshExpiresAt time.Time
}

// NewService builds a Service with sane defaults.
func NewService(
	repo *Repository,
	tokens *TokenManager,
	audit *storage.AuditLogger,
	cacheService cache.Service,
	captchaService captcha.Service,
	mailService mailpkg.Service,
) (*Service, error) {
	if repo == nil {
		return nil, fmt.Errorf("auth service requires repository")
	}
	if tokens == nil {
		return nil, fmt.Errorf("auth service requires token manager")
	}
	if audit == nil {
		audit = storage.NewAuditLogger(nil)
	}
	if cacheService == nil {
		return nil, fmt.Errorf("auth service requires cache service")
	}
	if captchaService == nil {
		return nil, fmt.Errorf("auth service requires captcha service")
	}
	if mailService == nil {
		return nil, fmt.Errorf("auth service requires mail service")
	}
	return &Service{
		repo:    repo,
		tokens:  tokens,
		audit:   audit,
		cache:   cacheService,
		captcha: captchaService,
		mail:    mailService,
		clockFn: time.Now,
	}, nil
}

// Register creates a new user account.
func (s *Service) Register(ctx context.Context, email, password string) (UserAccount, error) {
	normalized, err := normalizeEmail(email)
	if err != nil {
		s.audit.Log("auth.register", map[string]any{
			"status": "validation_failed",
			"reason": err.Error(),
		})
		return UserAccount{}, err
	}
	if err := validatePassword(password); err != nil {
		s.audit.Log("auth.register", map[string]any{
			"status": "validation_failed",
			"reason": err.Error(),
		})
		return UserAccount{}, err
	}

	hash, err := HashPassword(password)
	if err != nil {
		s.audit.Log("auth.register", map[string]any{
			"status": "error",
			"reason": err.Error(),
		})
		return UserAccount{}, err
	}

	uuidValue := uuid.NewString()

	user, err := s.repo.CreateUser(ctx, normalized, hash, uuidValue)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			s.audit.Log("auth.register", map[string]any{
				"status": "conflict",
				"reason": ErrEmailAlreadyRegistered.Error(),
				"email":  normalized,
			})
			return UserAccount{}, ErrEmailAlreadyRegistered
		}
		s.audit.Log("auth.register", map[string]any{
			"status": "error",
			"reason": err.Error(),
			"email":  normalized,
		})
		return UserAccount{}, err
	}

	s.audit.Log("auth.register", map[string]any{
		"status": "success",
		"userId": user.ID,
	})
	return user, nil
}

// Login authenticates a user and provisions tokens plus session state.
func (s *Service) Login(ctx context.Context, email, password, fingerprint string) (AuthResult, error) {
	normalized, err := normalizeEmail(email)
	if err != nil {
		s.audit.Log("auth.login", map[string]any{
			"status": "validation_failed",
			"reason": err.Error(),
		})
		return AuthResult{}, err
	}

	user, err := s.repo.GetUserByEmail(ctx, normalized)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.audit.Log("auth.login", map[string]any{
				"status": "invalid_credentials",
				"email":  normalized,
			})
			return AuthResult{}, ErrInvalidCredentials
		}
		s.audit.Log("auth.login", map[string]any{
			"status": "error",
			"reason": err.Error(),
			"email":  normalized,
		})
		return AuthResult{}, err
	}

	match, err := VerifyPassword(user.PasswordHash, password)
	if err != nil {
		s.audit.Log("auth.login", map[string]any{
			"status": "error",
			"userId": user.ID,
			"reason": err.Error(),
		})
		return AuthResult{}, err
	}
	if !match {
		s.audit.Log("auth.login", map[string]any{
			"status": "invalid_credentials",
			"userId": user.ID,
		})
		return AuthResult{}, ErrInvalidCredentials
	}

	result, err := s.issueSession(ctx, user, fingerprint, "auth.login")
	if err != nil {
		s.audit.Log("auth.login", map[string]any{
			"status": "error",
			"userId": user.ID,
			"reason": err.Error(),
		})
		return AuthResult{}, err
	}
	return result, nil
}

// Refresh exchanges a valid refresh token for new credentials.
func (s *Service) Refresh(ctx context.Context, refreshToken, fingerprint string) (AuthResult, error) {
	if refreshToken == "" {
		s.audit.Log("auth.refresh", map[string]any{
			"status": "validation_failed",
			"reason": "refresh token required",
		})
		return AuthResult{}, fmt.Errorf("refresh token required")
	}

	hash := HashRefreshToken(refreshToken)
	session, err := s.repo.FindActiveSession(ctx, hash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.audit.Log("auth.refresh", map[string]any{
				"status": "not_found",
				"reason": ErrSessionNotFound.Error(),
			})
			return AuthResult{}, ErrSessionNotFound
		}
		s.audit.Log("auth.refresh", map[string]any{
			"status": "error",
			"reason": err.Error(),
		})
		return AuthResult{}, err
	}
	if session.ExpiresAt.Before(s.clockFn()) {
		s.audit.Log("auth.refresh", map[string]any{
			"status":    "expired",
			"sessionId": session.ID,
		})
		return AuthResult{}, ErrSessionNotFound
	}

	user, err := s.repo.GetUserByID(ctx, session.UserID)
	if err != nil {
		s.audit.Log("auth.refresh", map[string]any{
			"status":    "error",
			"sessionId": session.ID,
			"reason":    err.Error(),
		})
		return AuthResult{}, err
	}

	if err := s.repo.RevokeSession(ctx, session.ID, s.clockFn()); err != nil {
		s.audit.Log("auth.refresh", map[string]any{
			"status":    "error",
			"sessionId": session.ID,
			"reason":    err.Error(),
		})
		return AuthResult{}, err
	}

	result, err := s.issueSession(ctx, user, fingerprint, "auth.refresh")
	if err != nil {
		s.audit.Log("auth.refresh", map[string]any{
			"status": "error",
			"userId": user.ID,
			"reason": err.Error(),
		})
		return AuthResult{}, err
	}
	return result, nil
}

// Logout revokes the provided refresh token.
func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		s.audit.Log("auth.logout", map[string]any{
			"status": "validation_failed",
			"reason": "refresh token required",
		})
		return fmt.Errorf("refresh token required")
	}
	hash := HashRefreshToken(refreshToken)
	session, err := s.repo.FindActiveSession(ctx, hash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.audit.Log("auth.logout", map[string]any{
				"status": "not_found",
				"reason": ErrSessionNotFound.Error(),
			})
			return ErrSessionNotFound
		}
		s.audit.Log("auth.logout", map[string]any{
			"status": "error",
			"reason": err.Error(),
		})
		return err
	}
	if err := s.repo.RevokeSession(ctx, session.ID, s.clockFn()); err != nil {
		s.audit.Log("auth.logout", map[string]any{
			"status":    "error",
			"sessionId": session.ID,
			"reason":    err.Error(),
		})
		return err
	}
	s.audit.Log("auth.logout", map[string]any{
		"status":    "success",
		"sessionId": session.ID,
		"userId":    session.UserID,
	})
	return nil
}

func (s *Service) issueSession(ctx context.Context, user UserAccount, fingerprint string, event string) (AuthResult, error) {
	accessToken, accessExpiry, err := s.tokens.IssueAccessToken(user.ID, user.UUID)
	if err != nil {
		s.audit.Log(event, map[string]any{
			"status": "error",
			"userId": user.ID,
			"reason": err.Error(),
		})
		return AuthResult{}, err
	}

	refreshToken, refreshExpiry, err := s.tokens.GenerateRefreshToken()
	if err != nil {
		s.audit.Log(event, map[string]any{
			"status": "error",
			"userId": user.ID,
			"reason": err.Error(),
		})
		return AuthResult{}, err
	}

	session := UserSession{
		UserID:           user.ID,
		RefreshTokenHash: HashRefreshToken(refreshToken),
		ExpiresAt:        refreshExpiry,
		IssuedAt:         s.clockFn(),
	}
	if fingerprint != "" {
		session.ClientFingerprint = sql.NullString{String: fingerprint, Valid: true}
	}

	created, err := s.repo.CreateSession(ctx, session)
	if err != nil {
		s.audit.Log(event, map[string]any{
			"status": "error",
			"userId": user.ID,
			"reason": err.Error(),
		})
		return AuthResult{}, err
	}

	if err := s.repo.UpdateLastLogin(ctx, user.ID, s.clockFn()); err != nil {
		s.audit.Log("auth.login.update_last_login", map[string]any{
			"status": "error",
			"userId": user.ID,
			"error":  err.Error(),
		})
	}

	s.audit.Log(event, map[string]any{
		"status":    "success",
		"userId":    user.ID,
		"sessionId": created.ID,
	})

	return AuthResult{
		User:             user,
		AccessToken:      accessToken,
		AccessExpiresAt:  accessExpiry,
		RefreshToken:     refreshToken,
		RefreshExpiresAt: refreshExpiry,
	}, nil
}

func normalizeEmail(email string) (string, error) {
	trimmed := strings.TrimSpace(strings.ToLower(email))
	if trimmed == "" {
		return "", fmt.Errorf("email required")
	}
	if _, err := mail.ParseAddress(trimmed); err != nil {
		return "", fmt.Errorf("invalid email: %w", err)
	}
	return trimmed, nil
}

func validatePassword(password string) error {
	if len(password) < 10 {
		return fmt.Errorf("password must be at least 10 characters")
	}
	return nil
}

// GenerateCaptcha 生成图形验证码
func (s *Service) GenerateCaptcha(ctx context.Context) (captchaID, imageBase64 string, expiresIn int, err error) {
	captchaID, imageBase64, err = s.captcha.Generate(ctx)
	if err != nil {
		s.audit.Log("auth.captcha.generate", map[string]any{
			"status": "error",
			"reason": err.Error(),
		})
		return "", "", 0, fmt.Errorf("failed to generate captcha: %w", err)
	}

	s.audit.Log("auth.captcha.generate", map[string]any{
		"status":    "success",
		"captchaId": captchaID,
		"expiresIn": 300,
	})

	return captchaID, imageBase64, 300, nil
}

// SendVerificationCode 发送邮箱验证码
func (s *Service) SendVerificationCode(ctx context.Context, email, captchaID, captchaCode string) (int, error) {
	// 验证邮箱格式
	normalized, err := normalizeEmail(email)
	if err != nil {
		s.audit.Log("auth.send_code", map[string]any{
			"status": "validation_failed",
			"reason": err.Error(),
		})
		return 0, err
	}

	// 验证图形验证码
	valid, err := s.captcha.Verify(ctx, captchaID, captchaCode)
	if err != nil || !valid {
		s.audit.Log("auth.send_code", map[string]any{
			"status": "invalid_captcha",
			"email":  normalized,
		})
		return 0, ErrInvalidCaptcha
	}

	// 检查频率限制
	limited, err := s.cache.CheckRateLimit(ctx, normalized)
	if err != nil {
		s.audit.Log("auth.send_code", map[string]any{
			"status": "error",
			"email":  normalized,
			"reason": err.Error(),
		})
		return 0, err
	}
	if limited {
		s.audit.Log("auth.send_code", map[string]any{
			"status": "rate_limited",
			"email":  normalized,
		})
		return 0, ErrRateLimited
	}

	// 生成6位数字验证码
	code, err := generateEmailCode()
	if err != nil {
		s.audit.Log("auth.send_code", map[string]any{
			"status": "error",
			"email":  normalized,
			"reason": err.Error(),
		})
		return 0, err
	}

	// 存储到缓存
	codeData := &cache.EmailCodeData{
		Code:      code,
		Attempts:  0,
		CreatedAt: s.clockFn(),
	}
	if err := s.cache.SetEmailCode(ctx, normalized, codeData); err != nil {
		s.audit.Log("auth.send_code", map[string]any{
			"status": "error",
			"email":  normalized,
			"reason": err.Error(),
		})
		return 0, err
	}

	// 设置频率限制
	if err := s.cache.SetRateLimit(ctx, normalized, 60*time.Second); err != nil {
		s.audit.Log("auth.send_code", map[string]any{
			"status": "error",
			"email":  normalized,
			"reason": err.Error(),
		})
		return 0, err
	}

	// 发送邮件
	if err := s.mail.SendVerificationCode(normalized, code); err != nil {
		s.audit.Log("auth.send_code", map[string]any{
			"status": "error",
			"email":  normalized,
			"reason": err.Error(),
		})
		return 0, fmt.Errorf("failed to send email: %w", err)
	}

	s.audit.Log("auth.send_code", map[string]any{
		"status": "success",
		"email":  normalized,
	})

	return 600, nil // 10分钟有效期
}

// RegisterWithEmailCode 使用邮箱验证码注册
func (s *Service) RegisterWithEmailCode(ctx context.Context, email, password, emailCode string) (UserAccount, error) {
	// 验证邮箱和密码格式
	normalized, err := normalizeEmail(email)
	if err != nil {
		s.audit.Log("auth.register", map[string]any{
			"status": "validation_failed",
			"reason": err.Error(),
		})
		return UserAccount{}, err
	}
	if err := validatePassword(password); err != nil {
		s.audit.Log("auth.register", map[string]any{
			"status": "validation_failed",
			"reason": err.Error(),
		})
		return UserAccount{}, err
	}

	// 验证邮箱验证码
	codeData, err := s.cache.GetEmailCode(ctx, normalized)
	if err != nil {
		s.audit.Log("auth.register", map[string]any{
			"status": "invalid_code",
			"email":  normalized,
			"reason": "code not found or expired",
		})
		return UserAccount{}, ErrInvalidVerificationCode
	}

	// 检查尝试次数
	if codeData.Attempts >= 5 {
		s.audit.Log("auth.register", map[string]any{
			"status": "too_many_attempts",
			"email":  normalized,
		})
		return UserAccount{}, ErrTooManyAttempts
	}

	// 验证码是否匹配
	if codeData.Code != emailCode {
		// 增加尝试次数
		_ = s.cache.IncrementEmailCodeAttempts(ctx, normalized)
		s.audit.Log("auth.register", map[string]any{
			"status":   "invalid_code",
			"email":    normalized,
			"attempts": codeData.Attempts + 1,
		})
		return UserAccount{}, ErrInvalidVerificationCode
	}

	// 删除验证码
	_ = s.cache.DeleteEmailCode(ctx, normalized)

	// 创建用户账号
	hash, err := HashPassword(password)
	if err != nil {
		s.audit.Log("auth.register", map[string]any{
			"status": "error",
			"reason": err.Error(),
		})
		return UserAccount{}, err
	}

	uuidValue := uuid.NewString()

	user, err := s.repo.CreateUser(ctx, normalized, hash, uuidValue)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			s.audit.Log("auth.register", map[string]any{
				"status": "conflict",
				"reason": ErrEmailAlreadyRegistered.Error(),
				"email":  normalized,
			})
			return UserAccount{}, ErrEmailAlreadyRegistered
		}
		s.audit.Log("auth.register", map[string]any{
			"status": "error",
			"reason": err.Error(),
			"email":  normalized,
		})
		return UserAccount{}, err
	}

	s.audit.Log("auth.register", map[string]any{
		"status": "success",
		"userId": user.ID,
	})
	return user, nil
}

// generateEmailCode 生成6位数字验证码
func generateEmailCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}
