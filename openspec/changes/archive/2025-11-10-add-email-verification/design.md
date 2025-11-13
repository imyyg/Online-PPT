# Design: 邮箱验证码验证系统

## Overview

实现基于验证码的邮箱验证系统，包含验证码生成、存储、发送和验证四个核心环节。采用 SMTP 协议发送邮件，支持灵活配置多种邮件服务提供商。

## Architecture

### 组件职责

```
┌─────────────────┐      ┌──────────────────┐      ┌─────────────────┐
│  HTTP Handler   │─────▶│  Auth Service    │─────▶│  Mail Service   │
│                 │      │                  │      │                 │
│ - 获取图形验证码  │      │ - 生成验证码      │      │ - SMTP 发送     │
│ - 发送邮箱验证码  │      │ - 验证逻辑        │      │ - 模板渲染      │
│ - 注册（验证码）  │      │                  │      └─────────────────┘
└─────────────────┘      └──────────────────┘
                                  │
                    ┌─────────────┴─────────────┐
                    ▼                           ▼
          ┌─────────────────┐         ┌─────────────────┐
          │ Captcha Service │         │  Cache Service  │
          │                 │         │   (Redis)       │
          │ - 生成图形验证码  │         │                 │
          │ - 验证图形验证码  │         │ - SET/GET/DEL   │
          └─────────────────┘         │ - TTL 过期      │
                                      │ - EXISTS 检查    │
                                      └────────┬─────────┘
                                              │
                                              ▼
                                      ┌─────────────────┐
                                      │  Redis 6.0+     │
                                      │                 │
                                      │ - Key-Value存储 │
                                      │ - 自动过期      │
                                      │ - 原子操作      │
                                      └─────────────────┘
```

### 核心流程

#### 1. 获取图形验证码流程

```
用户请求图形验证码
    │
    ├─ 生成唯一 captcha_id (UUID)
    │
    ├─ 生成随机字符串（4-6位，字母数字混合）
    │
    ├─ 渲染为图片（带干扰线和噪点）
    │
    ├─ 缓存 captcha_id -> code 映射（5分钟过期）
    │
    └─ 返回 { captcha_id, image_base64 }
```

#### 2. 发送邮箱验证码流程

```
用户请求发送邮箱验证码
    │
    ├─ 验证邮箱格式
    │
    ├─ 验证图形验证码（captcha_id + captcha_code）
    │   ├─ 失败：返回错误
    │   └─ 成功：从缓存删除图形验证码（一次性）
    │
    ├─ 检查发送频率（60秒限制，基于邮箱）
    │
    ├─ 生成6位随机数字码
    │
    ├─ 缓存 email -> {code, attempts, created_at}（10分钟过期）
    │
    ├─ 通过 SMTP 发送邮件
    │
    └─ 返回成功响应（不暴露验证码）
```

#### 3. 注册流程（含邮箱验证码验证）

```
用户提交注册信息
    │
    ├─ 验证邮箱、密码格式
    │
    ├─ 从缓存获取邮箱验证码信息
    │
    ├─ 检查验证码是否存在
    │
    ├─ 检查是否过期（created_at + 10分钟）
    │
    ├─ 检查验证次数（最多5次）
    │
    ├─ 比对验证码（常量时间比较）
    │   ├─ 失败：增加尝试次数，返回错误
    │   └─ 成功：继续
    │
    ├─ 检查邮箱是否已注册
    │
    ├─ 哈希密码
    │
    ├─ 创建用户（状态为 active）
    │
    ├─ 从缓存删除验证码
    │
    └─ 返回成功响应（可选：自动登录）
```

## Data Model

### Redis 数据结构

#### 图形验证码

**Key 格式**: `captcha:{captcha_id}`  
**Value**: 验证码字符串（如 "AB12"）  
**TTL**: 300 秒（5分钟）

```redis
SET captcha:550e8400-e29b-41d4-a716-446655440000 "AB12" EX 300
GET captcha:550e8400-e29b-41d4-a716-446655440000
DEL captcha:550e8400-e29b-41d4-a716-446655440000
```

#### 邮箱验证码

