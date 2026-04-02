type ToastVariant = 'success' | 'error' | 'warning' | 'info'

interface Toast {
  id: number
  message: string
  variant: ToastVariant
}

let nextId = 0

export function useToast() {
  const toasts = useState<Toast[]>('toasts', () => [])

  function showToast(message: string, variant: ToastVariant = 'info', duration = 4000) {
    const id = nextId++
    toasts.value.push({ id, message, variant })

    if (toasts.value.length > 3) {
      toasts.value.shift()
    }

    setTimeout(() => {
      toasts.value = toasts.value.filter(t => t.id !== id)
    }, duration)
  }

  return { toasts, showToast }
}
