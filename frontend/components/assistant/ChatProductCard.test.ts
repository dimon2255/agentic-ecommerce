import { mountSuspended } from '@nuxt/test-utils/runtime'
import ChatProductCard from './ChatProductCard.vue'

describe('ChatProductCard', () => {
  const product = { id: 'p1', name: 'Widget', slug: 'widget', base_price: 19.99, images: ['https://example.com/w.jpg'] }

  it('renders product name and price', async () => {
    const wrapper = await mountSuspended(ChatProductCard, { props: { product } })
    expect(wrapper.text()).toContain('Widget')
    expect(wrapper.text()).toContain('$19.99')
  })

  it('renders image when available', async () => {
    const wrapper = await mountSuspended(ChatProductCard, { props: { product } })
    expect(wrapper.find('img').exists()).toBe(true)
  })

  it('shows fallback when no images', async () => {
    const noImg = { ...product, images: [] }
    const wrapper = await mountSuspended(ChatProductCard, { props: { product: noImg } })
    expect(wrapper.find('img').exists()).toBe(false)
    expect(wrapper.find('svg').exists()).toBe(true)
  })
})
