<template>
  <aside class="left-menu" v-if="!store.isPresenting">
    <div class="menu-inner">
      <button class="toolbar-btn btn-user" @click="showDevToast" title="User Center">
        <User class="w-5 h-5 icon" />
      </button>
      <button class="toolbar-btn btn-settings" @click="toggleSettings" :class="{ active: showSettings }" title="Settings">
        <Settings class="w-5 h-5 icon" />
      </button>
      <button class="toolbar-btn btn-add" @click="openCreateModal" title="New PPT">
        <PlusSquare class="w-5 h-5 icon" />
      </button>
    </div>

    <!-- Centered popup for dev notice -->
    <div v-if="showToast" class="fixed inset-0 z-50 flex items-center justify-center">
      <div class="px-3 py-2 text-sm bg-gray-800 border border-gray-700 rounded-lg text-gray-100 shadow-xl">
        {{ toastMessage }}
      </div>
    </div>

    <!-- Create PPT Modal -->
    <div v-if="showCreate" class="fixed inset-0 z-50 flex items-center justify-center" @click.self="closeCreate">
      <div class="bg-gray-800 border border-gray-700 rounded-xl shadow-2xl w-full max-w-sm overflow-hidden">
        <div class="flex items-center justify-between p-4 border-b border-gray-700">
          <h3 class="text-lg font-semibold">New PPT</h3>
          <button class="w-8 h-8 rounded-lg bg-white/5 hover:bg-white/10 border border-white/20" @click="closeCreate">✕</button>
        </div>
        <form @submit.prevent="createPpt" class="p-4 space-y-3">
          <label class="flex items-center justify-between gap-3">
            <span>name</span>
            <input v-model="createForm.name" type="text" class="w-36 px-2 py-1 bg-gray-700 border border-gray-600 rounded-lg text-gray-100" placeholder="my-ppt" pattern="[a-zA-Z0-9-_]+" required />
          </label>
          <label class="flex items-center justify-between gap-3">
            <span>title</span>
            <input v-model="createForm.title" type="text" class="w-36 px-2 py-1 bg-gray-700 border border-gray-600 rounded-lg text-gray-100" placeholder="Presentation Title" />
          </label>
          <label class="flex items-center justify-between gap-3">
            <span>description</span>
            <input v-model="createForm.description" type="text" class="w-36 px-2 py-1 bg-gray-700 border border-gray-600 rounded-lg text-gray-100" placeholder="Presentation Description" />
          </label>
          <div class="flex items-center justify-end gap-2 pt-2">
            <button type="button" class="w-24 px-3 py-2 rounded-lg bg-gray-700 hover:bg-gray-600 text-gray-100" @click="closeCreate">Cancel</button>
            <button type="submit" class="w-24 px-3 py-2 rounded-lg bg-primary-600 hover:bg-primary-500 text-white">Create</button>
          </div>
        </form>
      </div>
    </div>

    <div v-if="showSettings" class="settings-panel" @click.self="showSettings=false">
      <div class="panel">
        <div class="panel-header">
          <h3 class="text-lg font-semibold">Settings</h3>
          <button class="close-btn" @click="showSettings=false">✕</button>
        </div>
        <div class="panel-body" @focusout="scheduleAutoSave">
          <label class="form-row">
            <span>Autoplay</span>
            <input type="checkbox" v-model="store.config.settings.autoPlay" @change="saveSettings" />
          </label>
          <label class="form-row">
            <span>Autoplay Interval (ms)</span>
            <input type="number" min="500" step="500" v-model.number="store.config.settings.autoPlayInterval" @blur="saveSettings" />
          </label>
          <label class="form-row">
            <span>Loop Slides</span>
            <input type="checkbox" v-model="store.config.settings.loop" @change="saveSettings" />
          </label>
          <label class="form-row">
            <span>Show Progress</span>
            <input type="checkbox" v-model="store.config.settings.showProgress" @change="saveSettings" />
          </label>
          <label class="form-row">
            <span>Show Thumbnails</span>
            <input type="checkbox" v-model="store.config.settings.showThumbnails" @change="saveSettings" />
          </label>
          <label class="form-row">
            <span>Keyboard Navigation</span>
            <input type="checkbox" v-model="store.config.settings.enableKeyboardNav" @change="saveSettings" />
          </label>
          <label class="form-row">
            <span>Touch Navigation</span>
            <input type="checkbox" v-model="store.config.settings.enableTouchNav" @change="saveSettings" />
          </label>
          <label class="form-row">
            <span>Auto Start on Home</span>
            <input type="checkbox" v-model="store.config.settings.autoStartOnHome" @change="saveSettings" />
          </label>
          <label class="form-row">
            <span>Auto Fullscreen on Home</span>
            <input type="checkbox" v-model="store.config.settings.autoFullscreenOnHome" @change="saveSettings" />
          </label>

          <div class="form-row">
            <span>Transition</span>
            <select v-model="store.config.theme.transition" @change="saveSettings">
              <option value="slide">Slide</option>
              <option value="zoom">Zoom</option>
              <option value="blur">Blur</option>
              <option value="flip">Flip</option>
              <option value="rotate">Rotate</option>
              <option value="skew">Skew</option>
              <option value="fade">Fade</option>
              <option value="cover">Cover</option>
              <option value="push">Push</option>
              <option value="cube">Cube</option>
              <option value="parallax">Parallax</option>
              <option value="zoomfade">Zoomfade</option>
              <option value="tilt">Tilt</option>
              <option value="random">Random</option>
            </select>
          </div>
          <label class="form-row">
            <span>Primary Color</span>
            <input type="color" v-model="store.config.theme.primaryColor" @input="saveSettings" />
          </label>
          <label class="form-row">
            <span>Font Family</span>
            <input type="text" v-model="store.config.theme.fontFamily" placeholder="system-ui" @blur="saveSettings" />
          </label>

          <!-- Removed manual Save button -->
        </div>
      </div>
    </div>
  </aside>
