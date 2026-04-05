import { useAssistantPanel } from './useAssistantPanel'

describe('useAssistantPanel', () => {
  it('starts with panel closed and no unread', () => {
    const { isOpen, hasUnread } = useAssistantPanel()
    expect(isOpen.value).toBe(false)
    expect(hasUnread.value).toBe(false)
  })

  it('open() sets isOpen true and clears unread', () => {
    const { isOpen, hasUnread, open, markUnread } = useAssistantPanel()
    markUnread()
    open()
    expect(isOpen.value).toBe(true)
    expect(hasUnread.value).toBe(false)
  })

  it('close() sets isOpen false', () => {
    const { isOpen, open, close } = useAssistantPanel()
    open()
    close()
    expect(isOpen.value).toBe(false)
  })

  it('toggle() opens when closed', () => {
    const { isOpen, toggle } = useAssistantPanel()
    toggle()
    expect(isOpen.value).toBe(true)
  })

  it('toggle() closes when open', () => {
    const { isOpen, open, toggle } = useAssistantPanel()
    open()
    toggle()
    expect(isOpen.value).toBe(false)
  })

  it('markUnread() sets hasUnread when panel is closed', () => {
    const { hasUnread, markUnread } = useAssistantPanel()
    markUnread()
    expect(hasUnread.value).toBe(true)
  })

  it('markUnread() does NOT set hasUnread when panel is open', () => {
    const { hasUnread, open, markUnread } = useAssistantPanel()
    open()
    markUnread()
    expect(hasUnread.value).toBe(false)
  })
})
