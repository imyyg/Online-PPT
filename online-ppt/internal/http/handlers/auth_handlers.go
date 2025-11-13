package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"online-ppt/internal/auth"
	"online-ppt/internal/config"
)

// AuthHandler exposes HTTP endpoints for authentication flows.
type AuthHandler struct {
	service *auth.Service
	cfg     *config.Config
}

// NewAuthHandler constructs a new AuthHandler instance.
func NewAuthHandler(service *auth.Service, cfg *config.Config) *AuthHandler {
	return &AuthHandler{service: service, cfg: cfg}
}

// Register handles POST /auth/register.
func (h *AuthHandler) Register(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	if _, err := h.service.Register(c.Request.Context(), req.Email, req.Password); err != nil {
		handleAuthError(c, err)
		return
	}

	fingerprint := clientFingerprint(c)
	result, err := h.service.Login(c.Request.Context(), req.Email, req.Password, fingerprint)
	if err != nil {
		handleAuthError(c, err)
		return
	}

	setRefreshCookie(c, result.RefreshToken, result.RefreshExpiresAt, h.cfg)
	c.JSON(http.StatusCreated, serializeAuthResult(result))
}

// Login handles POST /auth/login.
func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	fingerprint := clientFingerprint(c)
	result, err := h.service.Login(c.Request.Context(), req.Email, req.Password, fingerprint)
	if err != nil {
		handleAuthError(c, err)
		return
	}

	setRefreshCookie(c, result.RefreshToken, result.RefreshExpiresAt, h.cfg)
	c.JSON(http.StatusOK, serializeAuthResult(result))
}

// Refresh handles POST /auth/refresh.
func (h *AuthHandler) Refresh(c *gin.Context) {
	refreshToken := readRefreshTokenFromRequest(c)
	if refreshToken == "" {
		writeError(c, http.StatusBadRequest, "invalid_request", "refresh token required")
		return
	}

	fingerprint := clientFingerprint(c)
	result, err := h.service.Refresh(c.Request.Context(), refreshToken, fingerprint)
	if err != nil {
		handleAuthError(c, err)
		return
	}

	setRefreshCookie(c, result.RefreshToken, result.RefreshExpiresAt, h.cfg)
	c.JSON(http.StatusOK, serializeAuthResult(result))
}

// Logout handles POST /auth/logout.
func (h *AuthHandler) Logout(c *gin.Context) {
	refreshToken := readRefreshTokenFromRequest(c)
	if refreshToken == "" {
		writeError(c, http.StatusBadRequest, "invalid_request", "refresh token required")
		return
	}

	if err := h.service.Logout(c.Request.Context(), refreshToken); err != nil {
		handleAuthError(c, err)
		return
	}

	setRefreshCookie(c, "", time.Time{}, h.cfg)
	c.Status(http.StatusNoContent)
}

func serializeAuthResult(result auth.AuthResult) gin.H {
	expiresIn := int(time.Until(result.AccessExpiresAt).Seconds())
	if expiresIn < 0 {
		expiresIn = 0
	}

	var lastLogin *time.Time
	if result.User.LastLoginAt.Valid {
		lastLogin = &result.User.LastLoginAt.Time
	}

	return gin.H{
		"user": gin.H{
			"id":          result.User.ID,
			"email":       result.User.Email,
			"status":      result.User.Status,
			"lastLoginAt": lastLogin,
		},
		"accessToken": result.AccessToken,
		"expiresIn":   expiresIn,
	}
}

func handleAuthError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, auth.ErrEmailAlreadyRegistered):
		writeError(c, http.StatusConflict, "email_exists", "email already registered")
	case errors.Is(err, auth.ErrInvalidCredentials):
		writeError(c, http.StatusUnauthorized, "invalid_credentials", "invalid email or password")
	case errors.Is(err, auth.ErrSessionNotFound):
		writeError(c, http.StatusUnauthorized, "session_not_found", "refresh token invalid or expired")
	default:
		writeError(c, http.StatusInternalServerError, "server_error", err.Error())
	}
}

func writeError(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{
		"code":    code,
		"message": message,
	})
}

