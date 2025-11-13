<template>
  <section class="flex flex-col gap-8">
    <div class="flex flex-col gap-3">
      <span v-if="features.kicker" class="text-sm font-semibold uppercase tracking-[0.3em] text-emerald-300/80">{{ features.kicker }}</span>
      <h2 class="text-3xl font-semibold text-slate-50">{{ features.title }}</h2>
      <p v-if="features.subtitle" class="max-w-3xl text-base text-slate-400">{{ features.subtitle }}</p>
    </div>
    <div class="grid gap-6 md:grid-cols-3" data-testid="feature-list">
      <article
        v-for="item in features.items"
        :key="item.id"
        data-testid="feature-card"
        class="group flex flex-col gap-4 rounded-2xl border border-slate-800 bg-slate-900/50 p-6 transition hover:border-emerald-500/60"
      >
        <div class="flex h-11 w-11 items-center justify-center rounded-xl bg-emerald-500/15 text-emerald-200">
          <span v-if="item.icon" class="text-xl">{{ item.icon }}</span>
          <span v-else class="text-xl">★</span>
        </div>
        <h3 class="text-xl font-semibold text-slate-50">{{ item.title }}</h3>
        <p class="text-sm text-slate-300">{{ item.description }}</p>
        <ul v-if="item.bullets?.length" class="mt-auto space-y-2 text-sm text-slate-400">
          <li v-for="(bullet, idx) in item.bullets" :key="`${item.id}-bullet-${idx}`" class="flex items-start gap-2">
            <span class="mt-1 h-1.5 w-1.5 rounded-full bg-emerald-400"></span>
            <span>{{ bullet }}</span>
          </li>
        </ul>
      </article>
    </div>
  </section>
</template>

<script setup>
const props = defineProps({
  features: {
    type: Object,
    required: true
  }
})

const features = props.features
</script>
