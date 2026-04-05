import { mountSuspended } from '@nuxt/test-utils/runtime'
import PriceDisplay from './PriceDisplay.vue'

describe('PriceDisplay', () => {
  it('renders base price formatted to 2 decimals', async () => {
    const wrapper = await mountSuspended(PriceDisplay, { props: { basePrice: 29.9 } })
    expect(wrapper.text()).toContain('$29.90')
  })

  it('renders priceOverride when provided', async () => {
    const wrapper = await mountSuspended(PriceDisplay, { props: { basePrice: 50, priceOverride: 39.99 } })
    expect(wrapper.text()).toContain('$39.99')
    expect(wrapper.text()).not.toContain('$50.00')
  })

  it('renders basePrice when priceOverride is null', async () => {
    const wrapper = await mountSuspended(PriceDisplay, { props: { basePrice: 25, priceOverride: null } })
    expect(wrapper.text()).toContain('$25.00')
  })
})
