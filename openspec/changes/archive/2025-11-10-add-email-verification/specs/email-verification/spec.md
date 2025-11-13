# Capability: Email Verification (简化版)

邮箱验证能力，提供图形验证码和邮箱验证码生成、缓存和验证功能。

## ADDED Requirements

### Requirement: 生成和验证图形验证码

系统 SHALL 提供图形验证码（CAPTCHA）生成和验证功能，防止机器人攻击。

**Rationale**: 图形验证码是防止自动化注册的第一道防线。

#### Scenario: 获取图形验证码

- **GIVEN** 用户访问注册页面
- **WHEN** 调用 `GET /api/v1/auth/captcha`
- **THEN** 生成唯一的 captcha_id (UUID)
- **AND** 生成4-6位随机字符串（字母数字混合）
- **AND** 渲染为PNG图片（带干扰线和噪点）
- **AND** 缓存 captcha_id -> code 映射（5分钟TTL）
- **AND** 返回 `{captcha_id, image_base64, expires_in}`

#### Scenario: 验证图形验证码（发送邮箱验证码时）

- **GIVEN** 用户提交 captcha_id 和 captcha_code
- **WHEN** 调用发送邮箱验证码接口
- **THEN** 从缓存查询 captcha_id 对应的 code
- **AND** 比对验证码（不区分大小写）
- **AND** 验证成功后从缓存删除（一次性使用）
- **AND** 验证失败返回 `invalid_captcha` 错误

---

### Requirement: 发送邮箱验证码

系统 SHALL 在验证图形验证码后，生成并发送邮箱验证码。

**Rationale**: 邮箱验证码用于验证用户邮箱所有权，MUST 先通过图形验证码防止滥用。

#### Scenario: 成功发送邮箱验证码

- **GIVEN** 用户提供合法邮箱和正确的图形验证码
- **AND** 距离上次发送已超过60秒
- **WHEN** 调用 `POST /api/v1/auth/send-verification-code`
- **THEN** 生成6位随机数字码（000000-999999）
- **AND** 缓存 email -> {code, attempts:0, created_at}（10分钟TTL）
- **AND** 通过 SMTP 发送邮件
- **AND** 缓存发送时间用于频率限制（60秒TTL）
- **AND** 返回 `{message, expires_in:600}`

#### Scenario: 图形验证码错误

- **GIVEN** 用户提供的图形验证码不正确
- **WHEN** 调用发送邮箱验证码接口
- **THEN** 返回 HTTP 400
- **AND** 错误代码为 `invalid_captcha`
- **AND** 不发送邮件，不生成邮箱验证码

#### Scenario: 发送频率受限

- **GIVEN** 用户在60秒内已发送过验证码
- **WHEN** 再次请求发送
- **THEN** 返回 HTTP 429
- **AND** 错误代码为 `rate_limited`
- **AND** 响应包含 `retry_after` 秒数

---

### Requirement: 注册时验证邮箱验证码

系统 SHALL 在用户注册时验证邮箱验证码，验证通过后创建账号。

**Rationale**: 确保用户拥有邮箱访问权限，防止使用虚假邮箱注册。

#### Scenario: 验证码正确且未过期

- **GIVEN** 用户提交邮箱、密码和正确的邮箱验证码
- **AND** 验证码在有效期内（10分钟）
- **AND** 验证尝试次数少于5次
- **WHEN** 调用 `POST /api/v1/auth/register`
- **THEN** 验证成功
- **AND** 从缓存删除验证码
- **AND** 检查邮箱是否已注册
- **AND** 创建用户账号（状态为 active）
- **AND** 返回用户信息和访问令牌

#### Scenario: 验证码错误

- **GIVEN** 用户提交错误的邮箱验证码
- **WHEN** 调用注册接口
- **THEN** 验证失败
- **AND** 尝试次数加1
- **AND** 返回 HTTP 400 + `invalid_code` 错误
- **AND** 响应包含剩余尝试次数

#### Scenario: 验证码已过期

- **GIVEN** 验证码创建时间超过10分钟
- **WHEN** 用户提交该验证码
- **THEN** 验证失败
- **AND** 返回 HTTP 400 + `code_expired` 错误
- **AND** 提示用户重新获取验证码

