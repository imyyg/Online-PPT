<template>
  <section class="grid gap-12 lg:grid-cols-[minmax(0,1fr)_420px]">
    <div class="flex flex-col gap-6">
      <span v-if="hero.kicker" class="text-sm font-semibold uppercase tracking-[0.35em] text-emerald-300/80">{{ hero.kicker }}</span>
      <h1 data-testid="hero-title" class="text-4xl font-semibold leading-tight text-slate-50 sm:text-5xl">
        {{ hero.title }}
      </h1>
      <p v-if="hero.subtitle" data-testid="hero-subtitle" class="max-w-2xl text-lg text-slate-300">
        {{ hero.subtitle }}
      </p>
      <p v-if="hero.copy" class="max-w-2xl text-base text-slate-400">
        {{ hero.copy }}
      </p>
      <div class="flex flex-col gap-4 sm:flex-row sm:items-center">
        <a
          v-if="hero.primaryCta"
          :href="hero.primaryCta.href"
          :target="hero.primaryCta.target"
          :aria-label="hero.primaryCta.ariaLabel || hero.primaryCta.label"
          data-testid="hero-primary-cta"
          class="inline-flex min-w-[200px] items-center justify-center rounded-full bg-emerald-500 px-6 py-3 text-lg font-semibold text-slate-950 transition hover:bg-emerald-400"
          @click="emitClick(hero.primaryCta)"
        >
          {{ hero.primaryCta.label }}
        </a>
        <a
          v-if="hero.secondaryCta"
          :href="hero.secondaryCta.href"
          :target="hero.secondaryCta.target"
          :aria-label="hero.secondaryCta.ariaLabel || hero.secondaryCta.label"
          data-testid="hero-secondary-cta"
          class="inline-flex items-center gap-2 text-base font-semibold text-emerald-300 transition hover:text-emerald-200"
          @click="emitClick(hero.secondaryCta)"
        >
          <span>{{ hero.secondaryCta.label }}</span>
          <svg aria-hidden="true" class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M5 12h14" stroke-linecap="round" stroke-linejoin="round"></path>
            <path d="M13 6l6 6-6 6" stroke-linecap="round" stroke-linejoin="round"></path>
          </svg>
        </a>
      </div>
    </div>
    <div class="relative flex w-full items-center justify-center">
      <div class="absolute inset-0 rounded-[48px] bg-gradient-to-br from-emerald-500/10 via-sky-500/10 to-transparent blur-3xl"></div>
      <div class="relative flex w-full max-w-[420px] flex-col gap-4 rounded-3xl border border-slate-800 bg-slate-900/80 p-6 text-sm text-slate-300 shadow-[0_20px_80px_rgba(16,185,129,0.15)]">
        <p v-if="preview.heading" class="text-xs font-semibold uppercase tracking-[0.4em] text-emerald-300/80">{{ preview.heading }}</p>
        <h2 v-if="preview.title" class="text-xl font-semibold text-slate-50">{{ preview.title }}</h2>
        <ul v-if="preview.bullets?.length" class="grid gap-3 text-sm text-slate-300">
          <li
            v-for="(bullet, idx) in preview.bullets"
            :key="`preview-bullet-${idx}`"
            class="flex items-start gap-3"
          >
            <span :class="['mt-1 h-2 w-2 rounded-full', previewBulletClasses[idx] || 'bg-emerald-400']"></span>
            <span>{{ bullet }}</span>
          </li>
        </ul>
      </div>
    </div>
  </section>
</template>

<script setup>
import { computed } from 'vue'
import { trackHomepageCta } from '../../utils/analytics'

const props = defineProps({
  hero: {
    type: Object,
    required: true
  }
})

const preview = computed(() => props.hero?.preview || { bullets: [] })
const previewBulletClasses = ['bg-emerald-400', 'bg-sky-400', 'bg-indigo-400']

function emitClick(action) {
  if (!action) return
  const analyticsPayload = action.analytics ? { ...action.analytics } : {}
  analyticsPayload.content = analyticsPayload.content || action.label
  trackHomepageCta(analyticsPayload)
}
</script>
