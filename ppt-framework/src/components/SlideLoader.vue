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

    <!-- Save button overlay when there are unsaved changes -->
    <button v-if="isDirty" class="save-btn" @click="onSaveClick">保存</button>
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
const isDirty = ref(false)

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
let reloadTimer = null

// Inject: inline text editing and Shift+drag repositioning
let removeEditingListeners = null
const editorState = {
  container: null,
  editingEl: null,
  dragging: false,
  dragEl: null,
  startX: 0,
  startY: 0,
  baseX: 0,
  baseY: 0,
  // Added for edge-hover drag and snapping
  edgeHoverEl: null,
  baseRect: null,
  containerRect: null,
  width: 0,
  height: 0,
  guidesOverlay: null,
  hasMoved: false
}
const SNAP_THRESHOLD = 8
const EDGE_THRESHOLD = 6
const MOVE_THRESHOLD = 3
let lastFetchedHtml = ''
let lastFetchedDoc = null
let shadowBody = null

function ensureEditorStyles(rootOrDoc) {
  try {
    const styleId = 'slide-editor-style'
    const exists = rootOrDoc.getElementById?.(styleId)
    if (exists) return
    const styleEl = (rootOrDoc.createElement ? rootOrDoc.createElement('style') : document.createElement('style'))
    styleEl.id = styleId
    styleEl.textContent = `
      .editable-active { outline: 2px dashed #60a5fa; outline-offset: 2px; }
      .draggable-activated { cursor: move !important; }
      .drag-managed { will-change: transform; }
      .dragging-global { user-select: none !important; }
      .edge-drag-cursor { cursor: grab !important; }
      .edge-drag-active { cursor: grabbing !important; }
      .guides-overlay { position: absolute; inset: 0; pointer-events: none; }
      .guide-line-x, .guide-line-y { position: absolute; background: rgba(96,165,250,0.6); }
      .guide-line-x { width: 100%; height: 1px; left: 0; }
      .guide-line-y { height: 100%; width: 1px; top: 0; }
    `
    if (rootOrDoc.head) rootOrDoc.head.appendChild(styleEl)
    else if (rootOrDoc.appendChild) rootOrDoc.appendChild(styleEl)
  } catch {}
}

function findEditableTarget(el, container) {
  const allowed = ['H1','H2','H3','H4','H5','H6','P','SPAN','DIV','LI','DT','DD','EM','STRONG']
  let cur = el
  while (cur && cur !== container) {
    const tag = cur.tagName || ''
    const hasText = (cur.innerText || '').trim().length > 0
    if (allowed.includes(tag) && hasText) return cur
    cur = cur.parentNode
  }
  return null
}

function ensureGuidesOverlay(containerRoot) {
  try {
    const host = containerRoot.body ? containerRoot.body : containerRoot
    if (!editorState.guidesOverlay) {
      const el = (host.ownerDocument || document).createElement('div')
      el.className = 'guides-overlay'
      host.appendChild(el)
      editorState.guidesOverlay = el
    }
  } catch {}
}
function showGuides(y, x) {
  const overlay = editorState.guidesOverlay
  if (!overlay) return
  overlay.innerHTML = ''
  const doc = overlay.ownerDocument || document
  if (typeof y === 'number') {
    const h = doc.createElement('div')
    h.className = 'guide-line-x'
    h.style.top = `${y}px`
    overlay.appendChild(h)
  }
  if (typeof x === 'number') {
    const v = doc.createElement('div')
    v.className = 'guide-line-y'
    v.style.left = `${x}px`
    overlay.appendChild(v)
  }
}
function clearGuides() { if (editorState.guidesOverlay) editorState.guidesOverlay.innerHTML = '' }

function makeEditable(target) {
  if (!target) return
  if (editorState.editingEl && editorState.editingEl !== target) finishEditing()
  try {
    target.setAttribute('contenteditable', 'true')
    target.classList.add('editable-active')
    const onInput = () => { isDirty.value = true }
    const onBlur = () => { finishEditing(); try { target.removeEventListener('input', onInput) } catch {} }
    target.addEventListener('input', onInput)
    target.addEventListener('blur', onBlur, { once: true })
    target.focus()
  } catch {}
  editorState.editingEl = target
}

function finishEditing() {
  const el = editorState.editingEl
  if (!el) return
  try {
    el.removeAttribute('contenteditable')
    el.classList.remove('editable-active')
  } catch {}
  editorState.editingEl = null
}

