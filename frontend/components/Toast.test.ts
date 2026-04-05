import { mountSuspended } from '@nuxt/test-utils/runtime'
import Toast from './Toast.vue'

describe('Toast', () => {
  afterEach(() => {
    const { toasts } = useToast()
    toasts.value = []
  })

  it('renders toast messages from the queue', async () => {
    const { showToast } = useToast()
    showToast('Test message', 'success', 60000)
    await mountSuspended(Toast)
    expect(document.body.textContent).toContain('Test message')
  })

  it('each toast has role="alert"', async () => {
    const { showToast } = useToast()
    showToast('Alert toast', 'error', 60000)
    await mountSuspended(Toast)
    const alerts = document.body.querySelectorAll('[role="alert"]')
    expect(alerts.length).toBeGreaterThan(0)
  })

  it('renders no toasts when queue is empty', async () => {
    const { toasts } = useToast()
    toasts.value = []
    await mountSuspended(Toast)
    const alerts = document.body.querySelectorAll('[role="alert"]')
    expect(alerts.length).toBe(0)
  })
})