**Key 格式**: `email_code:{email}`  
**Value**: JSON 字符串，包含验证码和元数据  
**TTL**: 600 秒（10分钟）

```json
{
  "code": "123456",
  "attempts": 0,
  "created_at": "2025-11-10T10:00:00Z"
}
```

```redis
SET email_code:user@example.com '{"code":"123456","attempts":0,"created_at":"2025-11-10T10:00:00Z"}' EX 600
GET email_code:user@example.com
DEL email_code:user@example.com
```

#### 发送频率限制

**Key 格式**: `rate_limit:{email}`  
**Value**: "1" 或时间戳  
**TTL**: 60 秒

```redis
SET rate_limit:user@example.com "1" EX 60 NX
EXISTS rate_limit:user@example.com
```

### 数据库表

**无需新增表**，所有验证码相关数据存储在 Redis 中。

### Go 数据结构

```go
// 邮箱验证码缓存数据结构
type EmailCodeData struct {
    Code      string    `json:"code"`
    Attempts  int       `json:"attempts"`
    CreatedAt time.Time `json:"created_at"`
}

// Redis 缓存服务接口
type CacheService interface {
    // 图形验证码
    SetCaptcha(ctx context.Context, captchaID, code string) error
    GetCaptcha(ctx context.Context, captchaID string) (string, error)
    DeleteCaptcha(ctx context.Context, captchaID string) error
    
    // 邮箱验证码
    SetEmailCode(ctx context.Context, email string, data *EmailCodeData) error
    GetEmailCode(ctx context.Context, email string) (*EmailCodeData, error)
    IncrementAttempts(ctx context.Context, email string) error
    DeleteEmailCode(ctx context.Context, email string) error
    
    // 频率限制
    CheckRateLimit(ctx context.Context, email string) (bool, error)
    SetRateLimit(ctx context.Context, email string) error
}
```

## Security Considerations

### 1. 验证码生成

```go
// 使用 crypto/rand 生成安全的随机数
func generateVerificationCode() (string, error) {
    max := big.NewInt(1000000)
    n, err := rand.Int(rand.Reader, max)
    if err != nil {
        return "", err
    }
    return fmt.Sprintf("%06d", n.Int64()), nil
}
```

### 2. 防止暴力破解

- 单个邮箱最多尝试5次验证
- 验证码10分钟后自动过期
- 验证成功后立即标记为已使用
- 使用常量时间比较防止时序攻击

### 3. 防止滥用

- 同一邮箱60秒内只能发送一次验证码
- 记录发送日志用于监控异常行为
- 考虑后续添加 IP 级别的频率限制

### 4. 数据安全

- 验证码存储为明文（不敏感，短期有效）
- 邮件内容不包含用户其他敏感信息
- 定期清理过期验证码（建议保留7天用于审计）

## Configuration

### 配置结构

```go
// SMTP 配置
type SMTPConfig struct {
    Host     string `yaml:"host"`
    Port     int    `yaml:"port"`
    Username string `yaml:"username"`
    Password string `yaml:"password"`
    From     string `yaml:"from"`
    FromName string `yaml:"from_name"`
    UseTLS   bool   `yaml:"use_tls"`
}

// Redis 配置
type RedisConfig struct {
    Host     string `yaml:"host"`
    Port     int    `yaml:"port"`
    Password string `yaml:"password"`
    DB       int    `yaml:"db"`
    PoolSize int    `yaml:"pool_size"`
}
```

### 配置示例（configs/app.yaml）

```yaml
# SMTP 邮件服务配置
smtp:
  host: smtp.gmail.com
  port: 587
  username: your-email@gmail.com
  password: your-app-password
  from: noreply@yourdomain.com
  from_name: Online PPT
  use_tls: true

# Redis 缓存配置
redis:
  host: localhost
  port: 6379
  password: ""           # 留空表示无密码
  db: 0                  # 使用数据库 0
  pool_size: 10          # 连接池大小
```

### 常见邮件服务商配置

**Gmail:**
```yaml
host: smtp.gmail.com
port: 587
use_tls: true
# 需要开启"应用专用密码"
```

**QQ 邮箱:**
```yaml
host: smtp.qq.com
port: 587
use_tls: true
# 需要开启 SMTP 服务并获取授权码
```