function toggleSelection(containerRoot, on) {
  try {
    const host = containerRoot.body ? containerRoot.body : containerRoot
    if (on) host.classList.add('dragging-global')
    else host.classList.remove('dragging-global')
  } catch {}
}

function detectEdgeTarget(containerRoot, e) {
  const t = findEditableTarget(e.target, containerRoot)
  if (!t) return null
  const rect = t.getBoundingClientRect()
  const nearLeft = Math.abs(e.clientX - rect.left) <= EDGE_THRESHOLD
  const nearRight = Math.abs(e.clientX - rect.right) <= EDGE_THRESHOLD
  const nearTop = Math.abs(e.clientY - rect.top) <= EDGE_THRESHOLD
  const nearBottom = Math.abs(e.clientY - rect.bottom) <= EDGE_THRESHOLD
  if (nearLeft || nearRight || nearTop || nearBottom) return t
  return null
}

function startDrag(target, startX, startY, containerRoot) {
  editorState.dragging = true
  editorState.dragEl = target
  editorState.startX = startX
  editorState.startY = startY
  editorState.hasMoved = false
  const tx = parseFloat(target.style.getPropertyValue('--drag-tx') || '0') || 0
  const ty = parseFloat(target.style.getPropertyValue('--drag-ty') || '0') || 0
  editorState.baseX = tx
  editorState.baseY = ty
  target.classList.add('edge-drag-active','drag-managed')
  // Cache rects for snapping
  editorState.baseRect = target.getBoundingClientRect()
  editorState.width = editorState.baseRect.width
  editorState.height = editorState.baseRect.height
  editorState.containerRect = (containerRoot.body ? containerRoot.body : containerRoot).getBoundingClientRect()
  ensureGuidesOverlay(containerRoot)
  toggleSelection(containerRoot, true)
}

function applySnap(dx, dy) {
  const base = editorState.baseRect
  const cont = editorState.containerRect
  if (!base || !cont) return { dx, dy }
  let ndx = dx, ndy = dy
  let guideX = null, guideY = null
  const newLeft = base.left + dx
  const newTop = base.top + dy
  const newCenterX = newLeft + editorState.width / 2
  const newCenterY = newTop + editorState.height / 2
  const contCenterX = cont.left + cont.width / 2
  const contCenterY = cont.top + cont.height / 2
  // Snap to container center lines
  if (Math.abs(newCenterX - contCenterX) <= SNAP_THRESHOLD) { ndx += (contCenterX - newCenterX); guideX = contCenterX - cont.left }
  if (Math.abs(newCenterY - contCenterY) <= SNAP_THRESHOLD) { ndy += (contCenterY - newCenterY); guideY = contCenterY - cont.top }
  // Snap to container edges
  if (Math.abs(newLeft - cont.left) <= SNAP_THRESHOLD) { ndx += (cont.left - newLeft); guideX = 0 }
  const newRight = newLeft + editorState.width
  if (Math.abs(newRight - cont.right) <= SNAP_THRESHOLD) { ndx += (cont.right - newRight); guideX = cont.width }
  if (Math.abs(newTop - cont.top) <= SNAP_THRESHOLD) { ndy += (cont.top - newTop); guideY = 0 }
  const newBottom = newTop + editorState.height
  if (Math.abs(newBottom - cont.bottom) <= SNAP_THRESHOLD) { ndy += (cont.bottom - newBottom); guideY = cont.height }
  return { dx: ndx, dy: ndy, guideX, guideY }
}

function updateDrag(clientX, clientY) {
  if (!editorState.dragging || !editorState.dragEl) return
  const dx = clientX - editorState.startX
  const dy = clientY - editorState.startY
  if (Math.abs(dx) + Math.abs(dy) > MOVE_THRESHOLD) editorState.hasMoved = true
  const snap = applySnap(dx, dy)
  const nx = editorState.baseX + snap.dx
  const ny = editorState.baseY + snap.dy
  try {
    editorState.dragEl.style.setProperty('--drag-tx', `${nx}px`)
    editorState.dragEl.style.setProperty('--drag-ty', `${ny}px`)
    editorState.dragEl.style.transform = `translate(var(--drag-tx, 0px), var(--drag-ty, 0px))`
  } catch {}
  showGuides(snap.guideY, snap.guideX)
}

