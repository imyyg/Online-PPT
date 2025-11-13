# Implementation Tasks

本文档列出实施邮箱验证码功能的所有任务，按照执行顺序排列。每个任务都应该是可独立验证的小步骤。

---

## Phase 1: 数据库和配置 (Foundation)

### 1.1 创建数据库迁移脚本

- [ ] 创建 `migrations/003_create_email_verification_codes.sql`
- [ ] 定义 `email_verification_codes` 表结构
- [ ] 添加必要的索引（email+is_used+created_at, expires_at）
- [ ] 在迁移脚本中添加注释说明

**验证**: 运行迁移脚本，检查表是否正确创建，索引是否存在

**文件**: `online-ppt/migrations/003_create_email_verification_codes.sql`

---

### 1.2 添加 SMTP 配置支持

- [ ] 在 `internal/config/config.go` 中定义 `SMTPConfig` 结构体
- [ ] 添加字段：Host, Port, Username, Password, From, FromName, UseTLS
- [ ] 在 `Config` 结构体中添加 `SMTP SMTPConfig` 字段
- [ ] 更新配置加载逻辑

**验证**: 编译通过，配置结构体可正常序列化/反序列化

**文件**: `online-ppt/internal/config/config.go`

---

### 1.3 更新配置示例文件

- [ ] 在 `configs/app.yaml` 中添加 smtp 配置块
- [ ] 使用占位符或示例值
- [ ] 添加配置注释说明各字段用途

**验证**: YAML 格式正确，可被解析

**文件**: `online-ppt/configs/app.yaml`

---

### 1.4 创建 SMTP 配置指导文档

- [ ] 创建 `docs/smtp-setup.md` 文档
- [ ] 列出常见邮件服务商配置示例（Gmail, QQ, 163, 阿里云）
- [ ] 说明如何获取 SMTP 凭证（应用专用密码、授权码等）
- [ ] 添加常见问题和故障排查步骤

**验证**: 文档可读性良好，示例配置完整

**文件**: `online-ppt/docs/smtp-setup.md`

---

## Phase 2: 邮件服务 (Mail Service)

### 2.1 创建邮件服务包结构

- [ ] 创建 `internal/mail/` 目录
- [ ] 创建 `sender.go` 文件定义 SMTP 发送器接口和实现
- [ ] 创建 `templates.go` 文件定义邮件模板

**验证**: 目录结构正确，文件可编译

**文件**: 
- `online-ppt/internal/mail/sender.go`
- `online-ppt/internal/mail/templates.go`

---

### 2.2 实现 SMTP 发送器

- [ ] 定义 `Sender` 接口，包含 `SendMail(to, subject, body string) error` 方法
- [ ] 实现 `SMTPSender` 结构体，使用 `net/smtp` 包
- [ ] 实现 TLS 连接支持
- [ ] 实现邮件头设置（From, To, Subject, Content-Type）
- [ ] 添加错误处理和重试逻辑（可选）

**验证**: 
- 单元测试验证邮件构造逻辑
- 使用真实 SMTP 服务器测试发送（可选）

**文件**: `online-ppt/internal/mail/sender.go`

---

### 2.3 创建邮件模板

- [ ] 定义验证码邮件 HTML 模板
- [ ] 使用 `html/template` 包实现模板渲染
- [ ] 模板变量：Code（验证码），ExpiresIn（有效期）
- [ ] 确保模板兼容主流邮件客户端（Gmail, Outlook等）

**验证**: 
- 单元测试验证模板渲染输出
- 手动检查渲染后的 HTML 格式

**文件**: `online-ppt/internal/mail/templates.go`

---

### 2.4 编写邮件服务测试

- [ ] 创建 `sender_test.go` 测试邮件构造逻辑
- [ ] 创建 `templates_test.go` 测试模板渲染
- [ ] 使用 mock SMTP 服务器进行集成测试（可选）

**验证**: 所有测试通过

**文件**: 
- `online-ppt/internal/mail/sender_test.go`
- `online-ppt/internal/mail/templates_test.go`

---

## Phase 3: 验证码服务 (Verification Service)

### 3.1 添加验证码数据模型

- [ ] 在 `internal/auth/models.go` 中定义 `VerificationCode` 结构体
- [ ] 字段：ID, Email, Code, CreatedAt, ExpiresAt, VerifiedAt, AttemptCount, IsUsed
- [ ] 实现 `Scan` 方法用于数据库行扫描

**验证**: 编译通过，结构体定义完整

**文件**: `online-ppt/internal/auth/models.go`

---

### 3.2 实现验证码生成器

