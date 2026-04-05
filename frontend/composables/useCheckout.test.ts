import { mockNuxtImport } from '@nuxt/test-utils/runtime'
import { createMockSupabaseClient } from './__mocks__/supabase'
import { createMockStripe } from './__mocks__/stripe'

vi.mock('@stripe/stripe-js', () => ({
  loadStripe: vi.fn(),
}))

import { loadStripe } from '@stripe/stripe-js'

const mockGet = vi.fn()
const mockPost = vi.fn()

mockNuxtImport('useApi', () => {
  return () => ({ get: mockGet, post: mockPost, patch: vi.fn(), del: vi.fn() })
})

mockNuxtImport('useRuntimeConfig', () => {
  return () => ({
    app: { baseURL: '/' },
    public: { apiBase: 'http://test-api', stripeKey: 'pk_test_123' },
  })
})

let mockClient = createMockSupabaseClient({ accessToken: 'checkout-token' })
mockNuxtImport('useSupabaseClient', () => () => mockClient)

import { useCheckout } from './useCheckout'

describe('useCheckout', () => {
  beforeEach(() => {
    mockGet.mockReset()
    mockPost.mockReset()
    mockClient = createMockSupabaseClient({ accessToken: 'checkout-token' })
    localStorage.clear()
    localStorage.setItem('session_id', 'sess-123')
    vi.mocked(loadStripe).mockReset()
  })

  describe('startCheckout', () => {
    it('posts to /checkout/start with email and shipping', async () => {
      const resp = { order_id: 'ord-1', client_secret: 'cs_test' }
      mockPost.mockResolvedValue(resp)
      const { startCheckout } = useCheckout()
      const result = await startCheckout('a@b.com', { city: 'NYC' })
      expect(mockPost).toHaveBeenCalledWith(
        '/checkout/start',
        { email: 'a@b.com', shipping_address: { city: 'NYC' } },
        expect.objectContaining({ Authorization: 'Bearer checkout-token' }),
      )
      expect(result).toEqual(resp)
    })

    it('sets loading during request', async () => {
      let resolvePost: (v: any) => void
      mockPost.mockReturnValue(new Promise(r => { resolvePost = r }))
      const { loading, startCheckout } = useCheckout()
      const promise = startCheckout('a@b.com', {})
      expect(loading.value).toBe(true)
      resolvePost!({ order_id: 'o1', client_secret: 'cs' })
      await promise
      expect(loading.value).toBe(false)
    })

    it('handles 409 price change response', async () => {
      const changes = [{ sku_id: 's1', sku_code: 'SC1', old_price: 10, new_price: 15 }]
      mockPost.mockRejectedValue({ statusCode: 409, data: { price_changes: changes } })
      const { priceChanges, startCheckout } = useCheckout()
      const result = await startCheckout('a@b.com', {})
      expect(result).toBeNull()
      expect(priceChanges.value).toEqual(changes)
    })

    it('handles generic error', async () => {
      mockPost.mockRejectedValue({ statusCode: 500, data: { error: 'Server down' } })
      const { error, startCheckout } = useCheckout()
      const result = await startCheckout('a@b.com', {})
      expect(result).toBeNull()
      expect(error.value).toBe('Server down')
    })

    it('uses fallback error message', async () => {
      mockPost.mockRejectedValue({ statusCode: 500, data: {} })
      const { error, startCheckout } = useCheckout()
      await startCheckout('a@b.com', {})
      expect(error.value).toBe('Checkout failed')
    })
  })

  describe('initStripe', () => {
    it('calls loadStripe with key from runtime config', async () => {
      const { mockStripe, mockElements, mockPaymentElement } = createMockStripe()
      vi.mocked(loadStripe).mockResolvedValue(mockStripe as any)
      const { initStripe } = useCheckout()
      await initStripe('cs_test_secret')
      expect(loadStripe).toHaveBeenCalledWith('pk_test_123')
    })

    it('creates payment element and mounts to #payment-element', async () => {
      const { mockStripe, mockElements, mockPaymentElement } = createMockStripe()
      vi.mocked(loadStripe).mockResolvedValue(mockStripe as any)
      const div = document.createElement('div')
      div.id = 'payment-element'
      document.body.appendChild(div)

      const { initStripe } = useCheckout()
      await initStripe('cs_test_secret')

      expect(mockStripe.elements).toHaveBeenCalledWith(expect.objectContaining({
        clientSecret: 'cs_test_secret',
        appearance: expect.objectContaining({ theme: 'night' }),
      }))
      expect(mockElements.create).toHaveBeenCalledWith('payment')
      expect(mockPaymentElement.mount).toHaveBeenCalledWith('#payment-element')

      document.body.removeChild(div)
    })

    it('throws if loadStripe returns null', async () => {
      vi.mocked(loadStripe).mockResolvedValue(null)
      const { initStripe } = useCheckout()
      await expect(initStripe('cs_test')).rejects.toThrow('Failed to load Stripe')
    })
  })

  describe('confirmPayment', () => {
    it('throws if Stripe not initialized', async () => {
      const { confirmPayment } = useCheckout()
      await expect(confirmPayment('ord-1')).rejects.toThrow('Stripe not initialized')
    })

    it('calls stripe.confirmPayment with return URL', async () => {
      const { mockStripe, mockElements } = createMockStripe()
      vi.mocked(loadStripe).mockResolvedValue(mockStripe as any)
      const div = document.createElement('div')
      div.id = 'payment-element'
      document.body.appendChild(div)

      const { initStripe, confirmPayment } = useCheckout()
      await initStripe('cs_test')
      const result = await confirmPayment('ord-1')

      expect(mockStripe.confirmPayment).toHaveBeenCalledWith(expect.objectContaining({
        confirmParams: expect.objectContaining({
          return_url: expect.stringContaining('/order/ord-1'),
        }),
      }))
      expect(result).toBeNull()

      document.body.removeChild(div)
    })

    it('returns error message on Stripe error', async () => {
      const { mockStripe } = createMockStripe()
      mockStripe.confirmPayment.mockResolvedValue({ error: { message: 'Card declined' } })
      vi.mocked(loadStripe).mockResolvedValue(mockStripe as any)
      const div = document.createElement('div')
      div.id = 'payment-element'
      document.body.appendChild(div)

      const { initStripe, confirmPayment } = useCheckout()
      await initStripe('cs_test')
      const result = await confirmPayment('ord-1')
      expect(result).toBe('Card declined')

      document.body.removeChild(div)
    })
  })

  describe('getOrder', () => {
    it('fetches order by ID with auth headers', async () => {
      const order = { id: 'ord-1', status: 'paid', total: 50 }
      mockGet.mockResolvedValue(order)
      const { getOrder } = useCheckout()
      const result = await getOrder('ord-1')
      expect(mockGet).toHaveBeenCalledWith('/orders/ord-1', expect.objectContaining({
        Authorization: 'Bearer checkout-token',
      }))
      expect(result).toEqual(order)
    })
  })
})