function endDrag(containerRoot) {
  if (!editorState.dragEl) { editorState.dragging = false; return }
  try { editorState.dragEl.classList.remove('edge-drag-active') } catch {}
  editorState.dragging = false
  editorState.dragEl = null
  editorState.baseRect = null
  clearGuides()
  toggleSelection(containerRoot, false)
  // Mark dirty only if moved; do not save immediately
  if (editorState.hasMoved) isDirty.value = true
}

function attachEditingFeatures(containerRoot) {
  if (!containerRoot) return
  editorState.container = containerRoot
  ensureEditorStyles(containerRoot instanceof ShadowRoot ? containerRoot : (containerRoot.ownerDocument || containerRoot))
  const onDblClick = (e) => {
    const target = findEditableTarget(e.target, containerRoot)
    if (target) { e.preventDefault(); makeEditable(target) }
  }
  const onKeyDown = (e) => {
    if (e.key === 'Escape' && editorState.editingEl) { e.preventDefault(); finishEditing() }
    if (e.key === 'Escape') { try { window.top?.document?.dispatchEvent(new CustomEvent('slide-esc')) } catch {} }
  }
  const onMouseMove = (e) => {
    if (editorState.dragging) { updateDrag(e.clientX, e.clientY); return }
    const t = detectEdgeTarget(containerRoot, e)
    if (t !== editorState.edgeHoverEl) {
      try { editorState.edgeHoverEl?.classList.remove('edge-drag-cursor') } catch {}
      editorState.edgeHoverEl = t
      try { t?.classList.add('edge-drag-cursor') } catch {}
    }
  }
  const onMouseDown = (e) => {
    if (e.button !== 0) return
    // Only start drag when hovering near edge
    if (!editorState.edgeHoverEl) return
    const t = editorState.edgeHoverEl
    e.preventDefault()
    startDrag(t, e.clientX, e.clientY, containerRoot)
  }
  const onMouseUp = (e) => { if (editorState.dragging) endDrag(containerRoot) }
  containerRoot.addEventListener('dblclick', onDblClick)
  containerRoot.addEventListener('keydown', onKeyDown)
  containerRoot.addEventListener('mousemove', onMouseMove)
  containerRoot.addEventListener('mousedown', onMouseDown)
  containerRoot.addEventListener('mouseup', onMouseUp)
  try { containerRoot.host?.setAttribute('tabindex', '0'); containerRoot.host?.focus?.() } catch {}
  removeEditingListeners = () => {
    containerRoot.removeEventListener('dblclick', onDblClick)
    containerRoot.removeEventListener('keydown', onKeyDown)
    containerRoot.removeEventListener('mousemove', onMouseMove)
    containerRoot.removeEventListener('mousedown', onMouseDown)
    containerRoot.removeEventListener('mouseup', onMouseUp)
  }
}

function attachEditingToIframe() {
  try {
    const frameDoc = slideFrame.value?.contentDocument
    if (!frameDoc) return
    attachEditingFeatures(frameDoc)
  } catch {}
}
function attachEditingToShadow() { try { if (!shadowRoot) return; attachEditingFeatures(shadowRoot) } catch {} }

async function saveCurrentHtml(fileOverride) {
  try {
    let html = ''
    if (props.loadMode === 'iframe') {
      const doc = slideFrame.value?.contentDocument
      if (!doc) return
      // Clone document for sanitization to avoid affecting live editing styles
      const cloneHtmlEl = doc.documentElement.cloneNode(true)
      // Remove injected editor style and overlays
      try { cloneHtmlEl.querySelector('#slide-editor-style')?.remove() } catch {}
      try { cloneHtmlEl.querySelector('.guides-overlay')?.remove() } catch {}
      // Strip contenteditable and temp editing classes
      try {
        cloneHtmlEl.querySelectorAll('[contenteditable]')?.forEach(el => {
          el.removeAttribute('contenteditable')
          el.classList.remove('editable-active','edge-drag-cursor','edge-drag-active','draggable-activated','drag-managed')
        })
        cloneHtmlEl.querySelector('body')?.classList.remove('dragging-global')
      } catch {}
      const doctype = '<!DOCTYPE html>'
      html = doctype + '\n' + cloneHtmlEl.outerHTML
    } else {
      if (!lastFetchedDoc || !shadowBody) return
      const clone = lastFetchedDoc.cloneNode(true)
      try { clone.getElementById('slide-editor-style')?.remove() } catch {}
      try { clone.body.innerHTML = shadowBody.innerHTML } catch {}
      // Strip contenteditable and temp editing classes
      try {
        clone.querySelectorAll('[contenteditable]')?.forEach(el => {
          el.removeAttribute('contenteditable')
          el.classList.remove('editable-active','edge-drag-cursor','edge-drag-active','draggable-activated','drag-managed')
        })
        clone.querySelector('body')?.classList.remove('dragging-global')
      } catch {}
      const doctype = '<!DOCTYPE html>'
      html = doctype + '\n' + clone.documentElement.outerHTML
    }
    const filePath = fileOverride || slidePath.value
    await fetch('/api/slides/save', {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ file: filePath, group: store.currentGroup, html })
    })
  } catch (e) {
    console.error('Auto-save failed:', e)
  }
}

