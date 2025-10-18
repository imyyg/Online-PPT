import { createApp } from 'vue'
import { createPinia } from 'pinia'
import './tw.css'
import './style.css'
import App from './App.vue'
import { useSlidesStore } from './stores/slides'

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)

// Preload slides config before first render
const store = useSlidesStore(pinia)
await store.loadConfig()

app.mount('#app')