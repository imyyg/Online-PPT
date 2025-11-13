<template>
  <section class="flex flex-col gap-6 rounded-3xl border border-slate-800 bg-slate-900/60 p-8 md:flex-row md:items-center md:justify-between">
    <div class="flex max-w-2xl flex-col gap-3">
      <span v-if="footer.kicker" class="text-sm font-semibold uppercase tracking-[0.3em] text-emerald-300/80">{{ footer.kicker }}</span>
      <h2 data-testid="footer-headline" class="text-3xl font-semibold text-slate-50">{{ footer.headline }}</h2>
      <p v-if="footer.description" class="text-base text-slate-400">{{ footer.description }}</p>
    </div>
    <div class="flex flex-col gap-3 md:flex-row md:items-center">
      <a
        v-if="footer.primaryCta"
        :href="footer.primaryCta.href"
        :target="footer.primaryCta.target"
        :aria-label="footer.primaryCta.ariaLabel || footer.primaryCta.label"
        data-testid="footer-primary-cta"
        class="inline-flex min-w-[200px] items-center justify-center rounded-full bg-emerald-500 px-6 py-3 text-lg font-semibold text-slate-950 transition hover:bg-emerald-400"
        @click="emitClick(footer.primaryCta)"
      >
        {{ footer.primaryCta.label }}
      </a>
      <p v-if="footer.helper" class="text-sm text-slate-400">{{ footer.helper }}</p>
    </div>
  </section>
</template>

<script setup>
import { trackHomepageCta } from '../../utils/analytics'

const props = defineProps({
  footer: {
    type: Object,
    required: true
  }
})

const footer = props.footer

function emitClick(action) {
  if (!action) return
  const analyticsPayload = action.analytics ? { ...action.analytics } : {}
  analyticsPayload.content = analyticsPayload.content || action.label
  trackHomepageCta(analyticsPayload)
}
</script>
