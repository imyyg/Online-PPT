<template>
  <div id="app" :class="{ 'presentation-mode': store.isPresenting }">
    <!-- Sidebar -->
    <aside 
      v-if="!store.isPresenting"
      class="sidebar"
      :class="{ 'collapsed': store.isSidebarCollapsed }"
    >
      <div class="sidebar-header">
        <h1 class="sidebar-title">{{ store.config.title }}</h1>
        <button 
          @click="store.toggleSidebar" 
          class="sidebar-toggle"
        >
          <ChevronLeft v-if="!store.isSidebarCollapsed" class="w-5 h-5" />
          <ChevronRight v-else class="w-5 h-5" />
        </button>
      </div>
      
      <div class="sidebar-content">
        <div v-if="!store.isSidebarCollapsed" class="sidebar-section">
          <div class="thumbnails-list">
              <SlidePreview
                v-for="(slide, index) in store.visibleSlides"
                :key="slide.id"
                :slide="slide"
                :index="index"
                :is-active="store.currentIndex === index"
                :readonly="false"
                @select="store.goToSlide(index)"
                @duplicate="onDuplicateSlide(slide.id)"
                @delete="onRequestDelete(slide)"
                @reorder="onReorder($event)"
              />
            </div>
        </div>
      </div>
      
      <!-- Removed sidebar-footer actions (Manage Slides, Create New PPT) -->
    </aside>

    <!-- Main content -->
    <main class="main-content">
      <Transition :name="transitionName" mode="out-in">
        <div :key="store.currentIndex" class="slide-container">
          <SlideLoader
            v-if="store.currentSlide"
            :slide="store.currentSlide"
            :load-mode="loadMode"
            :is-fullscreen="store.isPresenting || isFullscreen"
          />
          <div v-else class="no-slides">
            <FileX v-if="!store.loadingConfig" class="w-16 h-16 text-gray-600 mb-4" />
            <div v-else class="loading-placeholder w-16 h-16 mb-4 rounded-full bg-gray-700 animate-pulse" />
            <p class="text-xl text-gray-400">
              {{ store.loadingConfig ? 'Loading slides…' : 'No slides available' }}
            </p>

          </div>
          <!-- Keep controls inside slide-container to align progress bar -->
          <PresentationControls v-if="store.totalSlides > 0" />
        </div>
      </Transition>
    </main>

    <!-- Delete Confirm Bubble -->
    <div v-if="showDeleteConfirm" class="modal-backdrop" @click.self="cancelDelete">
      <div class="modal-panel">
        <div class="p-5 flex items-center gap-3 border-b border-gray-700">
          <AlertTriangle class="w-5 h-5 text-red-500" />
          <h3 class="text-lg font-semibold">Confirm Deletion</h3>
        </div>
        <div class="p-6 space-y-2 text-sm">
          <p class="text-gray-200">Deletion cannot be undone.</p>
          <p class="text-gray-400">Confirm deletion of Slide <span class="font-semibold">{{ pendingDeleteNumber }}</span>?</p>
        </div>
        <div class="flex items-center justify-end gap-3 p-5 border-t border-gray-700">
          <button class="px-4 py-2 rounded-lg bg-gray-700 hover:bg-gray-600 text-gray-100" @click="cancelDelete">Cancel</button>
          <button class="px-4 py-2 rounded-lg bg-red-600 hover:bg-red-500 text-white flex items-center gap-2" @click="confirmDelete">
            <Trash2 class="w-4 h-4" /> Delete
          </button>
        </div>
      </div>
    </div>

  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, watch } from 'vue'
import { useSlidesStore } from './stores/slides'
import SlidePreview from './components/SlidePreview.vue'
import SlideLoader from './components/SlideLoader.vue'
import PresentationControls from './components/PresentationControls.vue'
import { ChevronLeft, ChevronRight, FileX, AlertTriangle, Trash2 } from 'lucide-vue-next'

const store = useSlidesStore()

