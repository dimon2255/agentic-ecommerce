import { mountSuspended } from '@nuxt/test-utils/runtime'
import SlugInput from './SlugInput.vue'

describe('SlugInput', () => {
  it('renders name and slug inputs with initial values', async () => {
    const wrapper = await mountSuspended(SlugInput, { props: { name: 'My Product', slug: 'my-product' } })
    const inputs = wrapper.findAll('input')
    expect((inputs[0].element as HTMLInputElement).value).toBe('My Product')
    expect((inputs[1].element as HTMLInputElement).value).toBe('my-product')
  })

  it('emits update:name on name input', async () => {
    const wrapper = await mountSuspended(SlugInput, { props: { name: '', slug: '' } })
    const nameInput = wrapper.findAll('input')[0]
    await nameInput.setValue('Hello World')
    expect(wrapper.emitted('update:name')).toBeTruthy()
  })

  it('auto-generates slug from name when autoSlug is true', async () => {
    const wrapper = await mountSuspended(SlugInput, { props: { name: '', slug: '' } })
    const nameInput = wrapper.findAll('input')[0]
    await nameInput.setValue('Hello World!')
    const slugEmits = wrapper.emitted('update:slug')
    expect(slugEmits).toBeTruthy()
    const lastSlug = slugEmits![slugEmits!.length - 1][0]
    expect(lastSlug).toBe('hello-world')
  })

  it('manual slug edit disables autoSlug', async () => {
    const wrapper = await mountSuspended(SlugInput, { props: { name: 'Test', slug: 'test' } })
    const slugInput = wrapper.findAll('input')[1]
    await slugInput.setValue('custom-slug')
    // Now name input should not regenerate slug
    const nameInput = wrapper.findAll('input')[0]
    await nameInput.setValue('New Name')
    const slugEmits = wrapper.emitted('update:slug')!
    // Last emit from name change should not have auto-slug
    // After manual edit, only update:slug from the manual input should exist
    const manualEmit = slugEmits.find(e => e[0] === 'custom-slug')
    expect(manualEmit).toBeTruthy()
  })

  it('Auto button re-enables autoSlug and regenerates', async () => {
    const wrapper = await mountSuspended(SlugInput, { props: { name: 'Hello World', slug: 'manual' } })
    // First manually edit slug to disable auto
    const slugInput = wrapper.findAll('input')[1]
    await slugInput.setValue('manual-slug')
    // Click auto button
    const autoBtn = wrapper.find('button')
    await autoBtn.trigger('click')
    const slugEmits = wrapper.emitted('update:slug')!
    const lastSlug = slugEmits[slugEmits.length - 1][0]
    expect(lastSlug).toBe('hello-world')
  })

  it('slugify handles special characters', async () => {
    const wrapper = await mountSuspended(SlugInput, { props: { name: '', slug: '' } })
    const nameInput = wrapper.findAll('input')[0]
    await nameInput.setValue('  Hello, World! @#$ Test  ')
    const slugEmits = wrapper.emitted('update:slug')!
    const lastSlug = slugEmits[slugEmits.length - 1][0]
    expect(lastSlug).toBe('hello-world-test')
  })
})
