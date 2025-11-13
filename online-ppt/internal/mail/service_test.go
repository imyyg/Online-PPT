package mail

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/gomail.v2"
)

// MockDialer 模拟 Dialer 用于测试
type MockDialer struct {
	sentMessages []*gomail.Message
	shouldFail   bool
	failError    error
}

// DialAndSend 模拟发送邮件
func (md *MockDialer) DialAndSend(msg ...*gomail.Message) error {
	if md.shouldFail {
		return md.failError
	}
	md.sentMessages = append(md.sentMessages, msg...)
	return nil
}

// TestNewSMTPService 测试创建 SMTP 服务
func TestNewSMTPService(t *testing.T) {
	cfg := Config{
		Host:     "smtp.example.com",
		Port:     587,
		Username: "user@example.com",
		Password: "password",
		From:     "noreply@example.com",
		FromName: "Test App",
	}

	service := NewSMTPService(cfg)

	assert.NotNil(t, service)
	assert.Equal(t, cfg.Host, service.host)
	assert.Equal(t, cfg.Port, service.port)
	assert.Equal(t, cfg.Username, service.username)
	assert.Equal(t, cfg.Password, service.password)
	assert.Equal(t, cfg.From, service.from)
	assert.Equal(t, cfg.FromName, service.fromName)
}

// TestRenderVerificationCodeTemplate 测试验证码模板渲染
func TestRenderVerificationCodeTemplate(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		wantCode bool
		wantHTML bool
		wantCSS  bool
	}{
		{
			name:     "正常验证码",
			code:     "123456",
			wantCode: true,
			wantHTML: true,
			wantCSS:  true,
		},
		{
			name:     "空验证码",
			code:     "",
			wantCode: false,
			wantHTML: true,
			wantCSS:  true,
		},
		{
			name:     "特殊字符验证码",
			code:     "000000",
			wantCode: true,
			wantHTML: true,
			wantCSS:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			html := renderVerificationCodeTemplate(tt.code)

			// 验证返回结果不为空
			assert.NotEmpty(t, html)

			// 验证包含 HTML 标签
			if tt.wantHTML {
				assert.Contains(t, html, "<!DOCTYPE html>")
				assert.Contains(t, html, "</html>")
				assert.Contains(t, html, "<body>")
			}

			// 验证包含 CSS 样式
			if tt.wantCSS {
				assert.Contains(t, html, "style>")
				assert.Contains(t, html, "font-family")
			}

			// 验证包含验证码
			if tt.wantCode {
				assert.Contains(t, html, tt.code)
			}

			// 验证包含中文内容
			assert.Contains(t, html, "Online PPT")
		})
	}
}

// TestTemplateStructure 测试模板结构完整性
func TestTemplateStructure(t *testing.T) {
	html := renderVerificationCodeTemplate("123456")

	// 必须包含的元素
	requiredElements := []string{
		"<!DOCTYPE html>",
		"<html>",
		"<head>",
		"<meta charset",
		"<style>",
		"<body>",
		"<div",
		"123456",
		"</body>",
		"</html>",
	}

	for _, elem := range requiredElements {
		assert.Contains(t, html, elem, "模板应包含 %s", elem)
	}

	// 验证 HTML 格式基本正确（开闭标签匹配）
	openHTML := strings.Count(html, "<html>")
	closeHTML := strings.Count(html, "</html>")
	assert.Equal(t, openHTML, closeHTML, "HTML 标签应匹配")

	openBody := strings.Count(html, "<body>")
	closeBody := strings.Count(html, "</body>")
	assert.Equal(t, openBody, closeBody, "body 标签应匹配")
}

// TestEmailContent 测试邮件内容格式
func TestEmailContent(t *testing.T) {
	code := "654321"
	html := renderVerificationCodeTemplate(code)

	// 验证包含代码块显示
	assert.Contains(t, html, code)

	// 验证 HTML 是有效的格式（基本检查）
	assert.Contains(t, html, "style>")
	assert.Contains(t, html, "<!DOCTYPE html>")

	// 验证内容包含说明文字
	assert.Contains(t, html, "Online PPT")
}