const loadMode = ref('iframe')
const isFullscreen = ref(false)

const transitionPresets = ['slide','zoom','blur','flip','rotate','skew','fade','cover','push','cube','parallax','zoomfade','tilt']

// Restore slide-next/slide-prev transitions based on direction (with preset/random)
const transitionName = ref('slide-next')
let lastIndex = 0
function pickTransitionBase() {
  const t = store.config?.theme?.transition || 'slide'
  if (t === 'random') {
    return transitionPresets[Math.floor(Math.random() * transitionPresets.length)]
  }
  return transitionPresets.includes(t) ? t : 'slide'
}
watch(() => store.currentIndex, (newIndex) => {
  const dir = newIndex > lastIndex ? 'next' : 'prev'
  const base = pickTransitionBase()
  transitionName.value = `${base}-${dir}`
  lastIndex = newIndex
})



function onFullscreenChange() {
  isFullscreen.value = !!document.fullscreenElement
}

const numberBuffer = ref('')
let numberBufferTimer = null
function resetNumberBuffer() {
  numberBuffer.value = ''
  if (numberBufferTimer) { clearTimeout(numberBufferTimer); numberBufferTimer = null }
}
function commitNumberBuffer() {
  const n = parseInt(numberBuffer.value, 10)
  if (!Number.isNaN(n) && n > 0) {
    const target = Math.min(Math.max(n - 1, 0), store.totalSlides - 1)
    store.goToSlide(target)
  }
  resetNumberBuffer()
}
function scheduleCommit() {
  if (numberBufferTimer) clearTimeout(numberBufferTimer)
  numberBufferTimer = setTimeout(commitNumberBuffer, 800)
}

function onKeydown(e) {
  // Respect setting and avoid interfering with inputs or modals
  if (!store.config.settings.enableKeyboardNav) return
  const tag = (e.target && e.target.tagName) || ''
  const isInput = ['INPUT','TEXTAREA','SELECT'].includes(tag)
  const isEditable = e.target && e.target.isContentEditable
  if (isInput || isEditable) return
  if (showDeleteConfirm.value) return

  // Numeric jump: accumulate digits then jump after short delay
  if (e.key >= '0' && e.key <= '9') {
    e.preventDefault()
    numberBuffer.value += e.key
    scheduleCommit()
    return
  }

  switch (e.key) {
    case 'ArrowRight':
      e.preventDefault()
      store.nextSlide()
      break
    case 'ArrowLeft':
      e.preventDefault()
      store.prevSlide()
      break
    case ' ': // Space: next slide
      e.preventDefault()
      store.nextSlide()
      break
    case 'Enter': // immediate commit numeric buffer
      if (numberBuffer.value) {
        e.preventDefault()
        commitNumberBuffer()
      }
      break
    case 'Escape':
      e.preventDefault()
      // Stop autoplay and exit fullscreen/presentation
      store.config.settings.autoPlay = false
      if (document.fullscreenElement) {
        document.exitFullscreen?.()
      } else if (store.isPresenting) {
        store.togglePresentation()
      }
      break
  }
}

// Listen to ESC bridge from SlideLoader content
function onEscBridge() {
  const evt = new KeyboardEvent('keydown', { key: 'Escape' })
  onKeydown(evt)
}

onMounted(async () => {
  // Ensure we are on example group by default, load its config
  if (!store.currentGroup) store.setGroup('example')
  await store.loadConfig()
  document.addEventListener('fullscreenchange', onFullscreenChange)
  document.addEventListener('keydown', onKeydown)
  document.addEventListener('slide-esc', onEscBridge)
})

onUnmounted(() => {
  document.removeEventListener('fullscreenchange', onFullscreenChange)
  document.removeEventListener('keydown', onKeydown)
  document.removeEventListener('slide-esc', onEscBridge)
  if (numberBufferTimer) { clearTimeout(numberBufferTimer); numberBufferTimer = null }
})



