<template>
  <div class="slide-manager">
    <div class="manager-header">
      <h2 class="modal-title">Manage Slides</h2>
      <div class="header-actions">
        <button @click="showAddSlide = true" class="btn btn-primary">
          Add New Slide
        </button>
        <button @click="showCreateGroup = true" class="btn">
          Create New PPT (Group)
        </button>
      </div>
    </div>

    <div class="slides-grid">
      <SlidePreview
        v-for="(slide, index) in store.visibleSlides"
        :key="slide.id"
        :slide="slide"
        :index="index"
        :is-active="store.currentIndex === index"
        @select="store.goToSlide(index)"
        @edit="editSlide(slide)"
        @duplicate="duplicateSlide(slide.id)"
        @delete="confirmDelete(slide)"
        @reorder="handleReorder"
      />
    </div>

    <!-- Add/Edit Slide Modal -->
    <div v-if="showAddSlide" class="modal-overlay" @click.self="closeModal">
      <div class="modal-content">
        <h3 class="modal-title">{{ editingSlide ? 'Edit Slide' : 'Add New Slide' }}</h3>
        <form @submit.prevent="saveSlide" class="modal-form">
          <div class="form-group">
            <label for="slideTitle">Slide Title</label>
            <input
              id="slideTitle"
              v-model="slideForm.title"
              type="text"
              class="form-input"
              placeholder="Enter slide title"
              required
            />
          </div>
          
          <div class="form-group">
            <label for="slideFile">HTML File Name</label>
            <input
              id="slideFile"
              v-model="slideForm.file"
              type="text"
              class="form-input"
              placeholder="slide-name.html"
              pattern="[a-zA-Z0-9-_]+\.html"
              required
              :disabled="!!editingSlide"
            />
            <p class="form-hint">File will be created in {{ store.groupBasePath }}/slides/ directory</p>
          </div>
          
          <div class="form-group">
            <label for="slideTemplate">Template</label>
            <select
              id="slideTemplate"
              v-model="slideForm.template"
              class="form-input"
              :disabled="!!editingSlide"
            >
              <option value="blank">Blank Slide</option>
              <option value="title">Title Slide</option>
              <option value="content">Content Slide</option>
              <option value="image">Image Slide</option>
              <option value="list">List Slide</option>
            </select>
          </div>
          
          <div class="form-group">
            <label for="slideNotes">Speaker Notes</label>
            <textarea
              id="slideNotes"
              v-model="slideForm.notes"
              class="form-input"
              rows="3"
              placeholder="Add speaker notes..."
            />
          </div>

          <div class="modal-actions">
            <button type="button" class="btn" @click="closeModal">Cancel</button>
            <button type="submit" class="btn btn-primary">Save</button>
          </div>
        </form>
      </div>
    </div>

    <!-- Create Group Modal -->
    <div v-if="showCreateGroup" class="modal-overlay" @click.self="closeCreateGroup">
      <div class="modal-content modal-sm">
        <h3 class="modal-title">Create New PPT Group</h3>
        <form @submit.prevent="saveGroup" class="modal-form">
          <div class="form-group">
            <label for="groupName">Group Name</label>
            <input
              id="groupName"
              v-model="groupForm.name"
              type="text"
              class="form-input"
              placeholder="my-ppt"
              pattern="[a-zA-Z0-9-_]+"
              required
            />
            <p class="form-hint">Will create folder at /presentations/{{ groupForm.name }}/</p>
          </div>
          <div class="modal-actions">
            <button type="button" class="btn" @click="closeCreateGroup">Cancel</button>
            <button type="submit" class="btn btn-primary">Create</button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, watch } from 'vue'
import SlidePreview from './SlidePreview.vue'
import { useSlidesStore } from '../stores/slides'

const props = defineProps({
  autoOpenCreate: {
    type: Boolean,
    default: false
  },
  autoOpenGroup: {
    type: Boolean,
    default: false
  }
})

const store = useSlidesStore()

const showAddSlide = ref(false)
const showCreateGroup = ref(false)
const editingSlide = ref(null)
const slideToDelete = ref(null)

const slideForm = reactive({
  title: '',
  file: '',
  template: 'blank',
  notes: ''
})

const groupForm = reactive({ name: '' })

function resetCreateForm() {
  slideForm.title = ''
  slideForm.file = ''
  slideForm.template = 'blank'
  slideForm.notes = ''
}

onMounted(() => {
  if (props.autoOpenCreate) {
    resetCreateForm()
    showAddSlide.value = true
  }
  if (props.autoOpenGroup) {
    showCreateGroup.value = true
  }
})

