import { mount } from '@vue/test-utils'
import { describe, it, expect, vi } from 'vitest'
import HomepageTopNav from '../../src/components/homepage/HomepageTopNav.vue'
import { homepageContent } from '../../src/utils/homepageContent.js'
import * as analytics from '../../src/utils/analytics.js'

vi.mock('../../src/utils/analytics.js', () => ({
  trackHomepageCta: vi.fn(() => ({ delivered: true }))
}))

describe('HomepageTopNav', () => {
  it('renders register and login CTA with accessibility labels', () => {
    const wrapper = mount(HomepageTopNav, {
      props: {
        brand: 'Online PPT',
        navigation: homepageContent.navigation
      }
    })

    const login = wrapper.get('[data-testid="nav-login"]')
    const register = wrapper.get('[data-testid="nav-register"]')

    expect(login.text()).toBe(homepageContent.navigation.login.label)
    expect(register.text()).toBe(homepageContent.navigation.register.label)
    expect(login.attributes('aria-label')).toBeDefined()
    expect(register.attributes('aria-label')).toBeDefined()
    wrapper.unmount()
  })

  it('sends analytics payload with placement meta', async () => {
    const wrapper = mount(HomepageTopNav, {
      props: {
        brand: 'Online PPT',
        navigation: homepageContent.navigation
      }
    })

    const login = wrapper.get('[data-testid="nav-login"]')
    await login.trigger('click')

    expect(analytics.trackHomepageCta).toHaveBeenCalledWith(
      expect.objectContaining({ placement: 'header-login', content: homepageContent.navigation.login.label })
    )
    wrapper.unmount()
  })
})