**163 邮箱:**
```yaml
host: smtp.163.com
port: 465
use_tls: true
# 需要开启 SMTP 服务并设置授权密码
```

**阿里云企业邮箱:**
```yaml
host: smtp.qiye.aliyun.com
port: 465
use_tls: true
```

## Email Template

### 验证码邮件模板

```html
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>邮箱验证</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h2 style="color: #4a5568;">验证您的邮箱地址</h2>
        <p>您好，</p>
        <p>感谢您注册 Online PPT。请使用以下验证码完成邮箱验证：</p>
        <div style="background-color: #f7fafc; border: 2px solid #e2e8f0; border-radius: 4px; padding: 20px; text-align: center; margin: 20px 0;">
            <h1 style="color: #2d3748; margin: 0; font-size: 32px; letter-spacing: 8px;">{{.Code}}</h1>
        </div>
        <p style="color: #718096; font-size: 14px;">
            验证码有效期为 <strong>10 分钟</strong>，请尽快完成验证。
        </p>
        <p style="color: #718096; font-size: 14px;">
            如果这不是您的操作，请忽略此邮件。
        </p>
        <hr style="border: none; border-top: 1px solid #e2e8f0; margin: 30px 0;">
        <p style="color: #a0aec0; font-size: 12px; text-align: center;">
            © 2025 Online PPT. All rights reserved.
        </p>
    </div>
</body>
</html>
```

## API Design

### 1. 获取图形验证码

**Endpoint:** `GET /api/v1/auth/captcha`

**Response (Success):**
```json
{
  "captcha_id": "550e8400-e29b-41d4-a716-446655440000",
  "image": "data:image/png;base64,iVBORw0KGgoAAAANSUh...",
  "expires_in": 300
}
```

### 2. 发送邮箱验证码

**Endpoint:** `POST /api/v1/auth/send-verification-code`

**Request:**
```json
{
  "email": "user@example.com",
  "captcha_id": "550e8400-e29b-41d4-a716-446655440000",
  "captcha_code": "AB12"
}
```

**Response (Success):**
```json
{
  "message": "验证码已发送",
  "expires_in": 600
}
```

**Response (Invalid Captcha):**
```json
{
  "error": "invalid_captcha",
  "message": "图形验证码错误或已过期"
}
```

**Response (Rate Limited):**
```json
{
  "error": "rate_limited",
  "message": "请稍后再试",
  "retry_after": 45
}
```

### 3. 注册（含邮箱验证码验证）

**Endpoint:** `POST /api/v1/auth/register`

**Request:**
```json
{
  "email": "user@example.com",
  "password": "SecurePass123!",
  "email_code": "123456"
}
```

**Response (Success):**
```json
{
  "user": {
    "uuid": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "status": "active",
    "created_at": "2025-11-10T10:00:00Z"
  },
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2025-11-10T10:15:00Z"
}
```

**Response (Invalid Email Code):**
```json
{
  "error": "invalid_code",
  "message": "邮箱验证码错误或已过期",
  "attempts_remaining": 3
}
```

## Redis Setup

### 开发环境快速启动

#### 方式 1: 使用 Docker（推荐）

```bash
# 拉取 Redis 镜像
docker pull redis:7-alpine

# 启动 Redis 容器
docker run -d \
  --name online-ppt-redis \
  -p 6379:6379 \
  redis:7-alpine

# 验证连接
docker exec -it online-ppt-redis redis-cli ping
# 应该返回: PONG
```

#### 方式 2: 使用 Docker Compose

创建 `docker-compose.yml`:

```yaml
version: '3.8'
services:
  redis:
    image: redis:7-alpine
    container_name: online-ppt-redis
    ports:
      - "6379:6379"
    volumes:
      - redis-data:/data
    command: redis-server --appendonly yes
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 3s
      retries: 5

volumes:
  redis-data:
```

启动：
```bash
docker-compose up -d
```

#### 方式 3: 本地安装（WSL/Linux）

```bash
# Ubuntu/Debian
sudo apt update
sudo apt install redis-server

# 启动 Redis
sudo systemctl start redis-server

# 设置开机自启
sudo systemctl enable redis-server

# 验证
redis-cli ping
```

