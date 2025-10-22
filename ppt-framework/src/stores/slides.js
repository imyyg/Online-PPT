import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useSlidesStore = defineStore('slides', () => {
  // State
  const config = ref({
    title: 'My Presentation',
    author: '',
    description: '',
    theme: {
      primaryColor: '#3b82f6',
      fontFamily: 'system-ui',
      transition: 'slide'
    },
    settings: {
      autoPlay: false,
      autoPlayInterval: 5000,
      loop: false,
      showProgress: true,
      showThumbnails: true,
      enableKeyboardNav: true,
      enableTouchNav: true,
      autoStartOnHome: true,
      autoFullscreenOnHome: true
    },
    slides: []
  })
  
  const currentIndex = ref(0)
  const isPresenting = ref(false)
  const isSidebarCollapsed = ref(false)
  const loadingConfig = ref(false)
  const configReady = ref(false)

  // Current presentation group (folder under /presentations)
  const currentGroup = ref('example')
  const baseUrl = import.meta.env.BASE_URL || '/'
  const groupBasePath = computed(() => currentGroup.value ? `${baseUrl}presentations/${currentGroup.value}` : '')
  
  // Getters
  const currentSlide = computed(() => {
    const visibleSlides = config.value.slides.filter(s => s.visible !== false)
    return visibleSlides[currentIndex.value] || null
  })
  
  const visibleSlides = computed(() => {
    // 未显式设置 visible 的幻灯片默认可见
    return config.value.slides.filter(s => s.visible !== false)
  })
  
  const totalSlides = computed(() => {
    return visibleSlides.value.length
  })
  
  const progress = computed(() => {
    if (totalSlides.value === 0) return 0
    return ((currentIndex.value + 1) / totalSlides.value) * 100
  })
  
  const canGoNext = computed(() => {
    return currentIndex.value < totalSlides.value - 1
  })
  
  const canGoPrev = computed(() => {
    return currentIndex.value > 0
  })
  
  // Actions
  async function loadConfig() {
    try {
      loadingConfig.value = true
      configReady.value = false
      const url = `${groupBasePath.value}/slides.config.json`
      const response = await fetch(url)
      if (response.ok) {
        const data = await response.json()
        config.value = data
        configReady.value = true
        // Clamp current index to valid range
        const total = totalSlides.value
        if (currentIndex.value >= total) {
          currentIndex.value = Math.max(0, total - 1)
        }
      } else {
        // No config found for this group yet
        configReady.value = false
      }
    } catch (error) {
      console.error('Failed to load slides config:', error)
      configReady.value = false
    } finally {
      loadingConfig.value = false
    }
  }
  
  function setGroup(group) {
    currentGroup.value = group
    currentIndex.value = 0
  }
  
  function goToSlide(index) {
    if (index >= 0 && index < totalSlides.value) {
      currentIndex.value = index
    }
  }
  
  function nextSlide() {
    if (canGoNext.value) {
      currentIndex.value++
    } else if (config.value.settings.loop) {
      currentIndex.value = 0
    }
  }
  
  function prevSlide() {
    if (canGoPrev.value) {
      currentIndex.value--
    } else if (config.value.settings.loop) {
      currentIndex.value = totalSlides.value - 1
    }
  }
  
  function togglePresentation() {
    isPresenting.value = !isPresenting.value
  }
  
  function toggleSidebar() {
    isSidebarCollapsed.value = !isSidebarCollapsed.value
  }
  
  function addSlide(slide) {
    const newSlide = {
      id: `slide-${Date.now()}`,
      title: slide.title,
      file: slide.file,
      visible: true,
      notes: slide.notes || '',
      duration: null
    }
    config.value.slides.push(newSlide)
  }
  
  function removeSlide(id) {
    config.value.slides = config.value.slides.filter(s => s.id !== id)
  }
  
  function duplicateSlide(id) {
    const slide = config.value.slides.find(s => s.id === id)
    if (slide) {
      const base = slide.file.replace(/\.html$/, '')
      const newFile = `${base}-copy-${Date.now()}.html`
      const newSlide = {
        id: `slide-${Date.now()}`,
        title: `${slide.title} (Copy)`,
        file: newFile,
        visible: true,
        notes: slide.notes || '',
        duration: null
      }
      config.value.slides.push(newSlide)
      return newSlide
    }
    return null
  }
  
  function reorderSlides(fromIndex, toIndex) {
    const slides = config.value.slides
    const [removed] = slides.splice(fromIndex, 1)
    slides.splice(toIndex, 0, removed)
  }
  
  function updateSlide(id, updates) {
    const slide = config.value.slides.find(s => s.id === id)
    if (slide) {
      Object.assign(slide, updates)
    }
  }
  
  return {
    // State
    config,
    currentIndex,
    isPresenting,
    isSidebarCollapsed,
    loadingConfig,
    configReady,
    currentGroup,
    groupBasePath,
    
    // Getters
    currentSlide,
    visibleSlides,
    totalSlides,
    progress,
    canGoNext,
    canGoPrev,
    
    // Actions
    loadConfig,
    setGroup,
    goToSlide,
    nextSlide,
    prevSlide,
    togglePresentation,
    toggleSidebar,
    addSlide,
    removeSlide,
    duplicateSlide,
    reorderSlides,
    updateSlide
  }
})


