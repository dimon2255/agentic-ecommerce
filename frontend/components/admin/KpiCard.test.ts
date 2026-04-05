import { mountSuspended } from '@nuxt/test-utils/runtime'
import KpiCard from './KpiCard.vue'

describe('KpiCard', () => {
  it('renders label and value', async () => {
    const wrapper = await mountSuspended(KpiCard, { props: { label: 'Orders', value: 42 } })
    expect(wrapper.text()).toContain('Orders')
    expect(wrapper.text()).toContain('42')
  })

  it('shows prefix before value', async () => {
    const wrapper = await mountSuspended(KpiCard, { props: { label: 'Revenue', value: 1000, prefix: '$' } })
    expect(wrapper.text()).toContain('$')
    expect(wrapper.text()).toContain('1,000')
  })

  it('formats currency to 2 decimal places', async () => {
    const wrapper = await mountSuspended(KpiCard, { props: { label: 'Avg', value: 49.5, format: 'currency' } })
    expect(wrapper.text()).toContain('49.50')
  })

  it('shows loading skeleton when loading', async () => {
    const wrapper = await mountSuspended(KpiCard, { props: { label: 'Orders', value: 0, loading: true } })
    expect(wrapper.find('.animate-pulse').exists()).toBe(true)
  })

  it('does not show skeleton when not loading', async () => {
    const wrapper = await mountSuspended(KpiCard, { props: { label: 'Orders', value: 10 } })
    expect(wrapper.find('.animate-pulse').exists()).toBe(false)
  })
})