async function onSaveClick() {
  await saveCurrentHtml()
  isDirty.value = false
}

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
    if (!response.ok) { throw new Error(`HTTP error! status: ${response.status}`) }
    const html = await response.text()
    lastFetchedHtml = html
    const parser = new DOMParser()
    const doc = parser.parseFromString(html, 'text/html')
    lastFetchedDoc = doc
    if (!shadowRoot && shadowHost.value) { shadowRoot = shadowHost.value.attachShadow({ mode: 'open' }) }
    shadowRoot.innerHTML = ''
    const styles = doc.querySelectorAll('style, link[rel="stylesheet"]')
    for (const style of styles) { shadowRoot.appendChild(style.cloneNode(true)) }
    const bodyContent = doc.body.cloneNode(true)
    shadowBody = bodyContent
    try {
      bodyContent.style.margin = '0'; bodyContent.style.padding = '0'; bodyContent.style.height = '100vh'; bodyContent.style.overflow = 'hidden'; bodyContent.style.display = 'flex'; bodyContent.style.alignItems = 'center'; bodyContent.style.justifyContent = 'center'
    } catch {}
    shadowRoot.appendChild(bodyContent)
    attachEditingToShadow()
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
  } catch (e) { console.debug('Cross-origin restriction when accessing iframe content') }
  attachEditingToIframe()
  if (props.enablePointerProxy) attachIframeMouseProxy()
}

function onFrameError(error) {
  console.warn('Frame loading error (可能是因快速重载导致的中止):', error)
  // 延迟标记错误，避免因 debounce 重载导致的中止误报
  setTimeout(() => {
    if (loading.value) {
      loadingError.value = 'Failed to load slide content'
      loading.value = false
    }
  }, 300)
}

function forceIframeReload() {
  if (!slideFrame.value) return
  try {
    const url = slideUrl.value
    const glue = url.includes('?') ? '&' : '?'
    const newUrl = `${url}${glue}v=${Date.now()}`
    if (reloadTimer) { try { clearTimeout(reloadTimer) } catch {} }
    reloadTimer = setTimeout(() => {
      try { slideFrame.value.src = newUrl } catch {}
      reloadTimer = null
    }, 120)
  } catch (e) {
    // ignore
  }
}

// Watch for slide file path changes (减少无谓重载) and save previous if dirty
watch(() => slidePath.value, async (newPath, oldPath) => {
  // Save old slide first if there are unsaved changes
  if (oldPath && isDirty.value) {
    await saveCurrentHtml(oldPath)
    isDirty.value = false
  }
  loading.value = true
  loadingError.value = ''
  
  if (props.loadMode === 'shadow') {
    await loadShadowContent()
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

onUnmounted(async () => {
  try {
    if (isDirty.value) { await saveCurrentHtml(); isDirty.value = false }
  } catch {}
  shadowRoot = null
  if (removeIframeMouseListener) removeIframeMouseListener()
  if (removeShadowMouseListener) removeShadowMouseListener()
  if (removeEditingListeners) removeEditingListeners()
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

.save-btn {
  position: absolute;
  right: 1rem;
  bottom: 1rem;
  z-index: 10;
  padding: 0.5rem 0.9rem;
  border-radius: 0.5rem;
  background: rgba(96,165,250,0.25);
  color: white;
  border: 1px solid rgba(96,165,250,0.45);
  backdrop-filter: blur(4px);
}
.save-btn:hover { background: rgba(96,165,250,0.35); }
</style>
