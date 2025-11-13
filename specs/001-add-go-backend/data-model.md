# Data Model: 用户登录与PPT记录服务

## Overview
后端服务需要管理用户账户、认证会话以及指向本地 `ppt-framework/presentations/` 目录下用户私有子目录的 PPT 记录。所有实体均与单个用户关联，暂不支持跨用户共享。

## Entities

### UserAccount
- **Description**: 注册用户的核心账户信息。
- **Fields**:
  - `id` (INT AUTO_INCREMENT): 主键。
  - `uuid` (CHAR(36), unique): 为前端与路径命名提供的稳定标识，创建时生成。
  - `email` (VARCHAR(320), unique, lowercase): 采用 RFC 5321 格式校验；唯一索引。
  - `password_hash` (VARCHAR(255)): Argon2id 哈希结果，包含参数前缀。
  - `status` (ENUM: `active`, `locked`, `pending`): 账户状态，默认 `active`。
  - `last_login_at` (DATETIME NULL): 最近一次成功登录时间。
  - `created_at` (DATETIME): 默认 `CURRENT_TIMESTAMP`。
  - `updated_at` (DATETIME): `CURRENT_TIMESTAMP` 且 `ON UPDATE CURRENT_TIMESTAMP`。
- **Validation rules**:
  - 邮箱需通过正则校验并统一小写。
  - 密码在注册阶段需满足最少 10 个字符、包含字母与数字。
  - `uuid` 在创建后不可修改。

### UserSession
- **Description**: 记录 JWT 刷新令牌与客户端上下文，用于会话失效控制。
- **Fields**:
  - `id` (INT AUTO_INCREMENT): 主键。
  - `user_id` (INT, FK → UserAccount.id): 关联用户。
  - `refresh_token_hash` (VARCHAR(255)): 对刷新令牌做哈希存储，防止泄露。
  - `expires_at` (DATETIME): 刷新令牌失效时间。
  - `issued_at` (DATETIME): 签发时间。
  - `client_fingerprint` (VARCHAR(255) NULL): 记录设备/浏览器信息。
  - `revoked_at` (DATETIME NULL): 失效时间。
- **Validation rules**:
  - 刷新令牌长度固定，哈希采用与密码不同的盐以避免复用。
  - `expires_at` > `issued_at`。
  - 每个用户可同时拥有多个会话，但在注销时需按 `id` 或指纹精确撤销。

### PptRecord
- **Description**: 用户捕获的 PPT 目录条目。
- **Fields**:
  - `id` (INT AUTO_INCREMENT): 主键。
  - `user_id` (INT, FK → UserAccount.id): 记录归属。
  - `name` (VARCHAR(120)): 展示名称（允许字母、数字、下划线、中划线）。
  - `title` (VARCHAR(255) NULL): 用户友好的展示标题，支持中文、英文、数字、空格和常见标点，用于前台展示；若为空则前端回退显示 `name`。
  - `description` (VARCHAR(500) NULL): 备注说明。
  - `group_name` (VARCHAR(120)): 目录名，默认等于 `name` 的 slug 化结果，同字符规则。
  - `relative_path` (VARCHAR(255)): `presentations/<user-uuid>/<group_name>/slides`。
  - `canonical_path` (VARCHAR(512)): 清洗后的绝对路径（根据部署配置生成）。
  - `tags` (JSON NULL): 用于搜索/过滤。
  - `created_at` (DATETIME): 创建时间。
  - `updated_at` (DATETIME): 最近更新时间。
- **Validation rules**:
  - `group_name` 与 `name` 字符集一致，正则 `^[A-Za-z0-9_-]+$`。
  - `relative_path` 根据用户 `uuid` 在服务端生成并只读。
  - 同一用户下 `group_name` 唯一，以避免重复目录。
  - `tags` 限制为 10 个以内的小写 slug。

### AuditEvent (派生)
- **Description**: 用于记录重要操作日志（注册、登录失败、路径校验失败）。
- **Fields**:
  - `id` (INT AUTO_INCREMENT)
  - `user_id` (INT NULL, FK → UserAccount.id)
  - `event_type` (ENUM: `auth.register`, `auth.login`, `auth.logout`, `record.create`, `record.update`, `record.delete`, `record.path_error`)
  - `metadata` (JSON)
  - `created_at` (DATETIME)
- **Notes**: 可根据 MVP 范围选择仅写入结构化日志文件；若使用 MySQL，可存储为 JSON 字段并建立索引。

## Relationships
- `UserAccount 1 - n UserSession`
- `UserAccount 1 - n PptRecord`
- `UserAccount 1 - n AuditEvent`

所有外键在 MySQL 中采用 `ON DELETE CASCADE` 以确保账户注销后记录同步清理；AuditEvent 可选择软删除或长期保留。

## Derived States
- 登录成功：更新 `UserAccount.last_login_at`，新增 `UserSession`，写入 `AuditEvent`。
- 刷新令牌失效：更新 `UserSession.revoked_at` 并阻止进一步使用。
- PPT 路径失效：在 `AuditEvent` 中记录 `record.path_error`，并在客户端展示“需要修复”状态。

## Indexing Strategy
- `UserAccount.email` 唯一索引。
- `UserSession.user_id` + `revoked_at` 组合索引用于快速查询有效会话。
- `PptRecord.user_id` + `created_at` 排序索引用于列表。
- `PptRecord.user_id` + `group_name` 唯一索引确保路径唯一性。
- `PptRecord.tags` 可配合 MySQL JSON 索引进行查询（可选）。
