# 前端 API 对接完成

## 概述

前端项目已成功对接后端 API（`/specs/001-add-go-backend/contracts/api.yaml`），现在支持：

- ✅ 用户认证（注册、登录、登出）
- ✅ PPT 记录管理（CRUD 操作）
- ✅ 自动 Token 刷新
- ✅ 智能加载策略（API 优先，本地回退）
- ✅ 跨标签页同步

## 快速开始

### 1. 配置环境

```bash
# 确保 .env.local 已创建并配置正确的 API 地址
cat .env.local
# VITE_API_BASE_URL=http://localhost:8080/api/v1
```

### 2. 启动开发服务器

```bash
npm run dev
```

### 3. 确保后端服务运行

```bash
cd ../online-ppt
./server
```

## 项目结构

```
src/
├── api/                    # API 客户端层
│   ├── client.js          # Axios 配置（含 token 刷新）
│   ├── auth.js            # 认证 API
│   ├── ppts.js            # PPT 记录 API
│   └── index.js           # 统一导出
│
├── stores/                 # Pinia 状态管理
│   ├── auth.js            # 认证状态（用户、token）
│   ├── ppts.js            # PPT 记录状态
│   └── slides.js          # 幻灯片状态（已更新）
│
├── composables/            # 可复用组合式函数
│   └── useAuth.js         # 验证码、邮箱验证码
│
└── utils/                  # 工具函数
    ├── auth.js            # 认证守卫、跨标签同步
    └── errors.js          # 错误处理
```

## 使用示例

### 认证流程

```vue
<script setup>
import { useAuthStore } from '@/stores/auth'
import { useCaptcha, useVerificationCode } from '@/composables/useAuth'

const authStore = useAuthStore()
const { captcha, fetchCaptcha } = useCaptcha()
const { sendCode } = useVerificationCode()

// 1. 获取验证码
await fetchCaptcha()

// 2. 发送邮箱验证码
await sendCode(email, captcha.value.captcha_id, captchaCode)

// 3. 注册
await authStore.register(email, password, emailCode)

// 4. 登录
await authStore.login(email, password)

// 5. 登出
await authStore.logout()
</script>
```

### PPT 记录管理

```vue
<script setup>
import { usePptsStore } from '@/stores/ppts'

const pptsStore = usePptsStore()

// 获取列表
await pptsStore.fetchRecords()

// 创建记录
await pptsStore.createRecord({
  name: 'my-ppt',
  title: '我的演示',
  description: '描述',
  tags: ['标签1', '标签2']
})

// 更新记录
await pptsStore.updateRecord(id, { title: '新标题' })

// 删除记录
await pptsStore.deleteRecord(id)

// 搜索
pptsStore.setSearchQuery('关键词')

// 按标签筛选
pptsStore.setSelectedTag('标签1')
</script>
```

## 核心特性

### 1. 自动 Token 刷新

当 access token 过期时，axios 拦截器会：
1. 自动使用 refresh token 获取新 token
2. 重试原始请求
3. 失败则清除认证并跳转登录

### 2. 智能加载策略

`slides` store 的加载逻辑：
1. 用户已登录 → 尝试从 API 加载
2. API 失败 → 自动回退到本地文件
3. 用户未登录 → 直接加载本地文件

这确保了向后兼容性。

### 3. 跨标签页同步

使用 `localStorage` 事件监听：
- 在一个标签页登出
- 其他标签页自动同步登出状态

### 4. 友好的错误处理

```javascript
import { getErrorMessage, isAuthError } from '@/utils/errors'

try {
  await authApi.login(email, password)
} catch (error) {
  const message = getErrorMessage(error)
  if (isAuthError(error)) {
    console.log('认证失败')
  }
}
```

## 文档

- **[API_INTEGRATION.md](./API_INTEGRATION.md)** - 完整 API 使用文档
- **[QUICKSTART.md](./QUICKSTART.md)** - 快速开始指南
- **[API_INTEGRATION_SUMMARY.md](./API_INTEGRATION_SUMMARY.md)** - 功能总结
- **[INTEGRATION_CHECKLIST.md](./INTEGRATION_CHECKLIST.md)** - 完成清单

## 演示组件

已创建 `src/components/AuthDemo.vue` 演示所有功能：

```vue
<template>
  <AuthDemo />
</template>

<script setup>
import AuthDemo from '@/components/AuthDemo.vue'
</script>
```

## API 合约

完整的 API 规范：`/specs/001-add-go-backend/contracts/api.yaml`

## 下一步

1. 创建登录/注册页面 UI
2. 创建 PPT 管理页面
3. 在导航栏显示用户状态
4. 添加 Loading 和 Toast 提示
5. 根据需要添加更多功能

## 常见问题

### CORS 错误

确保后端配置了正确的 CORS：
```go
AllowOrigins: []string{"http://localhost:5173"}
```

### Token 自动刷新失败

检查：
1. refresh token 是否存储在 localStorage
2. 后端 `/auth/refresh` 端点是否正常
3. token 是否已完全过期

### 无法加载演示

检查：
1. 后端是否运行在 `http://localhost:8080`
2. `.env.local` 配置是否正确
3. 用户是否已登录（需要的话）

## 技术栈

- **Vue 3** - 前端框架
- **Pinia** - 状态管理
- **Axios** - HTTP 客户端
- **Vite** - 构建工具

## 兼容性

- ✅ 保持现有本地演示功能
- ✅ 新增 API 管理的演示
- ✅ 两者可以混合使用
- ✅ 平滑迁移路径

---

🎉 **前端 API 对接已完成！可以开始构建用户界面了。**
