<template>
  <div class="homepage-shell min-h-screen bg-slate-950 text-slate-50">
    <div class="mx-auto flex w-full max-w-6xl flex-col gap-20 px-6 pb-24 pt-12">
      <HomepageTopNav :brand="brand" :navigation="navigation" />
      <main class="homepage-main flex flex-col gap-20">
        <HomepageHero :hero="hero" />
        <div class="rounded-3xl border border-slate-800 bg-slate-900/40 p-10 text-slate-300">
          <slot name="features">
            <HomepageFeatures :features="features" />
          </slot>
        </div>
        <div class="rounded-3xl border border-slate-800 bg-slate-900/40 p-10 text-slate-300">
          <slot name="social-proof">
            <HomepageSocialProof :social-proof="socialProof" />
          </slot>
        </div>
      </main>
      <footer class="homepage-footer border-t border-slate-800 pt-10 text-slate-400">
        <slot name="footer">
          <HomepageFooterCta :footer="footerCta" />
        </slot>
      </footer>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import HomepageTopNav from './HomepageTopNav.vue'
import HomepageHero from './HomepageHero.vue'
import HomepageFeatures from './HomepageFeatures.vue'
import HomepageSocialProof from './HomepageSocialProof.vue'
import HomepageFooterCta from './HomepageFooterCta.vue'
import {
  getHomepageNavigation,
  getHomepageHero,
  getHomepageFeatures,
  getHomepageSocialProof,
  getHomepageFooterCta
} from '../../utils/homepageContent.js'

const props = defineProps({
  brand: {
    type: String,
    default: 'Online PPT'
  }
})

const navigation = computed(() => getHomepageNavigation())
const hero = computed(() => getHomepageHero())
const features = computed(() => getHomepageFeatures())
const socialProof = computed(() => getHomepageSocialProof())
const footerCta = computed(() => getHomepageFooterCta())

const brand = computed(() => props.brand)
</script>
