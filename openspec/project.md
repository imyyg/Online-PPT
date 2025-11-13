# Project Context

## Purpose

Online-PPT 是一个基于 Web 的演示文稿（PPT）管理和展示系统，支持用户创建、管理和展示 HTML 格式的幻灯片。

**核心目标**:
- 提供用户账户体系，支持邮箱注册登录
- 管理用户创建的 PPT 记录，通过记录快速定位到 PPT 所在文件夹
- 提供友好的 Web 界面进行幻灯片管理和演示
- 保持幻灯片文件的独立性，系统仅存储路径引用

## Tech Stack

### Backend (online-ppt/)
- **Language**: Go 1.22
- **Web Framework**: Gin
- **Database**: MySQL 8.x (开发环境)
- **Authentication**: JWT (golang-jwt/jwt/v5)
- **Password Hashing**: Argon2 (golang.org/x/crypto/argon2)
- **Configuration**: YAML (gopkg.in/yaml.v3)
- **Testing**: testify, go-sqlmock
- **UUID**: google/uuid

### Frontend (ppt-framework/)
- **Framework**: Vue 3 + Composition API
- **Build Tool**: Vite 7
- **State Management**: Pinia
- **Styling**: Tailwind CSS 4
- **Icons**: Lucide Vue Next
- **Utilities**: VueUse

## Project Conventions

### Code Style

#### Go Backend
- 遵循 Go 官方代码规范和惯例
- 使用 `gofmt` 格式化代码
- 包名使用小写单数形式
- 接口命名遵循 `-er` 后缀惯例（如 `Sender`, `Repository`）
- 错误变量以 `Err` 前缀命名（如 `ErrInvalidCredentials`）
- 常量使用驼峰命名（如 `DefaultTimeout`）
- 私有字段和方法以小写字母开头，公开的以大写字母开头
- 禁止在代码中出现中文字符（注释除外）

#### Vue Frontend
- 遵循 Vue 3 Composition API 风格指南
- 组件名使用 PascalCase（如 `SlideLoader.vue`）
- Composables 使用 `use-` 前缀（如 `useSlides`）
- 事件处理函数使用 `handle-` 前缀（如 `handleClick`）
- 使用 `<script setup>` 语法

#### 通用规范
- 缩进使用 4 个空格（Go、Vue、YAML）
- 文档使用中文编写，技术术语保留英文
- 提交信息使用中文，格式清晰明确

### Architecture Patterns

#### Backend Architecture
```
cmd/server/          # 应用入口
internal/
  ├── auth/          # 认证模块（Service, Repository, Models）
  ├── records/       # PPT 记录模块
  ├── mail/          # 邮件服务（计划中）
  ├── http/          # HTTP 层
  │   ├── handlers/  # 请求处理器
  │   └── middleware/# 中间件
  ├── config/        # 配置管理
  └── storage/       # 数据存储和迁移
migrations/          # 数据库迁移脚本
tests/
  ├── integration/   # 集成测试
  └── e2e/          # 端到端测试
```

**分层架构**:
1. **HTTP Layer**: 处理 HTTP 请求/响应，参数验证
2. **Service Layer**: 业务逻辑，协调多个 Repository
3. **Repository Layer**: 数据访问，SQL 查询
4. **Storage Layer**: 数据库连接和迁移管理

**设计原则**:
- 依赖注入：通过构造函数注入依赖
- 接口隔离：定义清晰的接口边界
- 单一职责：每个包专注于单一领域
- 错误处理：使用明确的错误类型，避免 panic

#### Frontend Architecture
```
src/
  ├── components/    # Vue 组件
  ├── stores/        # Pinia 状态管理
  ├── composables/   # 组合式函数
  ├── utils/         # 工具函数
  └── assets/        # 静态资源
presentations/
  └── <group>/
      ├── slides.config.json  # 幻灯片配置
      └── slides/             # 幻灯片 HTML 文件
          ├── slide-1.html
          └── ...
```

**核心模式**:
- 使用 Pinia Store 管理幻灯片状态
- 组件通过 props 传递数据，通过 emit 传递事件
- 使用 Composables 封装可复用逻辑
- 幻灯片通过 iframe 或 Shadow DOM 加载

### Testing Strategy

