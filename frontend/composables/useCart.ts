interface CartItemSKU {
  sku_code: string
  price_override: number | null
  products: {
    name: string
    slug: string
    base_price: number
    images: string[]
  }
}

export interface CartItem {
  id: string
  sku_id: string
  quantity: number
  unit_price: number
  skus: CartItemSKU
}

export interface CartData {
  id: string
  items: CartItem[]
}

export function useCart() {
  const { get, post, patch, del } = useApi()
  const client = useSupabaseClient()

  const cart = useState<CartData | null>('cart', () => null)
  const loading = useState('cart-loading', () => false)
  const initialized = useState('cart-initialized', () => false)

  async function getHeaders(): Promise<Record<string, string>> {
    const headers: Record<string, string> = {}

    try {
      const { data: { session } } = await client.auth.getSession()
      if (session?.access_token) {
        headers['Authorization'] = `Bearer ${session.access_token}`
      }
    } catch {
      // No auth available
    }

    if (import.meta.client) {
      let sessionId = localStorage.getItem('session_id')
      if (!sessionId) {
        sessionId = crypto.randomUUID()
        localStorage.setItem('session_id', sessionId)
      }
      headers['X-Session-ID'] = sessionId
    }

    return headers
  }

  async function refresh(force = false) {
    if (!import.meta.client) return
    if (initialized.value && !force) return
    loading.value = true
    try {
      const headers = await getHeaders()
      cart.value = await get<CartData>('/cart', headers)
      initialized.value = true
    } catch {
      cart.value = null
    } finally {
      loading.value = false
    }
  }

  async function addItem(skuId: string, quantity: number = 1) {
    const headers = await getHeaders()
    cart.value = await post<CartData>('/cart/items', { sku_id: skuId, quantity }, headers)
  }

  async function updateItem(itemId: string, quantity: number) {
    const headers = await getHeaders()
    cart.value = await patch<CartData>(`/cart/items/${itemId}`, { quantity }, headers)
  }

  async function removeItem(itemId: string) {
    const headers = await getHeaders()
    await del(`/cart/items/${itemId}`, headers)
    await refresh(true)
  }

  const itemCount = computed(() => {
    if (!cart.value?.items) return 0
    return cart.value.items.reduce((sum, item) => sum + item.quantity, 0)
  })

  const total = computed(() => {
    if (!cart.value?.items) return 0
    return cart.value.items.reduce((sum, item) => sum + item.unit_price * item.quantity, 0)
  })

  return { cart, loading, itemCount, total, refresh, addItem, updateItem, removeItem }
}
