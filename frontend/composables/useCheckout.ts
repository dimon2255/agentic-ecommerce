import { loadStripe, type Stripe, type StripeElements } from '@stripe/stripe-js'

interface PriceChange {
  sku_id: string
  sku_code: string
  old_price: number
  new_price: number
}

interface StartCheckoutResponse {
  order_id: string
  client_secret: string
}

export interface OrderResponse {
  id: string
  status: string
  email: string
  shipping_address: any
  subtotal: number
  total: number
  items: Array<{
    product_name: string
    sku_code: string
    quantity: number
    unit_price: number
  }>
  created_at: string
}

export function useCheckout() {
  const { post, get } = useApi()
  const config = useRuntimeConfig()
  const client = useSupabaseClient()

  const loading = ref(false)
  const error = ref('')
  const priceChanges = ref<PriceChange[]>([])

  let stripe: Stripe | null = null
  let elements: StripeElements | null = null

  async function getHeaders(): Promise<Record<string, string>> {
    const headers: Record<string, string> = {}
    try {
      const { data: { session } } = await client.auth.getSession()
      if (session?.access_token) {
        headers['Authorization'] = `Bearer ${session.access_token}`
      }
    } catch {}
    if (import.meta.client) {
      const sessionId = localStorage.getItem('session_id')
      if (sessionId) {
        headers['X-Session-ID'] = sessionId
      }
    }
    return headers
  }

  async function startCheckout(email: string, shippingAddress: Record<string, string>) {
    loading.value = true
    error.value = ''
    priceChanges.value = []

    try {
      const headers = await getHeaders()
      const data = await post<StartCheckoutResponse>('/checkout/start', {
        email,
        shipping_address: shippingAddress,
      }, headers)
      return data
    } catch (err: any) {
      if (err.statusCode === 409 && err.data?.price_changes) {
        priceChanges.value = err.data.price_changes
        return null
      }
      error.value = err.data?.error || 'Checkout failed'
      return null
    } finally {
      loading.value = false
    }
  }

  async function initStripe(clientSecret: string) {
    stripe = await loadStripe(config.public.stripeKey as string)
    if (!stripe) throw new Error('Failed to load Stripe')
    const style = getComputedStyle(document.documentElement)
    elements = stripe.elements({
      clientSecret,
      appearance: {
        theme: 'night',
        variables: {
          colorPrimary: style.getPropertyValue('--accent').trim(),
          colorBackground: style.getPropertyValue('--bg-elevated').trim(),
          colorText: style.getPropertyValue('--text-primary').trim(),
          colorDanger: style.getPropertyValue('--color-error').trim(),
          fontFamily: 'DM Sans, sans-serif',
          borderRadius: '8px',
        },
      },
    })
    const paymentElement = elements.create('payment')
    paymentElement.mount('#payment-element')
  }

  async function confirmPayment(orderId: string) {
    if (!stripe || !elements) throw new Error('Stripe not initialized')
    const { error: stripeError } = await stripe.confirmPayment({
      elements,
      confirmParams: {
        return_url: `${window.location.origin}/order/${orderId}`,
      },
    })
    if (stripeError) {
      return stripeError.message || 'Payment failed'
    }
    return null
  }

  async function getOrder(orderId: string): Promise<OrderResponse> {
    const headers = await getHeaders()
    return await get<OrderResponse>(`/orders/${orderId}`, headers)
  }

  return { loading, error, priceChanges, startCheckout, initStripe, confirmPayment, getOrder }
}