- [ ] 在 `internal/auth/` 中创建 `verification.go` 文件
- [ ] 实现 `generateVerificationCode() (string, error)` 函数
- [ ] 使用 `crypto/rand` 生成6位随机数字
- [ ] 添加单元测试验证格式和随机性

**验证**: 
- 生成的验证码长度为6
- 所有字符均为数字
- 多次调用生成不同验证码

**文件**: `online-ppt/internal/auth/verification.go`

---

### 3.3 添加验证码 Repository 方法

- [ ] 在 `Repository` 中添加 `CreateVerificationCode(ctx, email, code, expiresAt) error`
- [ ] 在 `Repository` 中添加 `GetActiveVerificationCode(ctx, email) (*VerificationCode, error)`
- [ ] 在 `Repository` 中添加 `MarkCodeAsUsed(ctx, id) error`
- [ ] 在 `Repository` 中添加 `IncrementAttemptCount(ctx, id) error`
- [ ] 在 `Repository` 中添加 `MarkOldCodesAsUsed(ctx, email) error`
- [ ] 实现事务支持（创建新验证码时自动标记旧的为已使用）

**验证**: 
- 单元测试验证 SQL 语句正确性
- 集成测试验证数据库操作

**文件**: `online-ppt/internal/auth/repository.go`

---

### 3.4 实现验证码发送逻辑

- [ ] 在 `Service` 中添加 `mailSender mail.Sender` 字段
- [ ] 实现 `SendVerificationCode(ctx, email) error` 方法
- [ ] 检查邮箱格式
- [ ] 检查发送频率（60秒限制）
- [ ] 生成验证码
- [ ] 计算过期时间（当前时间 + 10分钟）
- [ ] 存储到数据库（同时标记旧验证码为已使用）
- [ ] 发送邮件
- [ ] 添加审计日志

**验证**:
- 单元测试验证业务逻辑
- Mock 邮件发送器测试
- 频率限制测试

**文件**: `online-ppt/internal/auth/service.go`

---

### 3.5 实现验证码验证逻辑

- [ ] 实现 `VerifyEmail(ctx, email, code) error` 方法
- [ ] 查询有效验证码
- [ ] 检查验证码是否存在
- [ ] 检查是否过期
- [ ] 检查尝试次数（最多5次）
- [ ] 使用常量时间比较验证码（`subtle.ConstantTimeCompare`）
- [ ] 验证失败时增加尝试次数
- [ ] 验证成功时标记为已使用并更新用户状态为 active
- [ ] 添加审计日志

**验证**:
- 单元测试覆盖所有场景（成功、失败、过期、超限）
- 测试并发验证场景

**文件**: `online-ppt/internal/auth/service.go`

---

### 3.6 定义验证码相关错误类型

- [ ] 在 `service.go` 中定义新错误：
  - `ErrCodeExpired`: 验证码已过期
  - `ErrCodeInvalid`: 验证码错误
  - `ErrCodeMaxAttempts`: 超过最大尝试次数
  - `ErrEmailAlreadyVerified`: 邮箱已验证
  - `ErrRateLimited`: 发送频率受限
  - `ErrMailSendFailed`: 邮件发送失败

**验证**: 编译通过，错误定义清晰

**文件**: `online-ppt/internal/auth/service.go`

---

### 3.7 编写验证码服务测试

- [ ] 创建 `verification_test.go` 测试验证码生成
- [ ] 测试验证码发送逻辑（mock 邮件和数据库）
- [ ] 测试验证码验证逻辑（所有分支）
- [ ] 测试频率限制
- [ ] 测试并发场景

**验证**: 所有测试通过，覆盖率 > 80%

**文件**: `online-ppt/internal/auth/verification_test.go`

---

## Phase 4: 修改注册登录流程

### 4.1 修改注册流程

- [ ] 修改 `Service.Register` 方法，创建用户时状态设为 pending
- [ ] 移除注册后自动登录逻辑
- [ ] 注册成功后自动调用 `SendVerificationCode`
- [ ] 更新返回值（不包含令牌）

**验证**: 
- 单元测试验证注册创建 pending 用户
- 验证不返回令牌

**文件**: `online-ppt/internal/auth/service.go`

---

### 4.2 修改登录流程

- [ ] 修改 `Service.Login` 方法，检查用户状态
- [ ] pending 状态返回 `ErrAccountPending` 错误
- [ ] locked 状态返回 `ErrAccountLocked` 错误
- [ ] active 状态正常处理

**验证**: 
- 单元测试验证状态检查逻辑
- 各状态用户的登录行为

**文件**: `online-ppt/internal/auth/service.go`

---

### 4.3 定义账号状态错误

