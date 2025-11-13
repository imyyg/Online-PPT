const DEFAULT_EVENT = 'cta_click'

function sanitizePayload(payload) {
  if (!payload || typeof payload !== 'object') {
    return { eventName: DEFAULT_EVENT }
  }
  const copied = { ...payload }
  if (!copied.eventName) {
    copied.eventName = DEFAULT_EVENT
  }
  if (typeof copied.timestamp !== 'string') {
    copied.timestamp = new Date().toISOString()
  }
  return copied
}

function emitToConsole(event) {
  if (typeof import.meta !== 'undefined' && import.meta.env && import.meta.env.DEV) {
    // eslint-disable-next-line no-console
    console.debug('[homepage-analytics]', event)
  }
}

function emitToDataLayer(event) {
  if (typeof window === 'undefined') {
    return false
  }
  if (Array.isArray(window.dataLayer)) {
    window.dataLayer.push({ event: event.eventName, ...event })
    return true
  }
  if (window.analytics && typeof window.analytics.track === 'function') {
    window.analytics.track(event.eventName, event)
    return true
  }
  return false
}

export function trackHomepageCta(payload) {
  const event = sanitizePayload(payload)
  const delivered = emitToDataLayer(event)
  if (!delivered) {
    emitToConsole(event)
  }
  return { delivered, event }
}
