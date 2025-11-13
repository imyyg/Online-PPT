package captcha

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/dchest/captcha"
	"github.com/google/uuid"
)

// Service 图形验证码服务接口
type Service interface {
	Generate(ctx context.Context) (captchaID, imageBase64 string, err error)
	Verify(ctx context.Context, captchaID, code string) (bool, error)
}

// CacheService 缓存服务接口
type CacheService interface {
	SetCaptcha(ctx context.Context, captchaID, code string) error
	GetCaptcha(ctx context.Context, captchaID string) (string, error)
	DeleteCaptcha(ctx context.Context, captchaID string) error
}

// service 图形验证码服务实现
type service struct {
	cache CacheService
}

// NewService 创建新的图形验证码服务
func NewService(cache CacheService) Service {
	return &service{
		cache: cache,
	}
}

// Generate 生成图形验证码
func (s *service) Generate(ctx context.Context) (captchaID, imageBase64 string, err error) {
	// 生成唯一 ID
	captchaID = uuid.New().String()

	// 生成验证码 ID (captcha 库内部使用)
	captchaData := captcha.New()

	// 获取验证码数字
	digits := captcha.RandomDigits(6)

	// 将数字转换为字符串用于存储
	var code strings.Builder
	for _, d := range digits {
		code.WriteByte(byte('0' + d))
	}

	// 生成图片
	var img bytes.Buffer
	if err := captcha.WriteImage(&img, captchaData, 120, 40); err != nil {
		return "", "", fmt.Errorf("failed to generate captcha image: %w", err)
	}

	// 编码为 Base64
	imageBase64 = "data:image/png;base64," + base64.StdEncoding.EncodeToString(img.Bytes())

	// 存储到缓存
	if err := s.cache.SetCaptcha(ctx, captchaID, code.String()); err != nil {
		return "", "", fmt.Errorf("failed to cache captcha: %w", err)
	}

	return captchaID, imageBase64, nil
}

// Verify 验证图形验证码
func (s *service) Verify(ctx context.Context, captchaID, code string) (bool, error) {
	// 从缓存获取
	cachedCode, err := s.cache.GetCaptcha(ctx, captchaID)
	if err != nil {
		return false, nil // 验证码不存在或已过期
	}

	// 验证码只能使用一次，立即删除
	_ = s.cache.DeleteCaptcha(ctx, captchaID)

	// 比较验证码
	return cachedCode == code, nil
}