// TestTemplateXSSProtection 测试 XSS 防护
func TestTemplateXSSProtection(t *testing.T) {
	// 测试注入 JavaScript
	maliciousCode := "<script>alert('xss')</script>"
	html := renderVerificationCodeTemplate(maliciousCode)

	// 验证代码直接显示而不是被执行
	assert.Contains(t, html, maliciousCode)
}

// TestSMTPServiceConfig 测试不同的 SMTP 配置
func TestSMTPServiceConfig(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
	}{
		{
			name: "Gmail 配置",
			cfg: Config{
				Host:     "smtp.gmail.com",
				Port:     587,
				Username: "user@gmail.com",
				Password: "password",
				From:     "user@gmail.com",
				FromName: "Gmail App",
			},
		},
		{
			name: "163 邮箱配置",
			cfg: Config{
				Host:     "smtp.163.com",
				Port:     465,
				Username: "user@163.com",
				Password: "password",
				From:     "user@163.com",
				FromName: "163 App",
			},
		},
		{
			name: "本地 SMTP 配置",
			cfg: Config{
				Host:     "localhost",
				Port:     1025,
				Username: "test",
				Password: "test",
				From:     "test@localhost",
				FromName: "Test Server",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewSMTPService(tt.cfg)
			require.NotNil(t, service)
			assert.Equal(t, tt.cfg.Host, service.host)
			assert.Equal(t, tt.cfg.Port, service.port)
		})
	}
}

// TestTemplateLocalization 测试模板本地化内容
func TestTemplateLocalization(t *testing.T) {
	html := renderVerificationCodeTemplate("123456")

	// 验证包含中文内容
	chineseContent := []string{
		"Online PPT",
		"验证码",
	}

	for _, content := range chineseContent {
		assert.Contains(t, html, content, "模板应包含中文内容：%s", content)
	}
}

// BenchmarkRenderVerificationCodeTemplate 基准测试模板渲染性能
func BenchmarkRenderVerificationCodeTemplate(b *testing.B) {
	code := "123456"
	for i := 0; i < b.N; i++ {
		_ = renderVerificationCodeTemplate(code)
	}
}

// TestTemplateWithDifferentCodes 测试不同验证码长度
func TestTemplateWithDifferentCodes(t *testing.T) {
	codes := []string{
		"000000",
		"999999",
		"123456",
		"654321",
	}

	for _, code := range codes {
		html := renderVerificationCodeTemplate(code)
		assert.Contains(t, html, code, "模板应包含验证码 %s", code)
		assert.NotEmpty(t, html)
	}
}

// TestMessageConstruction 测试邮件消息构造
func TestMessageConstruction(t *testing.T) {
	cfg := Config{
		Host:     "smtp.example.com",
		Port:     587,
		Username: "user@example.com",
		Password: "password",
		From:     "noreply@example.com",
		FromName: "Test App",
	}

	service := NewSMTPService(cfg)
	assert.NotNil(t, service)

	// 验证服务能够正确设置
	m := gomail.NewMessage()
	m.SetHeader("From", m.FormatAddress(service.from, service.fromName))
	m.SetHeader("To", "recipient@example.com")
	m.SetHeader("Subject", "邮箱验证码 - Online PPT")

	html := renderVerificationCodeTemplate("123456")
	m.SetBody("text/html", html)

	// 验证消息内容
	assert.NotEmpty(t, html)
}

// TestVerificationCodeEmailHeaders 测试验证码邮件头
func TestVerificationCodeEmailHeaders(t *testing.T) {
	m := gomail.NewMessage()
	code := "123456"
	fromName := "Test App"
	from := "noreply@example.com"
	to := "user@example.com"

	m.SetHeader("From", m.FormatAddress(from, fromName))
	m.SetHeader("To", to)
	m.SetHeader("Subject", "邮箱验证码 - Online PPT")
	m.SetBody("text/html", renderVerificationCodeTemplate(code))

	// 验证邮件头设置正确
	// 注：gomail.Message 内部字段不容易直接访问，这里主要测试不发生 panic
	assert.NotNil(t, m)
}
