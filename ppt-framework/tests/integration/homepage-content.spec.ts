import { mount } from '@vue/test-utils'
import { describe, it, expect, vi } from 'vitest'
import HomepageShell from '../../src/components/homepage/HomepageShell.vue'
import { homepageContent } from '../../src/utils/homepageContent.js'
import * as analytics from '../../src/utils/analytics.js'

vi.mock('../../src/utils/analytics.js', () => ({
  trackHomepageCta: vi.fn(() => ({ delivered: true }))
}))

describe('HomepageShell content sections', () => {
  it('renders feature cards and social proof entries from content map', () => {
    const wrapper = mount(HomepageShell)

    const featureItems = wrapper.findAll('[data-testid="feature-card"]')
    const testimonialItems = wrapper.findAll('[data-testid="testimonial-item"]')
    const footerHeadline = wrapper.get('[data-testid="footer-headline"]')

    expect(featureItems).toHaveLength(homepageContent.features.items.length)
    expect(testimonialItems.length).toBeGreaterThan(0)
    expect(footerHeadline.text()).toBe(homepageContent.footerCta.headline)

    wrapper.unmount()
  })

  it('emits analytics when footer CTA is clicked', async () => {
    const wrapper = mount(HomepageShell)
    const footerCta = wrapper.get('[data-testid="footer-primary-cta"]')
    await footerCta.trigger('click')

    expect(analytics.trackHomepageCta).toHaveBeenCalledWith(
      expect.objectContaining({ placement: 'footer-primary' })
    )
    wrapper.unmount()
  })
})
