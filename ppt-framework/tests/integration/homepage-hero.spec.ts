import { mount } from '@vue/test-utils'
import { describe, it, expect, vi } from 'vitest'
import HomepageHero from '../../src/components/homepage/HomepageHero.vue'
import { homepageContent } from '../../src/utils/homepageContent.js'
import * as analytics from '../../src/utils/analytics.js'

vi.mock('../../src/utils/analytics.js', () => ({
  trackHomepageCta: vi.fn(() => ({ delivered: true, event: { eventName: 'cta_click' } }))
}))

const heroContent = homepageContent.hero

describe('HomepageHero', () => {

  it('renders hero headline and value copy', () => {
    const wrapper = mount(HomepageHero, {
      props: {
        hero: heroContent
      }
    })

    expect(wrapper.get('[data-testid="hero-title"]').text()).toBe(heroContent.title)
    expect(wrapper.get('[data-testid="hero-subtitle"]').text()).toContain(heroContent.subtitle)
  expect(wrapper.get('[data-testid="hero-primary-cta"]').text()).toBe(heroContent.primaryCta.label)
    wrapper.unmount()
  })

  it('tracks analytics on primary CTA click', async () => {
    const wrapper = mount(HomepageHero, {
      props: {
        hero: heroContent
      }
    })

    const button = wrapper.get('[data-testid="hero-primary-cta"]')
    await button.trigger('click')

    expect(analytics.trackHomepageCta).toHaveBeenCalled()
    expect(analytics.trackHomepageCta).toHaveBeenLastCalledWith(
      expect.objectContaining({ placement: 'hero-primary', content: heroContent.primaryCta.label })
    )
    wrapper.unmount()
  })
})