// 排序：支持 before/after 和持久化
function onReorder({ from, to, position }) {
  const visible = store.visibleSlides
  const toIndex = position === 'before' ? to : to + 1
  // Map visible indices to absolute indices in config.slides
  const mapVisibleToAbsolute = (visIdx) => {
    let count = -1
    for (let i = 0; i < store.config.slides.length; i++) {
      if (store.config.slides[i].visible) {
        count++
        if (count === visIdx) return i
      }
    }
    return -1
  }
  const absFrom = mapVisibleToAbsolute(from)
  const absTo = mapVisibleToAbsolute(toIndex)
  if (absFrom === -1 || absTo === -1) return

  // Frontend reorder for instant feedback
  store.reorderSlides(absFrom, absTo)

  // Persist to backend
  fetch('/api/slides/reorder', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ from: absFrom, to: absTo, group: store.currentGroup })
  })
    .then(r => r.json())
    .then(async (json) => {
      if (!json.ok) {
        console.error('Reorder failed:', json.error)
        return
      }
      await store.loadConfig()
    })
    .catch(e => console.error('Persist reorder error:', e))
}

// 添加：复制到末尾
function onDuplicateSlide(slideId) {
  const duplicated = store.duplicateSlide(slideId)
  if (duplicated) {
    // 先在前端把复制的页面移动到末尾，提高即时反馈
    const from = store.visibleSlides.findIndex(s => s.id === duplicated.id)
    const to = store.totalSlides - 1
    if (from !== -1 && to !== -1 && from !== to) {
      store.reorderSlides(from, to)
    }

    // 调用后端实际复制文件到分组目录
    fetch('/api/slides/duplicate', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        sourceFile: duplicated.file.replace(/-copy-\d+\.html$/, '.html'),
        sourceTitle: duplicated.title.replace(/ \(Copy\)$/, ''),
        group: store.currentGroup
      })
    })
      .then(r => r.json())
      .then(async (json) => {
        if (!json.ok) {
          console.error('Duplicate slide failed:', json.error)
          return
        }
        // 重新加载配置以与磁盘同步，然后再次将副本移动到末尾（仅前端持久）
        await store.loadConfig()
        const newIndex = store.visibleSlides.findIndex(s => s.file === duplicated.file || s.title === duplicated.title)
        const last = store.totalSlides - 1
        if (newIndex !== -1 && newIndex !== last) {
          store.reorderSlides(newIndex, last)
        }
      })
      .catch(e => console.error('Duplicate slide error:', e))
  }
}

// 添加：删除（英文二次确认）
const showDeleteConfirm = ref(false)
const pendingDelete = ref(null)
const pendingDeleteNumber = ref(null)

function onRequestDelete(slide) {
  pendingDelete.value = slide
  const idx = store.visibleSlides.findIndex(s => s.id === slide.id)
  pendingDeleteNumber.value = idx !== -1 ? (idx + 1) : null
  showDeleteConfirm.value = true
}

function cancelDelete() {
  pendingDelete.value = null
  pendingDeleteNumber.value = null
  showDeleteConfirm.value = false
}

async function confirmDelete() {
  const s = pendingDelete.value
  if (!s) return
  showDeleteConfirm.value = false

  // 先前端隐藏，提升即时反馈
  store.removeSlide(s.id)

  try {
    const res = await fetch('/api/slides/delete', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ id: s.id, file: s.file, group: store.currentGroup })
    })
    const json = await res.json()
    if (!json.ok) {
      console.error('Delete slide failed:', json.error)
      return
    }
    await store.loadConfig()
    if (store.currentIndex >= store.totalSlides) {
      store.goToSlide(Math.max(0, store.totalSlides - 1))
    }
  } catch (e) {
    console.error('Delete slide error:', e)
  } finally {
    pendingDelete.value = null
  }
}

