package mail

import (
	"fmt"

	"gopkg.in/gomail.v2"
)

// Service 邮件服务接口
type Service interface {
	SendVerificationCode(to, code string) error
}

// SMTPService SMTP 邮件服务实现
type SMTPService struct {
	host     string
	port     int
	username string
	password string
	from     string
	fromName string
}

// Config SMTP 配置
type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	FromName string
}

// NewSMTPService 创建新的 SMTP 邮件服务
func NewSMTPService(cfg Config) *SMTPService {
	return &SMTPService{
		host:     cfg.Host,
		port:     cfg.Port,
		username: cfg.Username,
		password: cfg.Password,
		from:     cfg.From,
		fromName: cfg.FromName,
	}
}

// SendVerificationCode 发送验证码邮件
func (s *SMTPService) SendVerificationCode(to, code string) error {
	m := gomail.NewMessage()

	// 设置发件人
	m.SetHeader("From", m.FormatAddress(s.from, s.fromName))

	// 设置收件人
	m.SetHeader("To", to)

	// 设置主题
	m.SetHeader("Subject", "邮箱验证码 - Online PPT")

	// 设置邮件内容
	body := renderVerificationCodeTemplate(code)
	m.SetBody("text/html", body)

	// 创建拨号器
	d := gomail.NewDialer(s.host, s.port, s.username, s.password)

	// 发送邮件
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// renderVerificationCodeTemplate 渲染验证码邮件模板
func renderVerificationCodeTemplate(code string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body {
            font-family: Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 600px;
            margin: 0 auto;
            padding: 20px;
        }
        .container {
            background: #f9f9f9;
            border-radius: 10px;
            padding: 30px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .header {
            text-align: center;
            margin-bottom: 30px;
        }
        .code {
            background: #ffffff;
            border: 2px solid #4CAF50;
            border-radius: 8px;
            font-size: 32px;
            font-weight: bold;
            color: #4CAF50;
            text-align: center;
            padding: 20px;
            margin: 20px 0;
            letter-spacing: 8px;
        }
        .notice {
            color: #666;
            font-size: 14px;
            margin-top: 20px;
            padding-top: 20px;
            border-top: 1px solid #ddd;
        }
        .footer {
            text-align: center;
            color: #999;
            font-size: 12px;
            margin-top: 30px;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h2>Online PPT 邮箱验证</h2>
        </div>
        <p>您好，</p>
        <p>您正在注册 Online PPT 账号，请使用以下验证码完成注册：</p>
        <div class="code">%s</div>
        <div class="notice">
            <p><strong>重要提示：</strong></p>
            <ul>
                <li>验证码有效期为 <strong>10分钟</strong></li>
                <li>请勿将验证码告知他人</li>
                <li>如果这不是您的操作，请忽略此邮件</li>
            </ul>
        </div>
        <div class="footer">
            <p>此邮件由系统自动发送，请勿直接回复</p>
            <p>&copy; 2025 Online PPT. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`, code)
}