- [ ] 添加 `ErrAccountPending` 错误
- [ ] 添加 `ErrAccountLocked` 错误

**验证**: 编译通过

**文件**: `online-ppt/internal/auth/service.go`

---

## Phase 5: HTTP 接口 (API Endpoints)

### 5.1 添加发送验证码端点

- [ ] 在 `AuthHandler` 中实现 `SendVerificationCode(c *gin.Context)` 方法
- [ ] 解析请求 JSON：`{ "email": "..." }`
- [ ] 调用 `service.SendVerificationCode`
- [ ] 处理错误映射（rate_limited, mail_send_failed 等）
- [ ] 返回成功响应：`{ "message": "...", "expires_in": 600 }`

**验证**: 
- 使用 curl 或 Postman 测试端点
- 验证错误响应格式

**文件**: `online-ppt/internal/http/handlers/auth_handlers.go`

---

### 5.2 添加验证邮箱端点

- [ ] 在 `AuthHandler` 中实现 `VerifyEmail(c *gin.Context)` 方法
- [ ] 解析请求 JSON：`{ "email": "...", "code": "..." }`
- [ ] 调用 `service.VerifyEmail`
- [ ] 处理错误映射（invalid_code, code_expired, max_attempts 等）
- [ ] 返回成功响应：`{ "message": "...", "verified": true }`

**验证**: 
- 使用 curl 或 Postman 测试端点
- 验证所有错误场景

**文件**: `online-ppt/internal/http/handlers/auth_handlers.go`

---

### 5.3 更新注册端点响应

- [ ] 修改 `Register` 方法的响应格式
- [ ] 移除令牌相关字段
- [ ] 添加提示消息

**验证**: 
- 测试注册流程返回正确格式
- 不包含 access_token 和 refresh_token

**文件**: `online-ppt/internal/http/handlers/auth_handlers.go`

---

### 5.4 更新登录端点错误处理

- [ ] 在 `handleAuthError` 中添加 pending 和 locked 状态处理
- [ ] pending: 返回 403 + "account_pending"
- [ ] locked: 返回 403 + "account_locked"

**验证**: 
- 测试 pending 用户登录返回 403
- 测试 locked 用户登录返回 403

**文件**: `online-ppt/internal/http/handlers/auth_handlers.go`

---

### 5.5 注册路由

- [ ] 在 `RegisterAuthRoutes` 中添加新路由：
  - `POST /api/v1/auth/send-verification-code`
  - `POST /api/v1/auth/verify-email`

**验证**: 
- 检查路由表，确认端点已注册
- 测试端点可访问

**文件**: `online-ppt/internal/http/router.go`

---

## Phase 6: 集成和测试

### 6.1 更新 main.go 初始化逻辑

- [ ] 在 `cmd/server/main.go` 中初始化邮件发送器
- [ ] 将邮件发送器传递给 Auth Service
- [ ] 确保配置正确加载

**验证**: 
- 应用启动成功
- 配置正确读取

**文件**: `online-ppt/cmd/server/main.go`

---

### 6.2 编写端到端集成测试

- [ ] 创建 `tests/integration/email_verification_test.go`
- [ ] 测试场景1：注册 → 发送验证码 → 验证邮箱 → 登录
- [ ] 测试场景2：频率限制验证
- [ ] 测试场景3：验证码过期处理
- [ ] 测试场景4：错误验证次数限制
- [ ] 使用 mock 邮件发送器（避免真实邮件）

**验证**: 所有集成测试通过

**文件**: `online-ppt/tests/integration/email_verification_test.go`

---

### 6.3 更新现有集成测试

- [ ] 修改 `auth_flow_test.go` 中的注册流程测试
- [ ] 添加邮箱验证步骤
- [ ] 更新断言（检查 pending 状态）

**验证**: 所有现有测试通过

**文件**: `online-ppt/tests/integration/auth_flow_test.go`

---

### 6.4 手动测试完整流程

- [ ] 配置真实 SMTP 服务器（如 Gmail）
- [ ] 执行完整流程：注册 → 收邮件 → 验证 → 登录
- [ ] 验证邮件送达和内容正确
- [ ] 测试频率限制（60秒内重复发送）
- [ ] 测试验证码过期（等待10分钟后验证）
- [ ] 测试错误验证（连续输入错误验证码5次）

**验证**: 用户体验流畅，所有功能正常

---

## Phase 7: 数据迁移和清理

### 7.1 创建数据迁移脚本

- [ ] 创建 SQL 脚本自动激活现有 pending 用户
- [ ] 脚本仅影响在功能上线前创建的用户
- [ ] 添加备份和回滚说明

