import { mountSuspended } from '@nuxt/test-utils/runtime'
import ConfirmDialog from './ConfirmDialog.vue'

const stubs = { Teleport: true }

describe('ConfirmDialog', () => {
  const baseProps = { open: true, title: 'Delete?', message: 'Are you sure?' }

  it('not rendered when open is false', async () => {
    const wrapper = await mountSuspended(ConfirmDialog, { props: { ...baseProps, open: false }, global: { stubs } })
    expect(wrapper.find('.fixed').exists()).toBe(false)
  })

  it('renders title, message, and buttons when open', async () => {
    const wrapper = await mountSuspended(ConfirmDialog, { props: baseProps, global: { stubs } })
    expect(wrapper.text()).toContain('Delete?')
    expect(wrapper.text()).toContain('Are you sure?')
    expect(wrapper.text()).toContain('Confirm')
    expect(wrapper.text()).toContain('Cancel')
  })

  it('emits confirm when confirm button clicked', async () => {
    const wrapper = await mountSuspended(ConfirmDialog, { props: baseProps, global: { stubs } })
    const confirmBtn = wrapper.findAll('button').find(b => b.text() === 'Confirm')!
    await confirmBtn.trigger('click')
    expect(wrapper.emitted('confirm')).toHaveLength(1)
  })

  it('emits cancel when cancel button clicked', async () => {
    const wrapper = await mountSuspended(ConfirmDialog, { props: baseProps, global: { stubs } })
    const cancelBtn = wrapper.findAll('button').find(b => b.text() === 'Cancel')!
    await cancelBtn.trigger('click')
    expect(wrapper.emitted('cancel')).toHaveLength(1)
  })

  it('uses custom confirmText', async () => {
    const wrapper = await mountSuspended(ConfirmDialog, { props: { ...baseProps, confirmText: 'Yes, delete' }, global: { stubs } })
    expect(wrapper.text()).toContain('Yes, delete')
  })
})
