package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"online-ppt/internal/auth"
	"online-ppt/internal/cache"
	"online-ppt/internal/captcha"
	"online-ppt/internal/config"
	internalhttp "online-ppt/internal/http"
	"online-ppt/internal/http/handlers"
	"online-ppt/internal/storage"
)

// TestEmailVerificationFlow 测试完整的邮箱验证流程
func TestEmailVerificationFlow(t *testing.T) {
	// 跳过集成测试（需要真实的 Redis 和 SMTP）
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// 初始化测试配置
	cfg := &config.Config{
		Server: config.ServerConfig{Addr: ":0"},
		Security: config.SecurityConfig{
			JWTSecret: "test-secret-key-for-testing-only-do-not-use-in-production",
		},
		Redis: config.RedisConfig{
			Host:     "127.0.0.1",
			Port:     6379,
			DB:       2,
			PoolSize: 10,
		},
	}

	// 初始化 Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		DB:   cfg.Redis.DB,
	})
	defer redisClient.Close()

	ctx := context.Background()
	err := redisClient.Ping(ctx).Err()
	require.NoError(t, err, "Redis should be available")

	// 初始化服务
	cacheService := cache.NewRedisService(redisClient)
	captchaService := captcha.NewService(cacheService)
	mailService := &mockMailService{} // 使用 mock 邮件服务

	// Mock 其他依赖
	authRepo := &mockAuthRepository{}
	tokenManager, _ := auth.NewTokenManager(cfg.Security.JWTSecret, 0, 0)
	auditLogger := storage.NewAuditLogger(nil)

	authService, err := auth.NewService(authRepo, tokenManager, auditLogger, cacheService, captchaService, mailService)
	require.NoError(t, err)

	authHandler := handlers.NewAuthHandler(authService, cfg)

	router := internalhttp.NewRouter(cfg)
	internalhttp.RegisterAuthRoutes(router, authHandler)

	// Test 1: 获取图形验证码
	t.Run("GetCaptcha", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/captcha", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.NotEmpty(t, resp["captcha_id"])
		assert.NotEmpty(t, resp["image"])
		assert.Equal(t, float64(300), resp["expires_in"])
	})

	// Test 2: 发送验证码（需要有效的图形验证码）
	t.Run("SendVerificationCode", func(t *testing.T) {
		// 先获取图形验证码
		captchaID, _, _, err := authService.GenerateCaptcha(ctx)
		require.NoError(t, err)

		// 从缓存获取验证码值（测试用）
		captchaCode, err := cacheService.GetCaptcha(ctx, captchaID)
		require.NoError(t, err)

		// 发送验证码请求
		payload := map[string]string{
			"email":        "test@example.com",
			"captcha_id":   captchaID,
			"captcha_code": captchaCode,
		}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/send-verification-code", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.Contains(t, resp["message"], "验证码已发送")
		assert.Equal(t, float64(600), resp["expires_in"])
	})
}

// mockMailService Mock 邮件服务
type mockMailService struct {
	SentEmails []struct {
		To   string
		Code string
	}
}

func (m *mockMailService) SendVerificationCode(to, code string) error {
	m.SentEmails = append(m.SentEmails, struct {
		To   string
		Code string
	}{To: to, Code: code})
	return nil
}

// mockAuthRepository Mock 认证仓库
type mockAuthRepository struct{}

func (m *mockAuthRepository) CreateUser(ctx context.Context, email, passwordHash, uuidValue string) (auth.UserAccount, error) {
	return auth.UserAccount{
		ID:     1,
		Email:  email,
		Status: "active",
	}, nil
}

func (m *mockAuthRepository) GetUserByEmail(ctx context.Context, email string) (auth.UserAccount, error) {
	return auth.UserAccount{}, fmt.Errorf("user not found")
}

// 实现其他必需的方法...
func (m *mockAuthRepository) GetUserByID(ctx context.Context, id int64) (auth.UserAccount, error) {
	return auth.UserAccount{}, fmt.Errorf("not implemented")
}

func (m *mockAuthRepository) CreateSession(ctx context.Context, userID int64, refreshToken, fingerprint string, expiresAt interface{}) (auth.Session, error) {
	return auth.Session{}, fmt.Errorf("not implemented")
}

func (m *mockAuthRepository) GetSessionByRefreshToken(ctx context.Context, token string) (auth.Session, error) {
	return auth.Session{}, fmt.Errorf("not implemented")
}

func (m *mockAuthRepository) RefreshSession(ctx context.Context, sessionID int64, newRefreshToken string, newExpiresAt interface{}) error {
	return fmt.Errorf("not implemented")
}

func (m *mockAuthRepository) RevokeSession(ctx context.Context, refreshToken string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockAuthRepository) UpdateLastLogin(ctx context.Context, userID int64, loginTime interface{}) error {
	return nil
}
