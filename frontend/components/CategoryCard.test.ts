import { mountSuspended } from '@nuxt/test-utils/runtime'
import CategoryCard from './CategoryCard.vue'

describe('CategoryCard', () => {
  const category = { id: 'c1', name: 'Electronics', slug: 'electronics' }

  it('renders category name', async () => {
    const wrapper = await mountSuspended(CategoryCard, { props: { category } })
    expect(wrapper.text()).toContain('Electronics')
  })

  it('links to catalog page', async () => {
    const wrapper = await mountSuspended(CategoryCard, { props: { category } })
    expect(wrapper.find('a').attributes('href')).toBe('/catalog/electronics')
  })

  it('applies a gradient class deterministically', async () => {
    const w1 = await mountSuspended(CategoryCard, { props: { category } })
    const w2 = await mountSuspended(CategoryCard, { props: { category } })
    const gradient1 = w1.find('[class*="bg-gradient"]')
    const gradient2 = w2.find('[class*="bg-gradient"]')
    expect(gradient1.exists()).toBe(true)
    expect(gradient1.classes()).toEqual(gradient2.classes())
  })
})
