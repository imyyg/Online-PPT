# 快速开始指南

本指南帮助你快速在前端项目中对接后端 API。

## 1. 安装依赖

已安装依赖：
- axios - HTTP 客户端

## 2. 配置环境变量

```bash
# 创建本地环境配置文件
cp .env.example .env.local
```

编辑 `.env.local`：
```env
VITE_API_BASE_URL=http://localhost:8080/api/v1
```

## 3. 使用示例

### 基础认证流程

```vue
<script setup>
import { ref } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useCaptcha, useVerificationCode } from '@/composables/useAuth'

const authStore = useAuthStore()
const { captcha, fetchCaptcha } = useCaptcha()
const { sendCode } = useVerificationCode()

const email = ref('')
const password = ref('')
const captchaCode = ref('')
const emailCode = ref('')

// 1. 获取验证码
await fetchCaptcha()

// 2. 发送邮箱验证码
await sendCode(email.value, captcha.value.captcha_id, captchaCode.value)

// 3. 注册
await authStore.register(email.value, password.value, emailCode.value)

// 4. 登录
await authStore.login(email.value, password.value)

// 5. 检查认证状态
console.log('Is authenticated:', authStore.isAuthenticated)
console.log('Current user:', authStore.user)

// 6. 退出
await authStore.logout()
</script>
```

### 管理 PPT 记录

```vue
<script setup>
import { onMounted } from 'vue'
import { usePptsStore } from '@/stores/ppts'

const pptsStore = usePptsStore()

// 加载记录列表
onMounted(async () => {
  await pptsStore.fetchRecords()
})

// 创建新记录
async function createPresentation() {
  await pptsStore.createRecord({
    name: 'my-presentation',
    title: '我的演示文稿',
    description: '这是一个示例',
    tags: ['tutorial', 'demo']
  })
}

// 更新记录
async function updatePresentation(id) {
  await pptsStore.updateRecord(id, {
    title: '更新后的标题'
  })
}

// 删除记录
async function deletePresentation(id) {
  await pptsStore.deleteRecord(id)
}
</script>

<template>
  <div>
    <div v-if="pptsStore.loading">Loading...</div>
    <div v-else>
      <div v-for="record in pptsStore.records" :key="record.id">
        <h3>{{ record.title || record.name }}</h3>
        <p>{{ record.description }}</p>
        <div v-if="record.tags?.length">
          <span v-for="tag in record.tags" :key="tag">{{ tag }}</span>
        </div>
      </div>
    </div>
  </div>
</template>
```

## 4. 集成到现有应用

slides store 已经自动集成了 API 支持：

- 当用户已登录且访问非 example 演示时，会尝试从 API 加载
- 如果 API 加载失败，会自动回退到本地文件加载
- 这意味着无缝兼容现有的本地演示和新的后端管理的演示

## 5. 开发调试

启动开发服务器：
```bash
npm run dev
```

确保后端服务器运行：
```bash
cd ../online-ppt
./server
```

## 6. 项目结构

```
src/
├── api/              # API 客户端和接口定义
├── stores/           # Pinia 状态管理
├── composables/      # 可复用的组合式函数
├── utils/            # 工具函数
└── components/       # Vue 组件
```

## 7. 下一步

查看完整的 API 文档：[API_INTEGRATION.md](./API_INTEGRATION.md)

## 常见问题

### CORS 错误

如果遇到 CORS 错误，确保后端已配置允许前端域名：
```go
// 在后端代码中
router.Use(cors.New(cors.Config{
    AllowOrigins: []string{"http://localhost:5173"},
    AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
    AllowHeaders: []string{"Authorization", "Content-Type"},
}))
```

### Token 过期

Token 会自动刷新，无需手动处理。如果刷新失败，用户会被重定向到登录页。

### 网络错误

使用错误处理工具检测网络问题：
```javascript
import { isNetworkError } from '@/utils/errors'

try {
  await authApi.login(email, password)
} catch (error) {
  if (isNetworkError(error)) {
    console.log('Network connection failed')
  }
}
```
