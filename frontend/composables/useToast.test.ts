import { useToast } from './useToast'

describe('useToast', () => {
  beforeEach(() => {
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('adds a toast to the queue', () => {
    const { toasts, showToast } = useToast()
    showToast('Hello', 'success')
    expect(toasts.value).toHaveLength(1)
    expect(toasts.value[0].message).toBe('Hello')
    expect(toasts.value[0].variant).toBe('success')
  })

  it('defaults to info variant', () => {
    const { toasts, showToast } = useToast()
    showToast('Test')
    expect(toasts.value[toasts.value.length - 1].variant).toBe('info')
  })

  it('assigns unique ids to each toast', () => {
    const { toasts, showToast } = useToast()
    showToast('First')
    showToast('Second')
    const ids = toasts.value.map(t => t.id)
    expect(new Set(ids).size).toBe(ids.length)
  })

  it('auto-dismisses after specified duration', () => {
    const { toasts, showToast } = useToast()
    showToast('Bye', 'info', 2000)
    expect(toasts.value.some(t => t.message === 'Bye')).toBe(true)

    vi.advanceTimersByTime(2000)
    expect(toasts.value.some(t => t.message === 'Bye')).toBe(false)
  })

  it('auto-dismisses after default 4000ms', () => {
    const { toasts, showToast } = useToast()
    showToast('Default')
    expect(toasts.value.some(t => t.message === 'Default')).toBe(true)

    vi.advanceTimersByTime(4000)
    expect(toasts.value.some(t => t.message === 'Default')).toBe(false)
  })

  it('enforces max 3 toasts (FIFO)', () => {
    const { toasts, showToast } = useToast()
    showToast('One', 'info', 60000)
    showToast('Two', 'info', 60000)
    showToast('Three', 'info', 60000)
    showToast('Four', 'info', 60000)

    expect(toasts.value).toHaveLength(3)
    expect(toasts.value[0].message).toBe('Two')
    expect(toasts.value[2].message).toBe('Four')
  })

  it('multiple toasts dismiss independently', () => {
    const { toasts, showToast } = useToast()
    showToast('Fast', 'info', 1000)
    showToast('Slow', 'info', 5000)

    vi.advanceTimersByTime(1000)
    expect(toasts.value.some(t => t.message === 'Fast')).toBe(false)
    expect(toasts.value.some(t => t.message === 'Slow')).toBe(true)

    vi.advanceTimersByTime(4000)
    expect(toasts.value.some(t => t.message === 'Slow')).toBe(false)
  })
})
