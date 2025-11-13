# Change: 为 PPT 记录添加 title 字段

## Why
当前 `PptRecord` 表只有 `name` 字段用于标识目录名称（受字符集限制），但用户需要一个更友好的展示标题，支持更丰富的字符集（包括中文、空格等）用于前台展示。这将提升用户体验，让用户可以为 PPT 设置更具描述性的标题。

## What Changes
- 在 `PptRecord` 数据模型中添加 `title` 字段
- `title` 字段为可选，支持中文、英文、数字、空格等常见字符
- `title` 字段长度限制为 255 个字符
- 如果用户未提供 `title`，前端展示时可回退到 `name` 字段
- 添加数据库迁移脚本以支持现有记录的兼容性

## Impact
- **Affected specs**: ppt-records（新建）
- **Affected code**: 
  - `internal/records/repository.go` - 添加 title 字段查询和更新
  - `internal/records/service.go` - 业务逻辑支持 title 字段
  - `internal/http/handlers/records_*.go` - API 端点支持 title 参数
  - `migrations/` - 新增迁移脚本 `003_add_ppt_record_title.sql`
  - 数据库 schema: `ppt_records` 表新增 `title` 列
- **Breaking changes**: 无，字段为可选，向后兼容