**验证**: 
- 在测试环境执行迁移
- 验证现有用户可正常登录

**文件**: `online-ppt/migrations/004_activate_existing_users.sql`

---

### 7.2 实现验证码清理任务

- [ ] 创建 `internal/tasks/cleanup.go` 文件
- [ ] 实现 `CleanupExpiredCodes(ctx) error` 函数
- [ ] 删除创建时间超过7天的验证码记录
- [ ] 记录清理日志

**验证**: 
- 单元测试验证清理逻辑
- 手动执行验证记录被删除

**文件**: `online-ppt/internal/tasks/cleanup.go`

---

### 7.3 添加定时清理任务（可选）

- [ ] 在 main.go 中启动后台 goroutine
- [ ] 每天执行一次清理任务
- [ ] 使用 context 支持优雅关闭

**验证**: 
- 应用运行24小时后检查日志
- 验证过期记录被清理

**文件**: `online-ppt/cmd/server/main.go`

---

## Phase 8: 文档和交付

### 8.1 更新 API 文档

- [ ] 更新 `specs/001-add-go-backend/contracts/api.yaml`
- [ ] 添加新端点文档（send-verification-code, verify-email）
- [ ] 更新注册和登录端点说明
- [ ] 添加错误响应文档

**验证**: OpenAPI 文档格式正确，可渲染

**文件**: `online-ppt/specs/001-add-go-backend/contracts/api.yaml`

---

### 8.2 编写用户指南

- [ ] 创建 `docs/email-verification-guide.md`
- [ ] 说明用户如何注册和验证邮箱
- [ ] 常见问题解答（收不到邮件、验证码过期等）

**验证**: 文档可读性良好

**文件**: `online-ppt/docs/email-verification-guide.md`

---

### 8.3 编写部署指南

- [ ] 更新 `README.md` 添加 SMTP 配置说明
- [ ] 添加环境变量配置示例
- [ ] 说明如何测试邮件发送功能

**验证**: 按文档可成功部署

**文件**: `online-ppt/README.md`

---

### 8.4 性能测试

- [ ] 使用负载测试工具（如 hey, wrk）测试验证码发送端点
- [ ] 测试 QPS 和响应时间
- [ ] 验证数据库索引效率
- [ ] 检查邮件发送是否成为瓶颈

**验证**: 
- 发送验证码 QPS > 100
- 验证邮箱 QPS > 200
- 响应时间 P95 < 500ms

---

### 8.5 安全审计

- [ ] 检查验证码生成使用 crypto/rand
- [ ] 验证使用常量时间比较
- [ ] 确认频率限制生效
- [ ] 检查敏感信息不记录在日志中
- [ ] 验证 SQL 注入防护

**验证**: 安全检查清单全部通过

---

## Dependencies

- **1.2 依赖 1.1**: 配置定义需要数据库表已创建
- **2.2 依赖 2.1**: SMTP 发送器实现需要包结构存在
- **3.4 依赖 2.2, 3.2, 3.3**: 发送逻辑需要邮件服务和数据访问
- **4.1 依赖 3.4**: 注册流程修改需要验证码发送功能
- **5.1, 5.2 依赖 3.4, 3.5**: HTTP 端点需要服务层方法
- **6.1 依赖所有前置**: 集成需要所有组件完成
- **7.1 依赖 6.2**: 数据迁移需要功能已测试完成

## Parallel Work Opportunities

可以并行进行的任务组：

**Group A (数据层)**:
- 1.1 数据库迁移
- 1.2-1.4 配置和文档

**Group B (邮件服务)**:
- 2.1-2.4 完整邮件服务实现

**Group C (验证码服务)**:
- 3.1-3.2 基础模型和生成器

完成 Group A + B + C 后可以并行：

**Group D**:
- 3.3-3.7 验证码服务逻辑

**Group E**:
- 4.1-4.3 注册登录修改

完成 D + E 后：

**Group F**:
- 5.1-5.5 HTTP 接口

最后串行：
- 6.1-6.4 集成测试
- 7.1-7.3 数据迁移和清理
- 8.1-8.5 文档和交付

---

## Completion Criteria

所有任务完成后，系统应满足以下条件：

✅ 用户可以注册账号（状态为 pending）
✅ 注册后自动发送验证码邮件
✅ 用户可以通过验证码激活账号
✅ 仅 active 状态用户可以登录
✅ 验证码10分钟后过期
✅ 同一邮箱60秒内只能发送一次验证码
✅ 验证失败最多尝试5次
✅ 所有单元测试和集成测试通过
✅ API 文档完整且准确
✅ 部署指南清晰可用
