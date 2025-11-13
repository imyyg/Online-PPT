package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	captchaKeyFormat   = "captcha:%s"
	emailCodeKeyFormat = "email_code:%s"
	rateLimitKeyFormat = "rate_limit:%s"
)

// EmailCodeData 邮箱验证码缓存数据结构
type EmailCodeData struct {
	Code      string    `json:"code"`
	Attempts  int       `json:"attempts"`
	CreatedAt time.Time `json:"created_at"`
}

// Service Redis 缓存服务接口
type Service interface {
	// Captcha operations
	SetCaptcha(ctx context.Context, captchaID, code string) error
	GetCaptcha(ctx context.Context, captchaID string) (string, error)
	DeleteCaptcha(ctx context.Context, captchaID string) error

	// Email code operations
	SetEmailCode(ctx context.Context, email string, data *EmailCodeData) error
	GetEmailCode(ctx context.Context, email string) (*EmailCodeData, error)
	DeleteEmailCode(ctx context.Context, email string) error
	IncrementEmailCodeAttempts(ctx context.Context, email string) error

	// Rate limiting
	SetRateLimit(ctx context.Context, email string, ttl time.Duration) error
	CheckRateLimit(ctx context.Context, email string) (bool, error)
}

// RedisService Redis 缓存服务实现
type RedisService struct {
	client *redis.Client
}

// NewRedisService 创建新的 Redis 缓存服务
func NewRedisService(client *redis.Client) *RedisService {
	return &RedisService{
		client: client,
	}
}

// Captcha operations

func (s *RedisService) SetCaptcha(ctx context.Context, captchaID, code string) error {
	key := fmt.Sprintf(captchaKeyFormat, captchaID)
	return s.client.Set(ctx, key, code, 5*time.Minute).Err()
}

func (s *RedisService) GetCaptcha(ctx context.Context, captchaID string) (string, error) {
	key := fmt.Sprintf(captchaKeyFormat, captchaID)
	return s.client.Get(ctx, key).Result()
}

func (s *RedisService) DeleteCaptcha(ctx context.Context, captchaID string) error {
	key := fmt.Sprintf(captchaKeyFormat, captchaID)
	return s.client.Del(ctx, key).Err()
}

// Email code operations

func (s *RedisService) SetEmailCode(ctx context.Context, email string, data *EmailCodeData) error {
	key := fmt.Sprintf(emailCodeKeyFormat, email)
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal email code data: %w", err)
	}
	return s.client.Set(ctx, key, jsonData, 10*time.Minute).Err()
}

func (s *RedisService) GetEmailCode(ctx context.Context, email string) (*EmailCodeData, error) {
	key := fmt.Sprintf(emailCodeKeyFormat, email)
	jsonData, err := s.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var data EmailCodeData
	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal email code data: %w", err)
	}
	return &data, nil
}

func (s *RedisService) DeleteEmailCode(ctx context.Context, email string) error {
	key := fmt.Sprintf(emailCodeKeyFormat, email)
	return s.client.Del(ctx, key).Err()
}

func (s *RedisService) IncrementEmailCodeAttempts(ctx context.Context, email string) error {
	// 获取当前数据
	data, err := s.GetEmailCode(ctx, email)
	if err != nil {
		return err
	}

	// 增加尝试次数
	data.Attempts++

	// 保存回 Redis (保持原有的 TTL)
	key := fmt.Sprintf(emailCodeKeyFormat, email)
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal email code data: %w", err)
	}

	// 获取剩余 TTL
	ttl, err := s.client.TTL(ctx, key).Result()
	if err != nil {
		return err
	}

	return s.client.Set(ctx, key, jsonData, ttl).Err()
}

// Rate limiting operations

func (s *RedisService) SetRateLimit(ctx context.Context, email string, ttl time.Duration) error {
	key := fmt.Sprintf(rateLimitKeyFormat, email)
	// 使用 SetNX 确保只有不存在时才设置
	return s.client.SetNX(ctx, key, "1", ttl).Err()
}

func (s *RedisService) CheckRateLimit(ctx context.Context, email string) (bool, error) {
	key := fmt.Sprintf(rateLimitKeyFormat, email)
	exists, err := s.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}
