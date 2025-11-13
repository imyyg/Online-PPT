# Capability: Email Verification

## Purpose

邮箱验证能力,提供验证码生成、发送和验证功能,用于确认用户邮箱所有权。
## Requirements
### Requirement: 图形验证码生成

系统 SHALL 提供图形验证码生成功能,用于防止自动化攻击。

#### Scenario: 生成验证码

- **GIVEN** 客户端请求图形验证码
- **WHEN** 系统收到请求
- **THEN** 系统生成唯一的验证码ID和图片
- **AND** 将验证码存储在缓存中(有效期5分钟)
- **AND** 返回验证码ID和Base64编码的图片

### Requirement: 邮箱验证码发送

系统 SHALL 支持发送邮箱验证码,确保只有通过图形验证码验证的请求才能发送。

#### Scenario: 成功发送验证码

- **GIVEN** 用户提供有效的邮箱和正确的图形验证码
- **WHEN** 用户请求发送邮箱验证码
- **THEN** 系统生成6位数字验证码
- **AND** 通过SMTP发送验证码邮件
- **AND** 将验证码存储在缓存中(有效期10分钟)

#### Scenario: 图形验证码无效

- **GIVEN** 用户提供的图形验证码不正确
- **WHEN** 用户请求发送邮箱验证码
- **THEN** 系统返回400错误
- **AND** 不发送邮件

#### Scenario: 发送频率限制

- **GIVEN** 用户在60秒内已向同一邮箱发送过验证码
- **WHEN** 用户再次请求发送验证码
- **THEN** 系统返回429错误
- **AND** 提示需要等待

### Requirement: 邮箱验证码验证

系统 SHALL 提供验证码验证功能,支持失败次数限制。

#### Scenario: 验证成功

- **GIVEN** 用户提供正确的邮箱验证码
- **WHEN** 系统验证验证码
- **THEN** 系统确认验证码有效
- **AND** 从缓存中删除验证码

#### Scenario: 验证码错误

- **GIVEN** 用户提供错误的验证码
- **WHEN** 系统验证验证码
- **THEN** 系统返回错误
- **AND** 增加失败计数

#### Scenario: 超过最大尝试次数

- **GIVEN** 用户已失败验证5次
- **WHEN** 用户再次尝试验证
- **THEN** 系统返回429错误
- **AND** 提示验证码已失效,需重新获取

### Requirement: 缓存服务

系统 SHALL 使用Redis作为缓存服务,存储验证码和频率限制数据。

#### Scenario: 数据存储

- **GIVEN** 系统需要存储验证码或限制数据
- **WHEN** 系统写入缓存
- **THEN** 数据存储在Redis中
- **AND** 设置适当的过期时间

#### Scenario: 缓存服务不可用

- **GIVEN** Redis服务不可用
- **WHEN** 系统尝试访问缓存
- **THEN** 系统返回503错误
- **AND** 记录错误日志

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

