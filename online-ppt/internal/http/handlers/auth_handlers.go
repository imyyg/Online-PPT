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
