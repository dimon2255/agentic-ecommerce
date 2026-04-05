import { mountSuspended } from '@nuxt/test-utils/runtime'
import StatusBadge from './StatusBadge.vue'

describe('StatusBadge', () => {
  it('renders status text', async () => {
    const wrapper = await mountSuspended(StatusBadge, { props: { status: 'active' } })
    expect(wrapper.text()).toBe('active')
  })

  it('renders custom label when provided', async () => {
    const wrapper = await mountSuspended(StatusBadge, { props: { status: 'active', label: 'Live' } })
    expect(wrapper.text()).toBe('Live')
  })

  it('applies correct class for known status', async () => {
    const wrapper = await mountSuspended(StatusBadge, { props: { status: 'paid' } })
    expect(wrapper.find('span').classes().join(' ')).toContain('text-[var(--color-success)]')
  })

  it('applies fallback class for unknown status', async () => {
    const wrapper = await mountSuspended(StatusBadge, { props: { status: 'unknown-status' } })
    expect(wrapper.find('span').classes().join(' ')).toContain('text-muted')
  })
})
