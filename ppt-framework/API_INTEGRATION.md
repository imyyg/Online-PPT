# 前端 API 对接文档

本文档说明如何在前端项目中使用后端 API。

## 目录结构

```
src/
├── api/
│   ├── client.js       # Axios 客户端配置
│   ├── auth.js         # 认证相关 API
│   ├── ppts.js         # PPT 记录管理 API
│   └── index.js        # 统一导出
├── stores/
│   ├── auth.js         # 认证状态管理
│   ├── ppts.js         # PPT 记录状态管理
│   └── slides.js       # 幻灯片展示状态管理
├── composables/
│   └── useAuth.js      # 认证相关组合式函数
└── utils/
    ├── auth.js         # 认证工具函数
    └── errors.js       # 错误处理工具
```

## 环境配置

1. 复制 `.env.example` 为 `.env.local`：
   ```bash
   cp .env.example .env.local
   ```

2. 配置后端 API 地址：
   ```env
   VITE_API_BASE_URL=http://localhost:8080/api/v1
   ```

## API 使用示例

### 认证 API

#### 获取验证码
```javascript
import { authApi } from '@/api'

const captcha = await authApi.getCaptcha()
// captcha = { captcha_id, image, expires_in }
```

#### 发送邮箱验证码
```javascript
const result = await authApi.sendVerificationCode(
  'user@example.com',
  captchaId,
  captchaCode
)
// result = { message, expires_in }
```

#### 用户注册
```javascript
const response = await authApi.register(
  'user@example.com',
  'password123',
  '123456' // 邮箱验证码
)
// response = { user, accessToken, expiresIn }
```

#### 用户登录
```javascript
const response = await authApi.login(
  'user@example.com',
  'password123'
)
// response = { user, accessToken, expiresIn }
```

#### 退出登录
```javascript
await authApi.logout()
```

#### 刷新令牌
```javascript
const response = await authApi.refresh()
// response = { user, accessToken, expiresIn }
```

### PPT 记录 API

#### 列出记录
```javascript
import { pptsApi } from '@/api'

const response = await pptsApi.list({
  q: 'keyword',              // 可选：搜索关键词
  tag: 'technology',         // 可选：标签筛选
  sort: 'created_at_desc',   // 可选：排序方式
  limit: 50,                 // 可选：每页数量
  offset: 0                  // 可选：偏移量
})
// response = { total, limit, offset, items: [...] }
```

#### 获取单个记录
```javascript
const record = await pptsApi.get(recordId)
// record = { id, name, title, description, tags, ... }
```

#### 创建记录
```javascript
const record = await pptsApi.create({
  name: 'my-presentation',
  title: '我的演示文稿',
  description: '这是一个示例演示',
  tags: ['technology', 'tutorial']
})
```

#### 更新记录
```javascript
const updated = await pptsApi.update(recordId, {
  name: 'updated-name',
  title: '更新后的标题',
  description: '更新后的描述',
  tags: ['new-tag']
})
```

#### 删除记录
```javascript
await pptsApi.delete(recordId)
```

## Pinia Store 使用

### 认证 Store

```javascript
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()

// 检查是否已认证
if (authStore.isAuthenticated) {
  console.log('User:', authStore.user)
}

// 登录
try {
  await authStore.login('user@example.com', 'password')
  console.log('Login successful')
} catch (error) {
  console.error('Login failed:', authStore.error)
}

// 注册
try {
  await authStore.register('user@example.com', 'password', '123456')
  console.log('Registration successful')
} catch (error) {
  console.error('Registration failed:', authStore.error)
}

// 退出
await authStore.logout()
```

### PPT 记录 Store

```javascript
import { usePptsStore } from '@/stores/ppts'

const pptsStore = usePptsStore()

// 获取记录列表
await pptsStore.fetchRecords()

// 访问记录
console.log('Records:', pptsStore.records)
console.log('Filtered:', pptsStore.filteredRecords)

// 搜索
pptsStore.setSearchQuery('keyword')

// 按标签筛选
pptsStore.setSelectedTag('technology')

// 创建记录
try {
  const record = await pptsStore.createRecord({
    name: 'my-ppt',
    title: '我的 PPT',
    description: '描述',
    tags: ['tag1', 'tag2']
  })
  console.log('Created:', record)
} catch (error) {
  console.error('Failed:', pptsStore.error)
}

// 更新记录
await pptsStore.updateRecord(recordId, { title: '新标题' })

// 删除记录
await pptsStore.deleteRecord(recordId)
```

## 组合式函数

### 验证码
```javascript
import { useCaptcha } from '@/composables/useAuth'

const { captcha, loading, error, fetchCaptcha } = useCaptcha()

await fetchCaptcha()
if (captcha.value) {
  console.log('Captcha ID:', captcha.value.captcha_id)
  console.log('Image:', captcha.value.image)
}
```

### 邮箱验证码
```javascript
import { useVerificationCode } from '@/composables/useAuth'

const { loading, error, sent, expiresIn, sendCode, reset } = useVerificationCode()

await sendCode('user@example.com', captchaId, captchaCode)
if (sent.value) {
  console.log('Code expires in:', expiresIn.value, 'seconds')
}
```

## 认证守卫

```javascript
import { requireAuth, redirectIfAuthenticated } from '@/utils/auth'

// 在需要认证的页面中
if (!requireAuth()) {
  // 用户将被重定向到登录页
}

// 在登录/注册页面中
if (redirectIfAuthenticated('/dashboard')) {
  // 已登录用户将被重定向到 dashboard
}
```

## 错误处理

```javascript
import { getErrorMessage, isAuthError, isValidationError } from '@/utils/errors'

try {
  await authApi.login(email, password)
} catch (error) {
  const message = getErrorMessage(error)
  
  if (isAuthError(error)) {
    console.log('Authentication failed')
  }
  
  if (isValidationError(error)) {
    console.log('Validation error:', message)
  }
}
```

## Token 自动刷新

axios 客户端已配置自动 token 刷新机制。当收到 401 响应时：

1. 自动使用 refresh token 获取新的 access token
2. 重试原始请求
3. 如果刷新失败，清除认证信息并重定向到登录页

## 注意事项

1. **Token 存储**：access token 和 user 信息存储在 localStorage 中
2. **跨标签同步**：已实现跨标签页的登出同步
3. **请求超时**：默认 10 秒超时
4. **错误处理**：所有 API 调用都应使用 try-catch 处理错误
5. **开发环境**：确保后端服务运行在正确的端口并配置了 CORS

## API 合约

完整的 API 规范请参考：`specs/001-add-go-backend/contracts/api.yaml`
