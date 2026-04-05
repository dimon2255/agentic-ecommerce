import { mountSuspended } from '@nuxt/test-utils/runtime'
import CartItem from './CartItem.vue'

const makeItem = (overrides = {}) => ({
  id: 'item-1',
  sku_id: 'sku-1',
  quantity: 2,
  unit_price: 25.00,
  skus: {
    sku_code: 'WIDGET-RED-M',
    price_override: null,
    products: { name: 'Widget', slug: 'widget', base_price: 25, images: ['https://example.com/w.jpg'] },
  },
  ...overrides,
})

describe('CartItem', () => {
  it('renders product name, SKU code, and unit price', async () => {
    const wrapper = await mountSuspended(CartItem, { props: { item: makeItem(), updating: false } })
    expect(wrapper.text()).toContain('Widget')
    expect(wrapper.text()).toContain('WIDGET-RED-M')
    expect(wrapper.text()).toContain('$25.00')
  })

  it('renders line total (unit_price * quantity)', async () => {
    const wrapper = await mountSuspended(CartItem, { props: { item: makeItem({ quantity: 3, unit_price: 10 }), updating: false } })
    expect(wrapper.text()).toContain('$30.00')
  })

  it('renders product image when available', async () => {
    const wrapper = await mountSuspended(CartItem, { props: { item: makeItem(), updating: false } })
    expect(wrapper.find('img').exists()).toBe(true)
  })

  it('emits update with decreased quantity on minus click', async () => {
    const wrapper = await mountSuspended(CartItem, { props: { item: makeItem({ quantity: 3 }), updating: false } })
    const minusBtn = wrapper.find('[aria-label="Decrease quantity"]')
    await minusBtn.trigger('click')
    expect(wrapper.emitted('update')?.[0]).toEqual(['item-1', 2])
  })

  it('emits update with increased quantity on plus click', async () => {
    const wrapper = await mountSuspended(CartItem, { props: { item: makeItem(), updating: false } })
    const plusBtn = wrapper.find('[aria-label="Increase quantity"]')
    await plusBtn.trigger('click')
    expect(wrapper.emitted('update')?.[0]).toEqual(['item-1', 3])
  })

  it('emits remove on remove click', async () => {
    const wrapper = await mountSuspended(CartItem, { props: { item: makeItem(), updating: false } })
    const removeBtn = wrapper.find('[aria-label="Remove item"]')
    await removeBtn.trigger('click')
    expect(wrapper.emitted('remove')?.[0]).toEqual(['item-1'])
  })

  it('minus button disabled when quantity is 1', async () => {
    const wrapper = await mountSuspended(CartItem, { props: { item: makeItem({ quantity: 1 }), updating: false } })
    const minusBtn = wrapper.find('[aria-label="Decrease quantity"]')
    expect(minusBtn.attributes('disabled')).toBeDefined()
  })
})
