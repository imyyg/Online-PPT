<template>
  <header class="flex flex-col gap-6 md:flex-row md:items-center md:justify-between">
    <div class="flex items-center gap-2 text-sm uppercase tracking-[0.3em] text-slate-400">
      <span class="h-2 w-2 rounded-full bg-emerald-400"></span>
      <span data-testid="nav-brand">{{ brand }}</span>
    </div>
  <nav class="flex flex-col items-start gap-3 text-sm md:flex-row md:items-center" data-testid="nav-cta-group">
      <a
        v-if="loginAction"
        :href="loginAction.href"
        :target="loginAction.target"
        class="inline-flex min-w-[140px] items-center justify-center rounded-full border border-slate-700 px-5 py-2 font-medium text-slate-200 transition hover:border-slate-500 hover:text-white"
        :aria-label="loginAction.ariaLabel || loginAction.label"
        data-testid="nav-login"
        @click="emitClick(loginAction)"
      >
        {{ loginAction.label }}
      </a>
      <a
        v-if="registerAction"
        :href="registerAction.href"
        :target="registerAction.target"
        class="inline-flex min-w-[160px] items-center justify-center rounded-full bg-emerald-500 px-5 py-2 font-semibold text-slate-950 transition hover:bg-emerald-400"
        :aria-label="registerAction.ariaLabel || registerAction.label"
        data-testid="nav-register"
        @click="emitClick(registerAction)"
      >
        {{ registerAction.label }}
      </a>
    </nav>
  </header>
</template>

<script setup>
import { computed } from 'vue'
import { trackHomepageCta } from '../../utils/analytics'

const props = defineProps({
  brand: {
    type: String,
    default: 'Online PPT'
  },
  navigation: {
    type: Object,
    required: true
  }
})

const registerAction = computed(() => props.navigation?.register || null)
const loginAction = computed(() => props.navigation?.login || null)

function emitClick(action) {
  if (!action) return
  const analyticsPayload = action.analytics ? { ...action.analytics } : {}
  analyticsPayload.content = analyticsPayload.content || action.label
  trackHomepageCta(analyticsPayload)
}
</script>