#### Backend Testing
- **Unit Tests**: 测试独立函数和方法（如密码哈希、验证码生成）
- **Integration Tests**: 测试 HTTP 端点和完整业务流程
- **Test Coverage**: 目标覆盖率 > 80%
- **Mocking**: 使用 `go-sqlmock` 模拟数据库，使用接口模拟外部依赖
- **Test Data**: 使用独立的测试数据库或内存数据库

**测试文件命名**: `*_test.go`

**测试运行**:
```bash
go test ./...                    # 运行所有测试
go test -v ./internal/auth/      # 运行特定包的测试
go test -cover ./...             # 查看覆盖率
```

#### Frontend Testing
- 当前阶段主要依赖手动测试
- 未来计划添加 Vitest 单元测试和 Playwright E2E 测试

### Git Workflow

- **主分支**: `main` - 生产环境代码
- **开发分支**: `develop` - 集成分支
- **功能分支**: `<feature-id>-<description>` (如 `001-add-go-backend`)
- **提交消息**: 使用中文，格式为 `<类型>: <简短描述>`
  - 类型示例: `feat`, `fix`, `refactor`, `docs`, `test`, `chore`
  - 示例: `feat: 添加邮箱验证码功能`, `fix: 修复登录状态检查错误`

## Domain Context

### PPT 路径规范
- PPT 文件存储路径: `presentations/<user-uuid>/<group>/slides/`
- `<user-uuid>`: 用户的 UUID 标识符
- `<group>`: 演示分组名称（字母、数字、下划线、中划线）
- `slides/`: 固定子目录，包含 HTML 幻灯片文件

### 用户状态
- `pending`: 注册但未验证邮箱（计划中）
- `active`: 正常活跃用户
- `locked`: 被锁定的账号

### 认证机制
- **Access Token**: 短期令牌（默认 15 分钟），用于 API 访问
- **Refresh Token**: 长期令牌（默认 7 天），存储在 HttpOnly Cookie 中
- **Session**: 存储在数据库中，支持设备指纹识别

### PPT 记录原则
- 系统仅存储 PPT 的元数据和路径引用
- 不修改实际的 HTML 幻灯片文件
- 路径验证确保目录存在且可访问
- 用户只能访问自己创建的记录

## Important Constraints

### Technical Constraints
- Go 版本: 1.22+（利用新特性和性能改进）
- MySQL 版本: 8.0+（支持更好的 JSON 和索引功能）
- 前端浏览器支持: 现代浏览器（Chrome, Firefox, Safari, Edge）
- 部署环境: WSL 或容器环境

### Business Constraints
- 当前阶段仅支持个人使用，不涉及团队协作
- 无需邮件验证码功能（已计划添加）
- 暂不支持跨用户共享 PPT
- 密码策略: 最小长度和复杂度要求

### Security Constraints
- 密码必须使用 Argon2 哈希存储
- JWT 令牌必须包含过期时间
- 敏感操作需要身份验证
- SQL 注入防护（使用参数化查询）
- XSS 防护（前端适当转义）

### Performance Constraints
- API 响应时间: P95 < 500ms
- 注册/登录响应: < 700ms
- PPT 记录查询: < 3s（100+ 条记录需分页）
- 数据库连接池: 根据负载调整

## External Dependencies

### Development Tools
- **Database**: MySQL 8.x 开发实例
- **Go Modules**: 依赖管理
- **Node/npm**: 前端构建工具

### Planned Dependencies
- **SMTP Service**: 邮件发送（支持 Gmail, QQ, 163, 阿里云等）
- **Email Verification**: 验证码功能（计划中）

### Third-party Libraries

**Backend**:
- `gin-gonic/gin`: HTTP 路由和中间件
- `golang-jwt/jwt`: JWT 令牌生成和验证
- `go-sql-driver/mysql`: MySQL 驱动
- `google/uuid`: UUID 生成
- `golang.org/x/crypto`: 密码哈希

**Frontend**:
- `vue`: 响应式 UI 框架
- `pinia`: 状态管理
- `vueuse/core`: Vue 组合式工具库
- `lucide-vue-next`: 图标库
- `tailwindcss`: CSS 框架

## Notes

- 项目分为两个独立的子项目: `online-ppt/` (后端) 和 `ppt-framework/` (前端)
- 后端使用 RESTful API 风格
- 前端直接访问文件系统读取幻灯片（开发模式）
- 生产环境需要配置静态文件服务
- 所有新功能变更应通过 OpenSpec 流程创建提案
