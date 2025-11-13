<template>
  <section class="flex flex-col gap-12">
    <div class="flex flex-col gap-4">
      <span v-if="socialProof.kicker" class="text-sm font-semibold uppercase tracking-[0.3em] text-emerald-300/80">{{ socialProof.kicker }}</span>
      <h2 class="text-3xl font-semibold text-slate-50">{{ socialProof.title }}</h2>
      <p v-if="socialProof.subtitle" class="max-w-3xl text-base text-slate-400">{{ socialProof.subtitle }}</p>
    </div>
    <div v-if="socialProof.logos?.length" class="flex flex-wrap items-center gap-6 opacity-80">
      <span
        v-for="logo in socialProof.logos"
        :key="logo.alt"
        class="text-sm font-medium uppercase tracking-[0.25em] text-slate-500"
      >
        {{ logo.alt }}
      </span>
    </div>
    <div v-if="socialProof.testimonials?.length" class="grid gap-6 md:grid-cols-3">
      <article
        v-for="quote in socialProof.testimonials"
        :key="quote.author"
        data-testid="testimonial-item"
        class="flex flex-col gap-4 rounded-2xl border border-slate-800 bg-slate-900/60 p-6"
      >
        <p class="text-base text-slate-200">“{{ quote.quote }}”</p>
        <div class="mt-auto text-sm text-slate-400">
          <div class="font-semibold text-slate-50">{{ quote.author }}</div>
          <div>{{ quote.role }}</div>
        </div>
      </article>
    </div>
  </section>
</template>

<script setup>
const props = defineProps({
  socialProof: {
    type: Object,
    required: true
  }
})

const socialProof = props.socialProof
</script>
