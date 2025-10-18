<template>
  <div class="presentation-controls" :class="{ 'hidden': hideControls }">
    <!-- Navigation buttons -->
    <button
      @click="store.prevSlide"
      :disabled="!store.canGoPrev"
      class="nav-btn nav-prev"
      :class="{ 'opacity-0': !showPrevButton }"
    >
      <ChevronLeft class="w-6 h-6" />
    </button>
    
    <button
      @click="store.nextSlide"
      :disabled="!store.canGoNext"
      class="nav-btn nav-next"
      :class="{ 'opacity-0': !showNextButton }"
    >
      <ChevronRight class="w-6 h-6" />
    </button>
    
    <!-- Top toolbar (only right controls) -->
    <div class="top-toolbar">
      <div class="toolbar-right">
        <button
          @click="toggleFullscreen"
          class="toolbar-btn"
        >
          <Maximize2 v-if="!isFullscreen" class="w-5 h-5" />
          <Minimize2 v-else class="w-5 h-5" />
        </button>
        
        <!-- Autoplay toggle -->
        <button
          @click="toggleAutoplay"
          class="toolbar-btn"
          :class="{ 'active': store.config.settings.autoPlay }"
          :title="store.config.settings.autoPlay ? 'Pause autoplay' : 'Start autoplay'"
        >
          <Clock v-if="!store.config.settings.autoPlay" class="w-5 h-5" />
          <Pause v-else class="w-5 h-5" />
        </button>
        
        <!-- Presentation mode toggle -->
        <button
          @click="store.togglePresentation"
          class="toolbar-btn"
          :class="{ 'active': store.isPresenting }"
          :title="store.isPresenting ? 'Exit presentation' : 'Enter presentation'"
        >
          <Play v-if="!store.isPresenting" class="w-5 h-5" />
          <StopCircle v-else class="w-5 h-5" />
        </button>
      </div>
    </div>
    
    <!-- Progress bar at very top of screen -->
    <div v-if="store.config.settings.showProgress" class="progress-bar">
      <div 
        class="progress-fill" 
        :style="{ width: `${store.progress}%`, backgroundColor: store.config.theme.primaryColor }"
      />
    </div>
  </div>
</template>

<style scoped>
@reference "../tw.css";
.presentation-controls {
  @apply pointer-events-none absolute inset-0;
}

.presentation-controls.hidden .top-toolbar,
.presentation-controls.hidden .progress-bar {
  @apply opacity-0;
}

/* Stick progress bar to the top edge of slide-container */
.progress-bar { @apply absolute top-0 left-0 right-0 h-0.5 bg-white/5 transition-opacity duration-300; }
.progress-fill { @apply h-full transition-all duration-300 ease-out; }

/* Transparent floating toolbar buttons */
.nav-btn {
  @apply absolute top-1/2 -translate-y-1/2 w-12 h-12 rounded-full;
  @apply bg-black/20 border border-white/10;
  @apply flex items-center justify-center text-white;
  @apply transition-all duration-300 pointer-events-auto;
  @apply hover:bg-black/40 hover:border-white/20;
  @apply disabled:opacity-30 disabled:cursor-not-allowed;
}

.nav-prev { @apply left-4; }
.nav-next { @apply right-4; }

.top-toolbar {
  @apply absolute top-0 left-0 right-0 h-16 bg-transparent;
  @apply flex items-center justify-end px-4;
  @apply transition-opacity duration-300 pointer-events-auto;
  transform: translate(-0.5rem, 0.5rem);
}

.toolbar-right { @apply flex items-center gap-2; }

.toolbar-btn { 
  @apply w-10 h-10 rounded-full flex items-center justify-center text-white transition-colors duration-150; 
  @apply bg-white/5 hover:bg-white/10 border border-white/20 backdrop-blur-sm;
}
.toolbar-btn.active { @apply ring-1 ring-white/40; }
</style>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { Play, StopCircle, ChevronLeft, ChevronRight, Maximize2, Minimize2, Clock, Pause } from 'lucide-vue-next'
import { useSlidesStore } from '../stores/slides'

const store = useSlidesStore()

const isFullscreen = ref(!!document.fullscreenElement)
const hideControls = ref(true)

const showPrevButton = computed(() => store.canGoPrev)
const showNextButton = computed(() => store.canGoNext)

// Auto-hide only when fullscreen or in presentation mode
const shouldAutoHide = computed(() => isFullscreen.value || store.isPresenting)

let hideTimer = null
function updateHideMode() {
  // If not auto-hide mode, keep controls visible and clear any timers
  if (!shouldAutoHide.value) {
    hideControls.value = false
    if (hideTimer) { clearTimeout(hideTimer); hideTimer = null }
  } else {
    // In auto-hide mode, default to hidden until user interaction
    hideControls.value = true
  }
}

function showControlsTemporarily() {
  hideControls.value = false
  if (hideTimer) clearTimeout(hideTimer)
  // Only schedule hide when auto-hide mode is active
  if (!shouldAutoHide.value) return
  hideTimer = setTimeout(() => {
    hideControls.value = true
  }, 5000) // 5s inactivity threshold
}

function onPointerProxy() {
  showControlsTemporarily()
}

function toggleFullscreen() {
  if (!document.fullscreenElement) {
    document.documentElement.requestFullscreen?.()
  } else {
    document.exitFullscreen?.()
  }
}

function toggleAutoplay() {
  const next = !store.config.settings.autoPlay
  store.config.settings.autoPlay = next
  if (next) {
    // Align with presentation mode: enter presentation and fullscreen
    if (!store.isPresenting) store.isPresenting = true
    if (!document.fullscreenElement) {
      document.documentElement.requestFullscreen?.()
        ?.catch?.((err) => console.warn('Failed to request fullscreen:', err))
    }
  }
}

let autoplayTimer = null
function startAutoplay() {
  stopAutoplay()
  const interval = store.config.settings.autoPlayInterval || 5000
  autoplayTimer = setInterval(() => {
    if (store.canGoNext) {
      store.nextSlide()
    } else {
      // Stop autoplay at last slide, exit fullscreen and exit presentation
      store.config.settings.autoPlay = false
      stopAutoplay()
      document.exitFullscreen?.()
      if (store.isPresenting) store.togglePresentation()
    }
  }, interval)
}
function stopAutoplay() {
  if (autoplayTimer) {
    clearInterval(autoplayTimer)
    autoplayTimer = null
  }
}

watch(() => store.config.settings.autoPlay, (val) => {
  if (val) startAutoplay()
  else stopAutoplay()
}, { immediate: true })

// React to fullscreen or presentation changes
watch(shouldAutoHide, () => updateHideMode(), { immediate: true })

function onFullscreenChange() {
  isFullscreen.value = !!document.fullscreenElement
}

onMounted(() => {
  document.addEventListener('slide-pointer-move', onPointerProxy)
  document.addEventListener('fullscreenchange', onFullscreenChange)
  updateHideMode()
})

onUnmounted(() => {
  document.removeEventListener('slide-pointer-move', onPointerProxy)
  document.removeEventListener('fullscreenchange', onFullscreenChange)
  stopAutoplay()
  if (hideTimer) { clearTimeout(hideTimer); hideTimer = null }
})
</script>
