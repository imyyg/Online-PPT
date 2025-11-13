package cache

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCacheService 缓存服务测试基础
func TestNewRedisService(t *testing.T) {
	// 创建模拟 Redis 客户端
	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
		DB:   2,
	})
	defer client.Close()

	service := NewRedisService(client)
	assert.NotNil(t, service)
	assert.NotNil(t, service.client)
}

// TestSetAndGetCaptcha 测试验证码的设置和获取
func TestSetAndGetCaptcha(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
		DB:   2,
	})
	defer client.Close()

	// 清空 test db
	_ = client.FlushDB(context.Background())
	defer client.FlushDB(context.Background())

	service := NewRedisService(client)
	ctx := context.Background()

	captchaID := "test-captcha-123"
	code := "abcdef1234567890"

	// 设置验证码
	err := service.SetCaptcha(ctx, captchaID, code)
	require.NoError(t, err)

	// 获取验证码
	retrievedCode, err := service.GetCaptcha(ctx, captchaID)
	require.NoError(t, err)
	assert.Equal(t, code, retrievedCode)
}

// TestGetCaptchaNotFound 测试获取不存在的验证码
func TestGetCaptchaNotFound(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
		DB:   2,
	})
	defer client.Close()

	_ = client.FlushDB(context.Background())
	defer client.FlushDB(context.Background())

	service := NewRedisService(client)
	ctx := context.Background()

	// 获取不存在的验证码
	_, err := service.GetCaptcha(ctx, "non-existent")
	assert.Error(t, err)
	assert.Equal(t, redis.Nil, err)
}

// TestDeleteCaptcha 测试删除验证码
func TestDeleteCaptcha(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
		DB:   2,
	})
	defer client.Close()

	_ = client.FlushDB(context.Background())
	defer client.FlushDB(context.Background())

	service := NewRedisService(client)
	ctx := context.Background()

	captchaID := "test-captcha-del"
	code := "test-code"

	// 设置
	err := service.SetCaptcha(ctx, captchaID, code)
	require.NoError(t, err)

	// 验证存在
	retrievedCode, err := service.GetCaptcha(ctx, captchaID)
	require.NoError(t, err)
	assert.Equal(t, code, retrievedCode)

	// 删除
	err = service.DeleteCaptcha(ctx, captchaID)
	require.NoError(t, err)

	// 验证已删除
	_, err = service.GetCaptcha(ctx, captchaID)
	assert.Error(t, err)
}

// TestSetAndGetEmailCode 测试邮件验证码的设置和获取
func TestSetAndGetEmailCode(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
		DB:   2,
	})
	defer client.Close()

	_ = client.FlushDB(context.Background())
	defer client.FlushDB(context.Background())

	service := NewRedisService(client)
	ctx := context.Background()

	email := "test@example.com"
	data := &EmailCodeData{
		Code:      "123456",
		Attempts:  0,
		CreatedAt: time.Now(),
	}

	// 设置邮件验证码
	err := service.SetEmailCode(ctx, email, data)
	require.NoError(t, err)

	// 获取邮件验证码
	retrievedData, err := service.GetEmailCode(ctx, email)
	require.NoError(t, err)
	assert.Equal(t, data.Code, retrievedData.Code)
	assert.Equal(t, data.Attempts, retrievedData.Attempts)
}

// TestGetEmailCodeNotFound 测试获取不存在的邮件验证码
func TestGetEmailCodeNotFound(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
		DB:   2,
	})
	defer client.Close()

	_ = client.FlushDB(context.Background())
	defer client.FlushDB(context.Background())

	service := NewRedisService(client)
	ctx := context.Background()

	// 获取不存在的邮件验证码
	_, err := service.GetEmailCode(ctx, "nonexistent@example.com")
	assert.Error(t, err)
	assert.Equal(t, redis.Nil, err)
}

// TestDeleteEmailCode 测试删除邮件验证码
func TestDeleteEmailCode(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
		DB:   2,
	})
	defer client.Close()

	_ = client.FlushDB(context.Background())
	defer client.FlushDB(context.Background())

	service := NewRedisService(client)
	ctx := context.Background()

	email := "test@example.com"
	data := &EmailCodeData{
		Code:      "123456",
		Attempts:  0,
		CreatedAt: time.Now(),
	}

	// 设置
	err := service.SetEmailCode(ctx, email, data)
	require.NoError(t, err)

	// 验证存在
	_, err = service.GetEmailCode(ctx, email)
	require.NoError(t, err)

	// 删除
	err = service.DeleteEmailCode(ctx, email)
	require.NoError(t, err)

	// 验证已删除
	_, err = service.GetEmailCode(ctx, email)
	assert.Error(t, err)
}

// TestIncrementEmailCodeAttempts 测试增加邮件验证码尝试次数
func TestIncrementEmailCodeAttempts(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
		DB:   2,
	})
	defer client.Close()

	_ = client.FlushDB(context.Background())
	defer client.FlushDB(context.Background())

	service := NewRedisService(client)
	ctx := context.Background()

	email := "test@example.com"
	data := &EmailCodeData{
		Code:      "123456",
		Attempts:  2,
		CreatedAt: time.Now(),
	}

	// 设置
	err := service.SetEmailCode(ctx, email, data)
	require.NoError(t, err)

	// 增加尝试次数
	err = service.IncrementEmailCodeAttempts(ctx, email)
	require.NoError(t, err)

	// 验证尝试次数已增加
	retrievedData, err := service.GetEmailCode(ctx, email)
	require.NoError(t, err)
	assert.Equal(t, 3, retrievedData.Attempts)
}