</template>

<script setup>
import { ref } from 'vue'
import { User, Settings, PlusSquare } from 'lucide-vue-next'
import { useSlidesStore } from '../stores/slides'

const store = useSlidesStore()
const showToast = ref(false)
const toastMessage = ref('Under development')
let toastTimer = null
function showDevToast() {
  toastMessage.value = 'Under development'
  showToast.value = true
  clearTimeout(toastTimer)
  toastTimer = setTimeout(() => { showToast.value = false }, 1000)
}

const showSettings = ref(false)
function toggleSettings() { showSettings.value = !showSettings.value }

// Create PPT modal state
const showCreate = ref(false)
const createForm = ref({ name: '', title: '', description: '' })
function openCreateModal() { showCreate.value = true }
function closeCreate() { showCreate.value = false; createForm.value = { name: '', title: '', description: '' } }

// Auto-save debounce on blur within settings panel
let autoSaveTimer = null
function scheduleAutoSave() {
  clearTimeout(autoSaveTimer)
  autoSaveTimer = setTimeout(() => {
    saveSettings()
  }, 300)
}

async function createPpt() {
  const name = (createForm.value.name || '').trim()
  const title = (createForm.value.title || '').trim()
  const description = (createForm.value.description || '').trim()
  if (!name) return
  try {
    const res = await fetch('/api/presentations/create', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ group: name, title, description })
    })
    const json = await res.json()
    if (!json.ok) {
      console.error('Create presentation failed:', json.error)
      return
    }
    // Switch to new PPT and reload config
    store.setGroup(name)
    await store.loadConfig()
    // If title/description provided, patch config in memory then persist via store or save
    if (title) store.config.title = title
    if (description) store.config.description = description
    closeCreate()
  } catch (e) {
    console.error('Create presentation error:', e)
  }
}

async function saveSettings() {
  try {
    const res = await fetch('/api/config/save', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ group: store.currentGroup, config: store.config })
    })
    const json = await res.json()
    if (!json.ok) {
      console.error('Save settings failed:', json.error)
      return
    }
    // Show saved toast
    toastMessage.value = 'Saved'
    showToast.value = true
    clearTimeout(toastTimer)
    toastTimer = setTimeout(() => { showToast.value = false }, 800)
  } catch (e) {
    console.error('Save settings error:', e)
  }
}
</script>

<style scoped>
@reference "../tw.css";
.left-menu {
  @apply bg-gray-800 flex flex-col items-center justify-between;
  @apply mr-px px-px;
  width: calc(var(--spacing, 0.25rem) * 10);
}
.menu-inner { @apply flex flex-col items-center gap-2 px-2 py-3; }
.toolbar-btn { 
  @apply w-10 h-10 rounded-full flex items-center justify-center text-white transition-colors duration-150; 
  @apply bg-white/5 hover:bg-white/10 border border-white/20 backdrop-blur-sm;
}
.btn-user:hover { @apply ring-1 ring-white/40; }
.btn-settings .icon { @apply transition-transform; }
.btn-settings:hover .icon { transform: rotate(45deg); }
.btn-add .icon { @apply transition-transform; }
.btn-add:hover .icon { transform: scale(1.15); }

.settings-panel {
  @apply fixed inset-0 z-40 flex items-start justify-start;
}
.panel {
  @apply mt-16 ml-16 bg-gray-800 border border-gray-700 rounded-xl shadow-2xl w-[22rem] overflow-hidden;
}
.panel-header { @apply flex items-center justify-between p-4 border-b border-gray-700; }
.panel-body { @apply p-4 space-y-3; }
.form-row { @apply flex items-center justify-between gap-3; }
.form-row input[type="number"],
.form-row select,
.form-row input[type="text"] { @apply w-36 px-2 py-1 bg-gray-700 border border-gray-600 rounded-lg text-gray-100; }
.form-row input[type="color"] { @apply h-9 w-9 p-1 bg-gray-700 border border-gray-600 rounded-lg; }
.close-btn { @apply w-8 h-8 rounded-lg bg-white/5 hover:bg-white/10 border border-white/20; }
</style>