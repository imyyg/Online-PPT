<template>
  <div 
    class="slide-preview group"
    :class="{ 
      'active': isActive,
      'dragging': isDragging,
      'readonly': readonly
    }"
    @click="$emit('select')"
    @dragstart="handleDragStart"
    @dragend="handleDragEnd"
    @dragover.prevent="handleDragOver"
    @dragenter="handleDragEnter"
    @dragleave="handleDragLeave"
    @drop="handleDrop"
    :draggable="!readonly"
  >
    <div class="preview-number">{{ index + 1 }}</div>

    <div class="preview-content">
      <div class="preview-scale">
        <SlideLoader
          :slide="slide"
          :load-mode="'iframe'"
          :is-fullscreen="false"
          :enable-pointer-proxy="false"
        />
      </div>
      <div class="preview-overlay" />

      <div v-if="!readonly" class="hover-actions">
        <button 
          @click.stop="$emit('duplicate')" 
          class="hover-action-btn" 
          title="Duplicate slide"
        >
          <Copy class="w-4 h-4" />
        </button>
        <button 
          @click.stop="$emit('delete')" 
          class="hover-action-btn hover-action-danger" 
          title="Delete this slide (physically deletes HTML file)"
        >
          <Trash2 class="w-4 h-4" />
        </button>
      </div>
    </div>


  </div>
</template>

<script setup>
import { ref } from 'vue'
import { Copy, Trash2 } from 'lucide-vue-next'
import SlideLoader from './SlideLoader.vue'

const props = defineProps({
  slide: {
    type: Object,
    required: true
  },
  index: {
    type: Number,
    required: true
  },
  isActive: {
    type: Boolean,
    default: false
  },
  readonly: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['select', 'duplicate', 'delete', 'reorder'])

const isDragging = ref(false)
const isDragOver = ref(false)
const insertBefore = ref(true)

function handleDragStart(event) {
  isDragging.value = true
  event.dataTransfer.effectAllowed = 'move'
  event.dataTransfer.setData('slideIndex', props.index.toString())
}

function handleDragEnd() {
  isDragging.value = false
  isDragOver.value = false
}

function handleDragEnter(event) {
  isDragOver.value = true
  const rect = event.currentTarget.getBoundingClientRect()
  const y = event.clientY - rect.top
  const ratio = y / rect.height
  const prev = insertBefore.value
  insertBefore.value = ratio < 0.45 ? true : (ratio > 0.55 ? false : prev)
}

function handleDragLeave() {
  isDragOver.value = false
}

function handleDragOver(event) {
  // Decide insert position by pointer Y with hysteresis to avoid jitter
  const rect = event.currentTarget.getBoundingClientRect()
  const y = event.clientY - rect.top
  const ratio = y / rect.height
  const prev = insertBefore.value
  insertBefore.value = ratio < 0.45 ? true : (ratio > 0.55 ? false : prev)
  event.dataTransfer.dropEffect = 'move'
}

function handleDrop(event) {
  event.preventDefault()
  isDragOver.value = false
  const fromIndex = Number.parseInt(event.dataTransfer.getData('slideIndex'))
  if (fromIndex !== props.index) {
    emit('reorder', { from: fromIndex, to: props.index, position: insertBefore.value ? 'before' : 'after' })
  }
}
</script>

<style scoped>
@reference "../tw.css";
.slide-preview {
  @apply relative bg-gray-800 rounded-lg overflow-hidden cursor-pointer transition-all duration-200;
  @apply border-2 border-transparent hover:border-gray-600;
}

.slide-preview.active { @apply border-primary-500 shadow-lg shadow-primary-500/20; }
.slide-preview.dragging { @apply opacity-50; }

.preview-number { @apply absolute top-2 left-2 w-6 h-6 bg-gray-700 rounded-full flex items-center justify-center text-xs font-medium z-10; }

.preview-content { @apply relative w-full aspect-video bg-gray-900 overflow-hidden; }

.preview-scale { @apply absolute inset-0; transform: scale(0.25); transform-origin: top left; width: 400%; height: 400%; }

.preview-overlay { @apply absolute inset-0 bg-transparent; }

.hover-actions { @apply absolute bottom-2 right-2 flex gap-1 opacity-0 transition-opacity; }
.slide-preview:hover .hover-actions { @apply opacity-100; }
.hover-action-btn { @apply p-1.5 rounded bg-gray-700/80 hover:bg-gray-600/80 border border-white/10 backdrop-blur-sm; }
.hover-action-danger { @apply hover:bg-red-600/80; }

.preview-info { @apply p-3; }
.preview-title { @apply text-sm font-medium text-gray-200 truncate; }
</style>