func clientFingerprint(c *gin.Context) string {
	fingerprint := c.GetHeader("X-Client-Fingerprint")
	if fingerprint == "" {
		fingerprint = c.GetHeader("X-Device-ID")
	}
	return fingerprint
}

func readRefreshTokenFromRequest(c *gin.Context) string {
	var body struct {
		RefreshToken string `json:"refreshToken"`
	}
	if c.Request.Body != nil && c.Request.ContentLength > 0 {
		_ = c.ShouldBindJSON(&body)
	}

	if body.RefreshToken != "" {
		return body.RefreshToken
	}

	if token, err := c.Cookie("refresh_token"); err == nil {
		return token
	}

	return ""
}

func setRefreshCookie(c *gin.Context, token string, expiresAt time.Time, cfg *config.Config) {
	if token == "" {
		c.SetCookie("refresh_token", "", -1, "/", "", false, true)
		return
	}

	duration := int(time.Until(expiresAt).Seconds())
	if duration <= 0 {
		duration = 0
	}

	c.SetCookie("refresh_token", token, duration, "/", "", false, true)
}

// GetCaptcha handles GET /auth/captcha - 获取图形验证码
func (h *AuthHandler) GetCaptcha(c *gin.Context) {
	captchaID, imageBase64, expiresIn, err := h.service.GenerateCaptcha(c.Request.Context())
	if err != nil {
		writeError(c, http.StatusInternalServerError, "server_error", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"captcha_id": captchaID,
		"image":      imageBase64,
		"expires_in": expiresIn,
	})
}

// SendVerificationCode handles POST /auth/send-verification-code - 发送邮箱验证码
func (h *AuthHandler) SendVerificationCode(c *gin.Context) {
	var req struct {
		Email       string `json:"email" binding:"required"`
		CaptchaID   string `json:"captcha_id" binding:"required"`
		CaptchaCode string `json:"captcha_code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	expiresIn, err := h.service.SendVerificationCode(
		c.Request.Context(),
		req.Email,
		req.CaptchaID,
		req.CaptchaCode,
	)
	if err != nil {
		handleVerificationError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "验证码已发送",
		"expires_in": expiresIn,
	})
}

// RegisterWithCode handles POST /auth/register - 使用邮箱验证码注册
func (h *AuthHandler) RegisterWithCode(c *gin.Context) {
	var req struct {
		Email     string `json:"email" binding:"required"`
		Password  string `json:"password" binding:"required"`
		EmailCode string `json:"email_code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	_, err := h.service.RegisterWithEmailCode(
		c.Request.Context(),
		req.Email,
		req.Password,
		req.EmailCode,
	)
	if err != nil {
		handleVerificationError(c, err)
		return
	}

	// 注册成功后自动登录
	fingerprint := clientFingerprint(c)
	result, err := h.service.Login(c.Request.Context(), req.Email, req.Password, fingerprint)
	if err != nil {
		handleAuthError(c, err)
		return
	}

	setRefreshCookie(c, result.RefreshToken, result.RefreshExpiresAt, h.cfg)
	c.JSON(http.StatusCreated, serializeAuthResult(result))
}

// handleVerificationError 处理验证码相关错误
func handleVerificationError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, auth.ErrInvalidCaptcha):
		writeError(c, http.StatusBadRequest, "invalid_captcha", "图形验证码错误或已过期")
	case errors.Is(err, auth.ErrRateLimited):
		writeError(c, http.StatusTooManyRequests, "rate_limited", "请求过于频繁，请稍后再试")
	case errors.Is(err, auth.ErrInvalidVerificationCode):
		writeError(c, http.StatusBadRequest, "invalid_code", "邮箱验证码错误或已过期")
	case errors.Is(err, auth.ErrTooManyAttempts):
		writeError(c, http.StatusTooManyRequests, "too_many_attempts", "验证失败次数过多，请重新获取验证码")
	case errors.Is(err, auth.ErrEmailAlreadyRegistered):
		writeError(c, http.StatusConflict, "email_exists", "邮箱已注册")
	default:
		writeError(c, http.StatusInternalServerError, "server_error", err.Error())
	}
}
