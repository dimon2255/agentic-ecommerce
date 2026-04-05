import { mountSuspended } from '@nuxt/test-utils/runtime'
import ProductCard from './ProductCard.vue'

describe('ProductCard', () => {
  const product = {
    id: 'p1', name: 'Test Product', slug: 'test-product', base_price: 29.99,
    description: 'A great product', images: ['https://example.com/img.jpg'],
  }

  it('renders product name and price', async () => {
    const wrapper = await mountSuspended(ProductCard, { props: { product } })
    expect(wrapper.text()).toContain('Test Product')
    expect(wrapper.text()).toContain('$29.99')
  })

  it('renders product image when available', async () => {
    const wrapper = await mountSuspended(ProductCard, { props: { product } })
    expect(wrapper.find('img').exists()).toBe(true)
    expect(wrapper.find('img').attributes('src')).toBe('https://example.com/img.jpg')
  })

  it('shows fallback icon when no images', async () => {
    const noImg = { ...product, images: [] }
    const wrapper = await mountSuspended(ProductCard, { props: { product: noImg } })
    expect(wrapper.find('img').exists()).toBe(false)
    expect(wrapper.find('svg').exists()).toBe(true)
  })

  it('links to product page', async () => {
    const wrapper = await mountSuspended(ProductCard, { props: { product } })
    expect(wrapper.find('a').attributes('href')).toBe('/product/test-product')
  })
})
