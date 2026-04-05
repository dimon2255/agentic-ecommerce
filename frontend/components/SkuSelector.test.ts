import { mountSuspended } from '@nuxt/test-utils/runtime'
import SkuSelector from './SkuSelector.vue'

const attributes = [
  { id: 'a1', name: 'Color', options: ['Red', 'Blue'] },
  { id: 'a2', name: 'Size', options: ['S', 'M'] },
]

const skus = [
  { id: 's1', sku_code: 'RED-S', price_override: null, attribute_values: [{ category_attribute_id: 'a1', value: 'Red' }, { category_attribute_id: 'a2', value: 'S' }] },
  { id: 's2', sku_code: 'RED-M', price_override: 15, attribute_values: [{ category_attribute_id: 'a1', value: 'Red' }, { category_attribute_id: 'a2', value: 'M' }] },
  { id: 's3', sku_code: 'BLUE-S', price_override: null, attribute_values: [{ category_attribute_id: 'a1', value: 'Blue' }, { category_attribute_id: 'a2', value: 'S' }] },
  // Blue+M intentionally missing to test availability
]

describe('SkuSelector', () => {
  it('renders attribute groups with labels', async () => {
    const wrapper = await mountSuspended(SkuSelector, { props: { skus, attributes } })
    expect(wrapper.text()).toContain('Color')
    expect(wrapper.text()).toContain('Size')
  })

  it('renders option buttons', async () => {
    const wrapper = await mountSuspended(SkuSelector, { props: { skus, attributes } })
    const buttons = wrapper.findAll('button')
    expect(buttons.length).toBe(4) // Red, Blue, S, M
  })

  it('clicking option emits select', async () => {
    const wrapper = await mountSuspended(SkuSelector, { props: { skus, attributes } })
    const buttons = wrapper.findAll('button')
    const redBtn = buttons.find(b => b.text() === 'Red')!
    await redBtn.trigger('click')
    expect(wrapper.emitted('select')).toBeTruthy()
  })

  it('emits null when selection is incomplete', async () => {
    const wrapper = await mountSuspended(SkuSelector, { props: { skus, attributes } })
    const buttons = wrapper.findAll('button')
    const redBtn = buttons.find(b => b.text() === 'Red')!
    await redBtn.trigger('click')
    // Only one attribute selected, should emit null
    expect(wrapper.emitted('select')![0][0]).toBeNull()
  })

  it('emits SKU when all attributes selected', async () => {
    const wrapper = await mountSuspended(SkuSelector, { props: { skus, attributes } })
    const buttons = wrapper.findAll('button')
    await buttons.find(b => b.text() === 'Red')!.trigger('click')
    await buttons.find(b => b.text() === 'S')!.trigger('click')
    const emits = wrapper.emitted('select')!
    const lastEmit = emits[emits.length - 1][0]
    expect(lastEmit).toMatchObject({ sku_code: 'RED-S' })
  })

  it('shows SKU code when all attributes selected', async () => {
    const wrapper = await mountSuspended(SkuSelector, { props: { skus, attributes } })
    const buttons = wrapper.findAll('button')
    await buttons.find(b => b.text() === 'Red')!.trigger('click')
    await buttons.find(b => b.text() === 'M')!.trigger('click')
    expect(wrapper.text()).toContain('RED-M')
  })

  it('unavailable options are disabled', async () => {
    const wrapper = await mountSuspended(SkuSelector, { props: { skus, attributes } })
    // Select Blue first
    const buttons = wrapper.findAll('button')
    await buttons.find(b => b.text() === 'Blue')!.trigger('click')
    // M should be disabled because Blue+M doesn't exist
    const mBtn = wrapper.findAll('button').find(b => b.text() === 'M')!
    expect(mBtn.attributes('disabled')).toBeDefined()
  })
})
