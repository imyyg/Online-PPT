<template>
  <div class="slide-loader" :class="{ 'fullscreen': isFullscreen }">
    <div v-if="loadingError" class="error-message">
      <AlertCircle class="w-12 h-12 text-red-500 mb-4" />
      <p class="text-lg">Failed to load slide: {{ slidePath }}</p>
      <p class="text-sm text-gray-400 mt-2">{{ loadingError }}</p>
    </div>

    <!-- Render iframe even during loading; overlay spinner separately -->
    <iframe
      v-if="loadMode === 'iframe'"
      ref="slideFrame"
      :src="slideUrl"
      :title="slide.title"
      class="slide-frame"
      @load="onFrameLoad"
      @error="onFrameError"
      sandbox="allow-scripts allow-same-origin"
    />

    <div
      v-else-if="loadMode === 'shadow'"
      ref="shadowHost"
      class="shadow-host"
    />

    <div v-if="loading && !loadingError" class="loading-spinner">
      <Loader2 class="w-12 h-12 animate-spin text-primary-500" />
      <p class="mt-4 text-gray-400">Loading slide...</p>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { AlertCircle, Loader2 } from 'lucide-vue-next'
import { useSlidesStore } from '../stores/slides'

const store = useSlidesStore()

const props = defineProps({
  slide: {
    type: Object,
    required: true
  },
  loadMode: {
    type: String,
    default: 'iframe',
    validator: (value) => ['iframe', 'shadow'].includes(value)
  },
  isFullscreen: {
    type: Boolean,
    default: false
  },
  enablePointerProxy: {
    type: Boolean,
    default: true
  }
})

const slideFrame = ref(null)
const shadowHost = ref(null)
const loading = ref(true)
const loadingError = ref('')

// resolve slide path within current group
const slidePath = computed(() => props.slide?.file || '')
const slideUrl = computed(() => {
  const base = store.groupBasePath || ''
  const normalizedBase = typeof base === 'string' ? base : base.value
  // Always require group base path; no legacy fallback
  return `${normalizedBase}/slides/${slidePath.value}`
})

let shadowRoot = null
let removeIframeMouseListener = null
let removeShadowMouseListener = null

function dispatchMouseProxy(x) {
  try {
    const evt = new CustomEvent('slide-pointer-move', { detail: { x } })
    document.dispatchEvent(evt)
  } catch (e) {
    // ignore
  }
}

function attachIframeMouseProxy() {
  if (!props.enablePointerProxy) return
  try {
    const win = slideFrame.value?.contentWindow
    const frameRect = slideFrame.value?.getBoundingClientRect()
    if (!win || !frameRect) return
    const handler = (e) => {
      const x = frameRect.left + e.clientX
      dispatchMouseProxy(x)
    }
    win.addEventListener('mousemove', handler)
    removeIframeMouseListener = () => win.removeEventListener('mousemove', handler)
  } catch (e) {
    // cross-origin or unready
  }
}

function attachShadowMouseProxy() {
  if (!props.enablePointerProxy) return
  if (!shadowHost.value || !shadowRoot) return
  const rect = shadowHost.value.getBoundingClientRect()
  const handler = (e) => {
    const x = rect.left + e.clientX
    dispatchMouseProxy(x)
  }
  shadowRoot.addEventListener('mousemove', handler)
  removeShadowMouseListener = () => shadowRoot.removeEventListener('mousemove', handler)
}

async function loadShadowContent() {
  try {
    loading.value = true
    loadingError.value = ''
    
    const response = await fetch(slideUrl.value)
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }
    
    const html = await response.text()
    
    // Create shadow root if not exists
    if (!shadowRoot && shadowHost.value) {
      shadowRoot = shadowHost.value.attachShadow({ mode: 'open' })
    }
    
    // Parse HTML and extract styles and content
    const parser = new DOMParser()
    const doc = parser.parseFromString(html, 'text/html')
    
    // Clear existing content
    shadowRoot.innerHTML = ''
    
    // Copy styles
    const styles = doc.querySelectorAll('style, link[rel="stylesheet"]')
    for (const style of styles) {
      shadowRoot.appendChild(style.cloneNode(true))
    }
    
    // Copy body content and enforce vertical centering
    const bodyContent = doc.body.cloneNode(true)
    try {
      bodyContent.style.margin = '0'
      bodyContent.style.padding = '0'
      bodyContent.style.height = '100vh'
      bodyContent.style.overflow = 'hidden'
      bodyContent.style.display = 'flex'
      bodyContent.style.alignItems = 'center'
      bodyContent.style.justifyContent = 'center'
    } catch (e) {
      // no-op
    }
    shadowRoot.appendChild(bodyContent)
    
    loading.value = false
    attachShadowMouseProxy()
  } catch (error) {
    console.error('Failed to load shadow content:', error)
    loadingError.value = error.message
    loading.value = false
  }
}

function onFrameLoad() {
  loading.value = false
  
  // Try to apply some styles to the iframe content and center content vertically
  try {
    const frameDoc = slideFrame.value.contentDocument
    if (frameDoc) {
      frameDoc.body.style.margin = '0'
      frameDoc.body.style.padding = '0'
      frameDoc.body.style.height = '100vh'
      frameDoc.body.style.overflow = 'hidden'
      frameDoc.body.style.display = 'flex'
      frameDoc.body.style.alignItems = 'center'
      frameDoc.body.style.justifyContent = 'center'
    }
  } catch (e) {
    console.debug('Cross-origin restriction when accessing iframe content')
  }

  // Forward mouse events from iframe so controls can reappear
  if (props.enablePointerProxy) attachIframeMouseProxy()
}

function onFrameError(error) {
  console.error('Frame loading error:', error)
  loadingError.value = 'Failed to load slide content'
  loading.value = false
}

function forceIframeReload() {
  if (!slideFrame.value) return
  try {
    const url = slideUrl.value
    const glue = url.includes('?') ? '&' : '?'
    const newUrl = `${url}${glue}v=${Date.now()}`
    slideFrame.value.src = newUrl
  } catch (e) {
    // ignore
  }
}

// Watch for slide changes
watch(() => props.slide, () => {
  loading.value = true
  loadingError.value = ''
  
  if (props.loadMode === 'shadow') {
    loadShadowContent()
  } else {
    forceIframeReload()
  }
}, { immediate: true })

// Watch for load mode changes
watch(() => props.loadMode, () => {
  if (props.loadMode === 'shadow') {
    loadShadowContent()
  }
})

onMounted(() => {
  if (props.loadMode === 'shadow') {
    loadShadowContent()
  }
})

onUnmounted(() => {
  shadowRoot = null
  if (removeIframeMouseListener) removeIframeMouseListener()
  if (removeShadowMouseListener) removeShadowMouseListener()
})
</script>

<style scoped>
@reference "../tw.css";
.slide-loader {
  @apply w-full h-full relative bg-gray-900 rounded-lg overflow-hidden;
}

.slide-loader.fullscreen {
  @apply rounded-none;
}

.error-message {
  @apply absolute inset-0 flex flex-col items-center justify-center text-center p-8;
}

.loading-spinner {
  @apply absolute inset-0 flex flex-col items-center justify-center;
}

.slide-frame {
  @apply w-full h-full border-0;
}

.shadow-host {
  @apply w-full h-full;
}

/* Ensure content fills the container */
.shadow-host :deep(*) {
  max-width: 100%;
  max-height: 100%;
}
</style>