onMounted(async () => {
  // Ensure we are on example group by default, load its config
  if (!store.currentGroup) store.setGroup('example')
  await store.loadConfig()
  document.addEventListener('fullscreenchange', onFullscreenChange)
  document.addEventListener('keydown', onKeydown)
})

onUnmounted(() => {
  document.removeEventListener('fullscreenchange', onFullscreenChange)
  document.removeEventListener('keydown', onKeydown)
  if (numberBufferTimer) { clearTimeout(numberBufferTimer); numberBufferTimer = null }
})
</script>

<style scoped>
@reference "./tw.css";
#app { @apply h-screen flex bg-gray-900 text-gray-100 overflow-hidden; }

.presentation-mode .sidebar { @apply hidden; }

/* Adjust sidebar width */
.sidebar { width: calc(var(--spacing, 0.25rem) * 75); @apply bg-gray-800 border-r border-gray-700 flex flex-col transition-all duration-300; }
/* Collapse width to minimal footprint */
.sidebar.collapsed { width: calc(var(--spacing, 0.25rem) * 8); }

.sidebar-header { padding-block: calc(var(--spacing, 0.25rem) * 2); @apply flex items-center justify-between px-4 border-b border-gray-700; }
/* Compact header in collapsed mode */
.sidebar.collapsed .sidebar-header { @apply justify-center p-2; }
.sidebar-title { @apply text-lg font-semibold truncate; }
.sidebar.collapsed .sidebar-title { @apply hidden; }
/* Smaller toggle in collapsed mode */
.sidebar-toggle { @apply p-2 rounded hover:bg-gray-700 transition-colors; }
.sidebar.collapsed .sidebar-toggle { @apply w-8 h-8 flex items-center justify-center; }

/* Reduce sidebar content padding */
.sidebar-content { padding: calc(var(--spacing, 0.25rem) * 2); @apply flex-1 overflow-y-auto; }
.sidebar-section { @apply mb-6; }

.section-title { /* removed: element no longer rendered */ }

.thumbnails-list { @apply space-y-3; }


.sidebar-footer { @apply p-4 border-t border-gray-700; }
.sidebar-action { @apply w-full flex items-center justify-center px-4 py-2 bg-gray-700 hover:bg-gray-600 rounded-lg transition-colors; }

.main-content { @apply flex-1 relative overflow-hidden; }

.slide-container { padding: calc(var(--spacing, 0.25rem) * 3); @apply w-full h-full flex items-center justify-center; }

.presentation-mode .slide-container { @apply p-0; }

.no-slides { @apply text-center; }

/* Scrollbar styling for sidebar */
.sidebar-content { 
  scrollbar-width: thin; /* Firefox */
  scrollbar-color: rgba(255,255,255,0.2) transparent;
}
.sidebar-content::-webkit-scrollbar {
  width: 8px;
}
.sidebar-content::-webkit-scrollbar-track {
  background: transparent;
}
.sidebar-content::-webkit-scrollbar-thumb {
  background-color: rgba(255,255,255,0.2);
  border-radius: 9999px;
  border: 2px solid transparent;
}

.modal-backdrop { @apply fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center p-4 z-50; }

.modal-panel { @apply bg-gray-800 rounded-xl w-full max-w-sm md:max-w-md shadow-2xl border border-gray-700 overflow-hidden flex flex-col; }

.modal-header { @apply flex items-center justify-between p-6 border-b border-gray-700; }

.close-btn { @apply p-2 rounded hover:bg-gray-700 transition-colors; }

/* Slide transitions */
.slide-next-enter-active,
.slide-next-leave-active,
.slide-prev-enter-active,
.slide-prev-leave-active {
  transition: all 0.3s ease-out;
}

.slide-next-enter-from { transform: translateX(100%); opacity: 0; }
.slide-next-leave-to { transform: translateX(-100%); opacity: 0; }

.slide-prev-enter-from { transform: translateX(-100%); opacity: 0; }
.slide-prev-leave-to { transform: translateX(100%); opacity: 0; }
</style>