#### Scenario: 超过最大尝试次数

- **GIVEN** 用户已尝试验证5次
- **WHEN** 再次提交验证码
- **THEN** 验证失败
- **AND** 从缓存删除验证码
- **AND** 返回 HTTP 429 + `max_attempts` 错误

---

### Requirement: 防止验证码滥用

系统 SHALL 实施多层防护措施，防止恶意用户滥用验证码功能。

**Rationale**: 保护邮件服务资源，防止骚扰和攻击。

#### Scenario: 多层防护生效

- **第一层**: 图形验证码（防止自动化脚本）
- **第二层**: 发送频率限制（60秒/邮箱）
- **第三层**: 验证尝试限制（5次/验证码）
- **第四层**: 验证码自动过期（10分钟TTL）

---

## API Specifications

### GET /api/v1/auth/captcha

获取图形验证码

**Response (200):**
```json
{
  "captcha_id": "uuid",
  "image": "data:image/png;base64,...",
  "expires_in": 300
}
```

---

### POST /api/v1/auth/send-verification-code

发送邮箱验证码（需先验证图形验证码）

**Request:**
```json
{
  "email": "user@example.com",
  "captcha_id": "uuid",
  "captcha_code": "AB12"
}
```

**Success (200):**
```json
{
  "message": "验证码已发送",
  "expires_in": 600
}
```

**Errors:**
- 400 `invalid_email`: 邮箱格式错误
- 400 `invalid_captcha`: 图形验证码错误或已过期
- 429 `rate_limited`: 发送频率受限
- 500 `mail_send_failed`: 邮件发送失败

---

### POST /api/v1/auth/register

注册账号（需验证邮箱验证码）

**Request:**
```json
{
  "email": "user@example.com",
  "password": "SecurePass123!",
  "email_code": "123456"
}
```

**Success (201):**
```json
{
  "user": {
    "uuid": "uuid",
    "email": "user@example.com",
    "status": "active",
    "created_at": "2025-11-10T10:00:00Z"
  },
  "access_token": "jwt_token",
  "expires_at": "2025-11-10T10:15:00Z"
}
```

**Errors:**
- 400 `invalid_code`: 邮箱验证码错误
- 400 `code_expired`: 验证码已过期
- 409 `email_exists`: 邮箱已注册
- 429 `max_attempts`: 超过最大尝试次数

---

## Configuration

### SMTP 配置

在 `configs/app.yaml` 中添加：

```yaml
smtp:
  host: smtp.gmail.com
  port: 587
  username: your-email@gmail.com
  password: your-app-password
  from: noreply@yourdomain.com
  from_name: Online PPT
  use_tls: true
```

---

## Implementation Notes

### Redis 缓存实现

使用 `github.com/redis/go-redis/v9` 客户端库：

```go
import (
    "github.com/redis/go-redis/v9"
)

// 初始化 Redis 客户端
func NewRedisClient(cfg *RedisConfig) *redis.Client {
    return redis.NewClient(&redis.Options{
        Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
        Password: cfg.Password,
        DB:       cfg.DB,
        PoolSize: cfg.PoolSize,
    })
}

// 缓存服务实现
type RedisCacheService struct {
    client *redis.Client
}

func (s *RedisCacheService) SetCaptcha(ctx context.Context, captchaID, code string) error {
    return s.client.Set(ctx, "captcha:"+captchaID, code, 5*time.Minute).Err()
}

func (s *RedisCacheService) GetCaptcha(ctx context.Context, captchaID string) (string, error) {
    return s.client.Get(ctx, "captcha:"+captchaID).Result()
}
```

### 图形验证码库

推荐使用 `github.com/dchest/captcha` 或自行实现简单的图形验证码生成器。

### 邮件发送

使用 `net/smtp` + `html/template` 实现 SMTP 发送和邮件模板渲染。

### 依赖库

在 `go.mod` 中添加：

```go
require (
    github.com/redis/go-redis/v9 v9.0.5
    github.com/dchest/captcha v1.0.0  // 可选
)
```
