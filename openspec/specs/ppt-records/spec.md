# ppt-records Specification

## Purpose
TBD - created by archiving change add-ppt-record-title. Update Purpose after archive.
## Requirements
### Requirement: PPT Record Title Field
系统 SHALL 允许用户为 PPT 记录设置自定义标题（title），用于前台友好展示。

#### Scenario: 创建 PPT 记录时提供 title
- **WHEN** 用户创建新的 PPT 记录并提供 title 字段
- **THEN** 系统应存储 title 值到 `ppt_records.title` 字段
- **AND** title 应支持中文、英文、数字、空格和常见标点符号
- **AND** title 长度不得超过 255 个字符

#### Scenario: 创建 PPT 记录时未提供 title
- **WHEN** 用户创建新的 PPT 记录但未提供 title 字段
- **THEN** 系统应允许 title 为空（NULL）
- **AND** 前端展示时应回退显示 name 字段值

#### Scenario: 更新 PPT 记录的 title
- **WHEN** 用户更新现有 PPT 记录的 title 字段
- **THEN** 系统应更新 `ppt_records.title` 为新值
- **AND** 更新操作应触发 `updated_at` 字段自动更新

#### Scenario: 查询 PPT 记录时返回 title
- **WHEN** 用户查询自己的 PPT 记录列表或详情
- **THEN** 响应应包含 title 字段
- **AND** 如果 title 为空，前端应显示 name 字段作为回退

#### Scenario: title 验证失败
- **WHEN** 用户提供的 title 超过 255 个字符
- **THEN** 系统应返回 400 错误
- **AND** 错误消息应明确指出字符长度限制

### Requirement: Database Migration for Title Field
系统 MUST 提供数据库迁移脚本以添加 title 字段，确保向后兼容。

#### Scenario: 执行迁移脚本
- **WHEN** 运行新的数据库迁移脚本
- **THEN** `ppt_records` 表应增加 `title` 列
- **AND** 列定义为 `VARCHAR(255) NULL`
- **AND** 现有记录的 title 字段应为 NULL
- **AND** 迁移应不影响现有数据的完整性