### 生产环境配置建议

```conf
# /etc/redis/redis.conf

# 绑定地址
bind 127.0.0.1

# 端口
port 6379

# 密码认证
requirepass your_strong_password_here

# 最大内存
maxmemory 256mb
maxmemory-policy allkeys-lru

# 持久化（可选）
appendonly yes
appendfilename "appendonly.aof"

# 日志
loglevel notice
logfile /var/log/redis/redis-server.log
```

## Implementation Phases

### Phase 0: 环境准备
1. 安装和配置 Redis（Docker 或本地）
2. 添加 Redis 配置到 `configs/app.yaml`
3. 添加 `github.com/redis/go-redis/v9` 依赖
4. 验证 Redis 连接

### Phase 1: Redis 缓存服务
1. 实现 Redis 连接管理（`internal/cache/redis.go`）
2. 实现缓存服务接口（`internal/cache/service.go`）
3. 编写 Redis 操作单元测试
4. 实现健康检查接口

### Phase 2: 邮件服务
1. 添加 SMTP 配置支持
2. 实现邮件发送服务（internal/mail 包）
3. 创建邮件模板
4. 编写邮件发送单元测试

### Phase 3: 业务逻辑
1. 在 Auth Service 中添加验证码发送逻辑
2. 实现验证码验证逻辑
3. 修改注册流程（创建 pending 状态用户）
4. 修改登录流程（检查 active 状态）

### Phase 4: HTTP 接口
1. 添加发送验证码端点
2. 添加验证邮箱端点
3. 更新错误处理
4. 编写集成测试

### Phase 5: 优化与维护
1. 实现过期验证码清理任务
2. 添加监控和日志
3. 编写配置指导文档
4. 性能测试和优化

## Error Handling

### 错误类型定义

```go
var (
    ErrCodeExpired       = errors.New("verification code expired")
    ErrCodeInvalid       = errors.New("invalid verification code")
    ErrCodeMaxAttempts   = errors.New("max verification attempts exceeded")
    ErrEmailAlreadyVerified = errors.New("email already verified")
    ErrRateLimited       = errors.New("too many requests")
    ErrMailSendFailed    = errors.New("failed to send email")
)
```

### HTTP 错误映射

| 错误类型 | HTTP 状态码 | 错误代码 |
|---------|-----------|---------|
| ErrCodeInvalid | 400 | invalid_code |
| ErrCodeExpired | 400 | code_expired |
| ErrCodeMaxAttempts | 429 | max_attempts |
| ErrRateLimited | 429 | rate_limited |
| ErrMailSendFailed | 500 | mail_send_failed |
| ErrEmailAlreadyVerified | 409 | already_verified |

## Testing Strategy

### 单元测试
- 验证码生成器测试（格式、唯一性）
- 验证逻辑测试（过期、错误次数、匹配）
- 频率限制测试
- 邮件模板渲染测试

### 集成测试
- 完整发送和验证流程
- 注册并验证邮箱流程
- 登录状态检查
- 并发发送测试

### 手动测试清单
- [ ] 使用真实 SMTP 服务器发送邮件
- [ ] 验证邮件到达率和内容正确性
- [ ] 测试不同邮件客户端的显示效果
- [ ] 验证频率限制有效性
- [ ] 测试错误场景的用户体验

## Monitoring & Observability

### 关键指标
- 验证码发送成功率
- 验证码验证成功率
- 平均验证时长
- 邮件发送失败次数
- 频率限制触发次数

### 日志记录
- 验证码发送（邮箱、时间戳）
- 验证码验证尝试（成功/失败、尝试次数）
- SMTP 连接错误
- 频率限制触发事件

## Future Enhancements

1. **多渠道支持**: 添加短信验证码作为备选方案
2. **第三方服务集成**: 支持 SendGrid、阿里云等专业邮件服务
3. **验证码重用**: 允许用户在有效期内重发相同验证码
4. **图形验证码**: 在发送验证码前添加人机验证
5. **邮件队列**: 使用消息队列异步处理邮件发送
6. **模板管理**: 支持动态配置邮件模板和多语言
