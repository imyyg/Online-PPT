<template>
  <div class="auth-demo p-6 max-w-md mx-auto">
    <h2 class="text-2xl font-bold mb-4">Authentication Demo</h2>
    
    <!-- Login Status -->
    <div v-if="authStore.isAuthenticated" class="mb-6 p-4 bg-green-100 rounded">
      <p class="text-green-800">Logged in as: {{ authStore.user?.email }}</p>
      <button @click="handleLogout" class="mt-2 px-4 py-2 bg-red-500 text-white rounded">
        Logout
      </button>
    </div>

    <div v-else class="space-y-6">
      <!-- Captcha -->
      <div class="captcha-section">
        <h3 class="font-semibold mb-2">Step 1: Get Captcha</h3>
        <button 
          @click="handleGetCaptcha" 
          :disabled="captchaLoading"
          class="px-4 py-2 bg-blue-500 text-white rounded disabled:opacity-50"
        >
          {{ captchaLoading ? 'Loading...' : 'Get Captcha' }}
        </button>
        
        <div v-if="captcha" class="mt-2">
          <img :src="`data:image/png;base64,${captcha.image}`" alt="Captcha" class="border" />
          <input 
            v-model="captchaCode" 
            placeholder="Enter captcha code"
            class="mt-2 w-full px-3 py-2 border rounded"
          />
        </div>
      </div>

      <!-- Email -->
      <div class="email-section">
        <h3 class="font-semibold mb-2">Step 2: Email</h3>
        <input 
          v-model="email" 
          type="email"
          placeholder="Email address"
          class="w-full px-3 py-2 border rounded"
        />
        <button 
          @click="handleSendCode" 
          :disabled="!captcha || !captchaCode || codeLoading"
          class="mt-2 px-4 py-2 bg-blue-500 text-white rounded disabled:opacity-50"
        >
          {{ codeLoading ? 'Sending...' : 'Send Verification Code' }}
        </button>
        <p v-if="codeSent" class="mt-2 text-green-600">
          Code sent! Expires in {{ codeExpiresIn }} seconds
        </p>
      </div>

      <!-- Password & Code -->
      <div class="auth-section">
        <h3 class="font-semibold mb-2">Step 3: Register/Login</h3>
        <input 
          v-model="password" 
          type="password"
          placeholder="Password (min 10 chars)"
          class="w-full px-3 py-2 border rounded mb-2"
        />
        <input 
          v-model="emailCode" 
          placeholder="Email verification code"
          class="w-full px-3 py-2 border rounded mb-2"
        />
        
        <div class="flex gap-2">
          <button 
            @click="handleRegister" 
            :disabled="authStore.loading"
            class="flex-1 px-4 py-2 bg-green-500 text-white rounded disabled:opacity-50"
          >
            Register
          </button>
          <button 
            @click="handleLogin" 
            :disabled="authStore.loading"
            class="flex-1 px-4 py-2 bg-blue-500 text-white rounded disabled:opacity-50"
          >
            Login
          </button>
        </div>
      </div>

      <!-- Error Display -->
      <div v-if="error" class="p-4 bg-red-100 text-red-800 rounded">
        {{ error }}
      </div>
    </div>

    <!-- PPT Records -->
    <div v-if="authStore.isAuthenticated" class="mt-8">
      <h3 class="text-xl font-bold mb-4">PPT Records</h3>
      
      <button 
        @click="loadRecords" 
        :disabled="pptsStore.loading"
        class="mb-4 px-4 py-2 bg-blue-500 text-white rounded disabled:opacity-50"
      >
        {{ pptsStore.loading ? 'Loading...' : 'Load Records' }}
      </button>

      <div v-if="pptsStore.records.length" class="space-y-2">
        <div 
          v-for="record in pptsStore.records" 
          :key="record.id"
          class="p-3 border rounded"
        >
          <h4 class="font-semibold">{{ record.title || record.name }}</h4>
          <p class="text-sm text-gray-600">{{ record.description }}</p>
          <div v-if="record.tags?.length" class="mt-1">
            <span 
              v-for="tag in record.tags" 
              :key="tag"
              class="inline-block px-2 py-1 text-xs bg-gray-200 rounded mr-1"
            >
              {{ tag }}
            </span>
          </div>
        </div>
      </div>
      <p v-else class="text-gray-500">No records found</p>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useAuthStore } from '../stores/auth'
import { usePptsStore } from '../stores/ppts'
import { useCaptcha, useVerificationCode } from '../composables/useAuth'
import { getErrorMessage } from '../utils/errors'

const authStore = useAuthStore()
const pptsStore = usePptsStore()

const { captcha, loading: captchaLoading, fetchCaptcha } = useCaptcha()
const { loading: codeLoading, sent: codeSent, expiresIn: codeExpiresIn, sendCode } = useVerificationCode()

const email = ref('')
const password = ref('')
const captchaCode = ref('')
const emailCode = ref('')
const error = ref(null)

async function handleGetCaptcha() {
  try {
    error.value = null
    await fetchCaptcha()
  } catch (err) {
    error.value = getErrorMessage(err)
  }
}

async function handleSendCode() {
  try {
    error.value = null
    await sendCode(email.value, captcha.value.captcha_id, captchaCode.value)
  } catch (err) {
    error.value = getErrorMessage(err)
  }
}

async function handleRegister() {
  try {
    error.value = null
    await authStore.register(email.value, password.value, emailCode.value)
  } catch (err) {
    error.value = getErrorMessage(err)
  }
}

async function handleLogin() {
  try {
    error.value = null
    await authStore.login(email.value, password.value)
  } catch (err) {
    error.value = getErrorMessage(err)
  }
}

async function handleLogout() {
  await authStore.logout()
}

async function loadRecords() {
  try {
    await pptsStore.fetchRecords()
  } catch (err) {
    error.value = getErrorMessage(err)
  }
}
</script>