watch(() => props.autoOpenCreate, (val) => {
  if (val) {
    resetCreateForm()
    showAddSlide.value = true
  }
})

watch(() => props.autoOpenGroup, (val) => {
  if (val) {
    showCreateGroup.value = true
  }
})

function editSlide(slide) {
  editingSlide.value = slide
  slideForm.title = slide.title
  slideForm.file = slide.file
  slideForm.notes = slide.notes || ''
  slideForm.template = 'blank'
}

async function saveSlide() {
  if (editingSlide.value) {
    // Update existing slide
    store.updateSlide(editingSlide.value.id, {
      title: slideForm.title,
      notes: slideForm.notes
    })
  } else {
    // Create new slide in current group
    store.addSlide({
      title: slideForm.title,
      file: slideForm.file,
      notes: slideForm.notes
    })
    
    // Create the HTML file with the selected template under group
    await createSlideFile(slideForm.file, slideForm.template, slideForm.title, store.currentGroup)
  }
  
  closeModal()
}

async function createSlideFile(filename, template, title, group) {
  try {
    const res = await fetch('/api/slides/create', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ file: filename, template, title, group })
    })
    const json = await res.json()
    if (!json.ok) {
      console.error('Create slide failed:', json.error)
    } else {
      // Reload config to reflect new slide
      await store.loadConfig()
    }
  } catch (e) {
    console.error('Create slide error:', e)
  }
}

function duplicateSlide(slideId) {
  const duplicated = store.duplicateSlide(slideId)
  if (duplicated) {
    fetch('/api/slides/duplicate', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ sourceFile: duplicated.file.replace(/-copy-\d+\.html$/, '.html'), sourceTitle: duplicated.title.replace(/ \(Copy\)$/,''), group: store.currentGroup })
    })
      .then(r => r.json())
      .then(async (json) => {
        if (!json.ok) {
          console.error('Duplicate slide failed:', json.error)
        } else {
          await store.loadConfig()
        }
      }).catch(e => console.error('Duplicate slide error:', e))
  }
}

function confirmDelete(slide) {
  slideToDelete.value = slide
}

function deleteSlide() {
  if (slideToDelete.value) {
    const s = slideToDelete.value
    // Call backend to physically delete under group
    fetch('/api/slides/delete', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ id: s.id, file: s.file, group: store.currentGroup })
    })
      .then(r => r.json())
      .then(async (json) => {
        if (!json.ok) {
          console.error('Delete slide failed:', json.error)
        } else {
          // Reload config and clamp index if needed
          await store.loadConfig()
          if (store.currentIndex >= store.totalSlides) {
            store.goToSlide(Math.max(0, store.totalSlides - 1))
          }
        }
      })
      .catch(e => console.error('Delete slide error:', e))
    slideToDelete.value = null
  }
}

function handleReorder({ from, to }) {
  store.reorderSlides(from, to)
}

function closeModal() {
  showAddSlide.value = false
  editingSlide.value = null
}

function closeCreateGroup() {
  showCreateGroup.value = false
  groupForm.name = ''
}

async function saveGroup() {
  const name = groupForm.name.trim()
  if (!name) return
  try {
    const res = await fetch('/api/presentations/create', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ group: name })
    })
    const json = await res.json()
    if (!json.ok) {
      console.error('Create group failed:', json.error)
      return
    }
    // Switch to new group and load its config
    store.setGroup(name)
    await store.loadConfig()
    closeCreateGroup()
  } catch (e) {
    console.error('Create group error:', e)
  }
}
</script>

<style scoped>
@reference "../style.css";
.slide-manager {
  @apply p-6;
}

.manager-header {
  @apply flex items-center justify-between mb-6;
}

.header-actions {
  @apply flex gap-2;
}

.slides-grid {
  @apply grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4;
}

.modal-overlay {
  @apply fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center p-4 z-50;
}

.modal-content {
  @apply bg-gray-800 rounded-lg p-6 max-w-md w-full max-h-[90vh] overflow-y-auto;
}

.modal-sm {
  @apply max-w-sm;
}

.modal-title {
  @apply text-xl font-semibold mb-4;
}

.modal-form {
  @apply space-y-4;
}

.form-group {
  @apply space-y-2;
}

.form-group label {
  @apply block text-sm font-medium text-gray-300;
}

.form-input {
  @apply w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg;
  @apply text-gray-100 placeholder-gray-400;
  @apply focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent;
  @apply disabled:opacity-50 disabled:cursor-not-allowed;
}

.form-hint {
  @apply text-xs text-gray-400;
}

.modal-actions {
  @apply flex gap-2 justify-end mt-6;
}
</style>
