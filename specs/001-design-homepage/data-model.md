# Data Model: 网站首页概念设计

## Entity: HomepageSection
- **Description**: 首页各语义区块（Hero、功能亮点、示例、社会证明、页脚），用于驱动内容组件渲染。
- **Fields**:
  - `id` (string): 区块唯一标识，如 `hero`, `features`, `showcase`, `social-proof`, `footer`。
  - `title` (string): 区块标题，Hero 可为空，需提供本地化预留。
  - `subtitle` (string): 补充文案，Hero 作为一句话价值主张必填。
  - `media` (object|null): 关联图像或插画的元数据，包含 `type`, `src`, `alt`，用于无障碍朗读。
  - `cta` (CTAButton|null): 区块级别的行动召唤，如“立即开始”。
  - `items` (array<object>): 功能亮点或客户证言列表，字段包括 `headline`, `body`, `icon`。
  - `analyticsTag` (string|null): 自定义埋点标记，便于区分点击来源。
- **Relationships**:
  - 包含多个 `CTAButton` 实例（通过 `cta` 或 `items[n].cta` 间接引用）。
- **Validation Rules**:
  - `id` 必须唯一且符合 `^[a-z0-9-]+$`。
  - Hero 区块必须提供 `subtitle` 与至少一个 CTA。
  - `items` 中的文案长度应控制在 60 字以内以保持可扫描性。

## Entity: CTAButton
- **Description**: 页面中所有主要与次要的点击按钮。
- **Fields**:
  - `label` (string): 按钮文案，需短于 16 个字符。
  - `variant` (enum): `primary`, `secondary`, `link`，用于 Tailwind 样式绑定。
  - `href` (string): 目标链接，注册/登录指向现有路径，示例 CTA 采用新标签打开。
  - `target` (enum|null): `_self` 或 `_blank`，默认 `_self`。
  - `ariaLabel` (string): 无障碍朗读文案，预设与 `label` 相同，可按需覆盖。
  - `analytics` (object): 埋点 payload，字段包含 `eventName`, `placement`, `content`。
- **Relationships**:
  - 与 `HomepageSection` 组合使用。
- **Validation Rules**:
  - `href` 必须为站内受控路径或允许的模板示例链接。
  - 左上角注册/登录按钮 `variant` 固定为 `secondary` 与 `primary` 对。
  - analytics `eventName` 统一为 `cta_click`，`placement` 需标识按钮位置（如 `header-register`）。

## Entity: ImpressionMetrics (虚拟数据模型)
- **Description**: 用于记录首页指标的分析字段参考（不直接存储，但供埋点 schema 使用）。
- **Fields**:
  - `eventName`: 固定 `cta_click`。
  - `placement`: CTA 所在区域，如 `hero-primary`, `header-login`。
  - `timestamp`: ISO 时间戳，由分析工具注入。
  - `additional`: 自定义 payload，如 `copy_variant`。
- **Usage**: 为产品分析团队提供埋点字段约束，以便仪表盘开发对齐。