// TestSetRateLimit 测试设置频率限制
func TestSetRateLimit(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
		DB:   2,
	})
	defer client.Close()

	_ = client.FlushDB(context.Background())
	defer client.FlushDB(context.Background())

	service := NewRedisService(client)
	ctx := context.Background()

	email := "test@example.com"

	// 设置频率限制
	err := service.SetRateLimit(ctx, email, 60*time.Second)
	require.NoError(t, err)

	// 验证限制已设置
	limited, err := service.CheckRateLimit(ctx, email)
	require.NoError(t, err)
	assert.True(t, limited)
}

// TestCheckRateLimitExpired 测试频率限制过期
func TestCheckRateLimitExpired(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
		DB:   2,
	})
	defer client.Close()

	_ = client.FlushDB(context.Background())
	defer client.FlushDB(context.Background())

	service := NewRedisService(client)
	ctx := context.Background()

	email := "test@example.com"

	// 设置很短的频率限制（1秒）
	err := service.SetRateLimit(ctx, email, 1*time.Second)
	require.NoError(t, err)

	// 验证限制已设置
	limited, err := service.CheckRateLimit(ctx, email)
	require.NoError(t, err)
	assert.True(t, limited)

	// 等待过期
	time.Sleep(1100 * time.Millisecond)

	// 验证限制已过期
	limited, err = service.CheckRateLimit(ctx, email)
	require.NoError(t, err)
	assert.False(t, limited)
}

// TestCheckRateLimitNotSet 测试检查未设置的频率限制
func TestCheckRateLimitNotSet(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
		DB:   2,
	})
	defer client.Close()

	_ = client.FlushDB(context.Background())
	defer client.FlushDB(context.Background())

	service := NewRedisService(client)
	ctx := context.Background()

	// 检查未设置的频率限制
	limited, err := service.CheckRateLimit(ctx, "never-set@example.com")
	require.NoError(t, err)
	assert.False(t, limited)
}

// TestCaptchaTTL 测试验证码 TTL 过期
func TestCaptchaTTL(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
		DB:   2,
	})
	defer client.Close()

	_ = client.FlushDB(context.Background())
	defer client.FlushDB(context.Background())

	service := NewRedisService(client)
	ctx := context.Background()

	captchaID := "ttl-test"
	code := "test-code"

	// 设置验证码（在实际代码中为 5 分钟）
	err := service.SetCaptcha(ctx, captchaID, code)
	require.NoError(t, err)

	// 验证 TTL 大约为 5 分钟
	key := "captcha:" + captchaID
	ttl := client.TTL(ctx, key).Val()
	assert.Greater(t, ttl, time.Minute*4)
	assert.LessOrEqual(t, ttl, time.Minute*5)
}

// TestEmailCodeTTL 测试邮件验证码 TTL
func TestEmailCodeTTL(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
		DB:   2,
	})
	defer client.Close()

	_ = client.FlushDB(context.Background())
	defer client.FlushDB(context.Background())

	service := NewRedisService(client)
	ctx := context.Background()

	email := "ttl@example.com"
	data := &EmailCodeData{
		Code:      "123456",
		Attempts:  0,
		CreatedAt: time.Now(),
	}

	err := service.SetEmailCode(ctx, email, data)
	require.NoError(t, err)

	// 验证 TTL 大约为 10 分钟
	key := "email_code:" + email
	ttl := client.TTL(ctx, key).Val()
	assert.Greater(t, ttl, time.Minute*9)
	assert.LessOrEqual(t, ttl, time.Minute*10)
}

// TestMultipleCodes 测试多个验证码并存
func TestMultipleCodes(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
		DB:   2,
	})
	defer client.Close()

	_ = client.FlushDB(context.Background())
	defer client.FlushDB(context.Background())

	service := NewRedisService(client)
	ctx := context.Background()

	// 设置多个邮件验证码
	emails := []string{"user1@example.com", "user2@example.com", "user3@example.com"}
	for i, email := range emails {
		data := &EmailCodeData{
			Code:      fmt.Sprintf("%06d", i+1),
			Attempts:  0,
			CreatedAt: time.Now(),
		}
		err := service.SetEmailCode(ctx, email, data)
		require.NoError(t, err)
	}

	// 验证每个都能独立获取
	for i, email := range emails {
		retrieved, err := service.GetEmailCode(ctx, email)
		require.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("%06d", i+1), retrieved.Code)
	}
}

// BenchmarkSetCaptcha 基准测试 SetCaptcha
func BenchmarkSetCaptcha(b *testing.B) {
	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
		DB:   2,
	})
	defer client.Close()
	defer client.FlushDB(context.Background())

	service := NewRedisService(client)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = service.SetCaptcha(ctx, "captcha-"+string(rune(i)), "code")
	}
}

// BenchmarkGetCaptcha 基准测试 GetCaptcha
func BenchmarkGetCaptcha(b *testing.B) {
	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
		DB:   2,
	})
	defer client.Close()
	_ = client.FlushDB(context.Background())
	defer client.FlushDB(context.Background())

	service := NewRedisService(client)
	ctx := context.Background()

	// 预先设置
	_ = service.SetCaptcha(ctx, "bench-captcha", "code")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.GetCaptcha(ctx, "bench-captcha")
	}
}
