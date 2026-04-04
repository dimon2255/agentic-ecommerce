export function useAssistantPanel() {
  const isOpen = useState('assistant-panel-open', () => false)
  const hasUnread = useState('assistant-panel-unread', () => false)

  function open() {
    isOpen.value = true
    hasUnread.value = false
  }

  function close() {
    isOpen.value = false
  }

  function toggle() {
    if (isOpen.value) {
      close()
    } else {
      open()
    }
  }

  function markUnread() {
    if (!isOpen.value) {
      hasUnread.value = true
    }
  }

  return { isOpen, hasUnread, open, close, toggle, markUnread }
}
