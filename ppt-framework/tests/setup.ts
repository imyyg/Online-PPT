import { afterEach, vi } from 'vitest'

if (typeof window !== 'undefined') {
  if (!window.matchMedia) {
    window.matchMedia = () => ({
      matches: false,
      addEventListener: () => {},
      removeEventListener: () => {},
      addListener: () => {},
      removeListener: () => {},
      dispatchEvent: () => false,
      onchange: null,
      media: ''
    })
  }
  if (!window.scrollTo) {
    window.scrollTo = () => {}
  }
}

afterEach(() => {
  vi.clearAllMocks()
})
