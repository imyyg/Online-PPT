import { createApp, watch } from 'vue'
import { createPinia } from 'pinia'
import './tw.css'
import './style.css'
import App from './App.vue'
import { useSlidesStore } from './stores/slides'

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)

const store = useSlidesStore(pinia)
// Read PPT name from path before loading config
const pathname = window.location.pathname || '/'
const BASE_URL = import.meta.env.BASE_URL || '/'
const parts = pathname.split('/').filter(Boolean)
const baseParts = BASE_URL.split('/').filter(Boolean)
const groupParts = parts.slice(baseParts.length)
const nameFromPath = groupParts[0] || 'example'
store.setGroup(nameFromPath)

// Preload slides config before first render
await store.loadConfig()

// Go to slide if provided (1-based)
const params = new URLSearchParams(window.location.search)
const slideParam = params.get('slide')
if (slideParam) {
  const n = parseInt(slideParam, 10)
  if (!Number.isNaN(n) && n > 0) {
    store.goToSlide(n - 1)
  }
}

// Sync current PPT name to path (example -> BASE_URL)
watch(() => store.currentGroup, (val) => {
  const url = new URL(window.location.href)
  const newPath = (val === 'example') ? BASE_URL : `${BASE_URL}${val}`
  url.pathname = newPath
  window.history.replaceState(null, '', url.toString())
})

// Keep slide number in query for deep-linking
watch(() => store.currentIndex, (idx) => {
  const url = new URL(window.location.href)
  const n = Number(idx) + 1
  url.searchParams.set('slide', String(n))
  window.history.replaceState(null, '', url.toString())
})

app.mount('#app')