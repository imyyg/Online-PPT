# Research: 用户登录与PPT记录服务

## 决策概览

### Go 版本
- **Decision**: 使用 Go 1.22 作为后端工具链。
- **Rationale**: 提供稳定的泛型支持与标准库改进，WSL 和容器镜像均已提供官方发行版，便于 CI 复用。
- **Alternatives considered**: Go 1.21（长期支持但缺少部分性能修复）；Go 1.20（已接近维护末期，放弃）。

### Web 框架
- **Decision**: 采用 Gin 作为 REST API 框架。
- **Rationale**: 社区成熟、文档完善、保持与标准库高度兼容；内置路由、中间件可满足认证、日志等需求；对 JSON 处理友好。
- **Alternatives considered**: Fiber（性能佳但依赖 fasthttp，生态包容性差）；Echo（功能接近但社区活跃度略低）。

### 数据存储
- **Decision**: 使用 MySQL 8.x 作为核心关系数据库，所有主键采用 INT AUTO_INCREMENT，用户表额外保存 UUID 字段。
- **Rationale**: 易于本地与生产部署，InnoDB 支持事务与外键；与常见运维体系兼容，后续水平扩展友好。
- **Alternatives considered**: SQLite（轻量但对多实例与权限控制支持不足）；PostgreSQL（功能强但当前团队更熟悉 MySQL）；文件存储（缺乏事务和并发控制）。

### 认证策略
- **Decision**: 采用基于 HMAC 的 JWT (Access Token + Refresh Token) 方案，并在服务端保存刷新令牌哈希以便吊销。
- **Rationale**: 便于前后端分离部署，后端可无状态扩展；配合 HttpOnly Cookie 能减少 XSS 风险。
- **Alternatives considered**: 纯 Session 存储（需要额外缓存层）；前端本地存储 Token（安全性较低）。

### 密码安全
- **Decision**: 使用 Argon2id（`golang.org/x/crypto/argon2`）进行密码哈希。
- **Rationale**: 抗 GPU 攻击、参数可调，Go 社区推荐方案。
- **Alternatives considered**: bcrypt（实现成熟但成本固定）；scrypt（配置复杂且 Go 支持度较低）。

### 测试策略
- **Decision**: 采用 Go `testing` + `httptest` + Testify；对数据库相关逻辑使用事务回滚的集成测试。
- **Rationale**: 组合轻量、学习成本低，兼容 Gin 与 MySQL；易于在 CI 中运行。
- **Alternatives considered**: Ginkgo/Gomega（DSL 较重）；仅用标准库断言（可读性差）。

### 性能与规模假设
- **Decision**: 以单用户 500 条记录为初始规模目标，API 响应 95% 不超过 500ms，列表分页 100 条以内。
- **Rationale**: 对应规格成功指标（SC-001 至 SC-003），MySQL 可通过索引轻松支撑；为后续扩容预留余量。
- **Alternatives considered**: 更高规模预估（需要提早规划分片/读写分离）；使用缓存层（当前无必要）。

### 路径校验策略
- **Decision**: 使用 `filepath.Clean` 清洗后验证路径是否符合 `presentations/<user-uuid>/<group>/slides` 模式，并检查 `<group>` 对应的名称规则。
- **Rationale**: 防止目录遍历攻击，确保记录落在合法目录层级；可跨平台复用。
- **Alternatives considered**: 仅通过字符串前缀判断（可能被 `../` 绕过）；数据库约束路径字符串（缺乏运行时校验）。

### 快速定位策略
- **Decision**: 为每个用户创建独立的 `presentations/<user-uuid>/` 子目录，并在记录表中缓存相对与绝对路径；目录不存在时自动创建。
- **Rationale**: 目录按用户隔离，避免命名冲突；缓存路径便于前端快速定位。
- **Alternatives considered**: 共用单一目录（高冲突风险）；仅存绝对路径（环境迁移困难）。
