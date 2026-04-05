import { mockNuxtImport } from '@nuxt/test-utils/runtime'
import { createMockSupabaseClient } from './__mocks__/supabase'
import type { CartData } from './useCart'

const mockGet = vi.fn()
const mockPost = vi.fn()
const mockPatch = vi.fn()
const mockDel = vi.fn()

mockNuxtImport('useApi', () => {
  return () => ({ get: mockGet, post: mockPost, patch: mockPatch, del: mockDel })
})

let mockClient = createMockSupabaseClient({ accessToken: 'user-token' })

mockNuxtImport('useSupabaseClient', () => () => mockClient)

import { useCart } from './useCart'

function makeCartData(items: Array<Partial<CartData['items'][0]>> = []): CartData {
  return {
    id: 'cart-1',
    items: items.map((item, i) => ({
      id: `item-${i}`,
      sku_id: `sku-${i}`,
      quantity: 1,
      unit_price: 10.00,
      skus: {
        sku_code: `CODE-${i}`,
        price_override: null,
        products: { name: `Product ${i}`, slug: `product-${i}`, base_price: 10.00, images: [] },
      },
      ...item,
    })),
  }
}

describe('useCart', () => {
  beforeEach(() => {
    mockGet.mockReset()
    mockPost.mockReset()
    mockPatch.mockReset()
    mockDel.mockReset()
    mockClient = createMockSupabaseClient({ accessToken: 'user-token' })
    localStorage.clear()
  })

  describe('getHeaders', () => {
    it('includes Authorization when session exists', async () => {
      const cartData = makeCartData()
      mockGet.mockResolvedValue(cartData)
      const { refresh } = useCart()
      await refresh(true)
      expect(mockGet).toHaveBeenCalledWith('/cart', expect.objectContaining({
        Authorization: 'Bearer user-token',
      }))
    })

    it('includes X-Session-ID from localStorage', async () => {
      localStorage.setItem('session_id', 'existing-session')
      const cartData = makeCartData()
      mockGet.mockResolvedValue(cartData)
      const { refresh } = useCart()
      await refresh(true)
      expect(mockGet).toHaveBeenCalledWith('/cart', expect.objectContaining({
        'X-Session-ID': 'existing-session',
      }))
    })

    it('generates and persists new session ID if none exists', async () => {
      mockGet.mockResolvedValue(makeCartData())
      const { refresh } = useCart()
      await refresh(true)
      const sessionId = localStorage.getItem('session_id')
      expect(sessionId).toBeTruthy()
      expect(mockGet).toHaveBeenCalledWith('/cart', expect.objectContaining({
        'X-Session-ID': sessionId,
      }))
    })
  })

  describe('refresh', () => {
    it('fetches cart and sets state', async () => {
      const cartData = makeCartData([{ quantity: 2 }])
      mockGet.mockResolvedValue(cartData)
      const { cart, refresh } = useCart()
      await refresh(true)
      expect(cart.value).toEqual(cartData)
    })

    it('skips fetch if already initialized (lazy init)', async () => {
      mockGet.mockResolvedValue(makeCartData())
      const { refresh } = useCart()
      await refresh(true)
      mockGet.mockClear()
      await refresh()
      expect(mockGet).not.toHaveBeenCalled()
    })

    it('fetches when force=true even if initialized', async () => {
      mockGet.mockResolvedValue(makeCartData())
      const { refresh } = useCart()
      await refresh(true)
      mockGet.mockClear()
      mockGet.mockResolvedValue(makeCartData([{ quantity: 5 }]))
      await refresh(true)
      expect(mockGet).toHaveBeenCalled()
    })

    it('sets cart to null on error', async () => {
      mockGet.mockRejectedValue(new Error('Network'))
      const { cart, refresh } = useCart()
      await refresh(true)
      expect(cart.value).toBeNull()
    })
  })

  describe('CRUD operations', () => {
    it('addItem() posts to /cart/items and updates cart', async () => {
      const cartData = makeCartData([{ sku_id: 'sku-new' }])
      mockPost.mockResolvedValue(cartData)
      const { cart, addItem } = useCart()
      await addItem('sku-new', 2)
      expect(mockPost).toHaveBeenCalledWith('/cart/items', { sku_id: 'sku-new', quantity: 2 }, expect.any(Object))
      expect(cart.value).toEqual(cartData)
    })

    it('updateItem() patches /cart/items/{id} and updates cart', async () => {
      const cartData = makeCartData([{ id: 'item-1', quantity: 3 }])
      mockPatch.mockResolvedValue(cartData)
      const { cart, updateItem } = useCart()
      await updateItem('item-1', 3)
      expect(mockPatch).toHaveBeenCalledWith('/cart/items/item-1', { quantity: 3 }, expect.any(Object))
      expect(cart.value).toEqual(cartData)
    })

    it('removeItem() deletes then refreshes', async () => {
      mockDel.mockResolvedValue(undefined)
      mockGet.mockResolvedValue(makeCartData())
      const { removeItem } = useCart()
      await removeItem('item-1')
      expect(mockDel).toHaveBeenCalledWith('/cart/items/item-1', expect.any(Object))
    })
  })

  describe('computed', () => {
    it('itemCount returns sum of quantities', async () => {
      mockGet.mockResolvedValue(makeCartData([{ quantity: 2 }, { quantity: 3 }]))
      const { itemCount, refresh } = useCart()
      await refresh(true)
      expect(itemCount.value).toBe(5)
    })

    it('itemCount returns 0 when cart is null', () => {
      const { cart, itemCount } = useCart()
      cart.value = null
      expect(itemCount.value).toBe(0)
    })

    it('total returns sum of unit_price * quantity', async () => {
      mockGet.mockResolvedValue(makeCartData([
        { unit_price: 10.00, quantity: 2 },
        { unit_price: 25.50, quantity: 1 },
      ]))
      const { total, refresh } = useCart()
      await refresh(true)
      expect(total.value).toBe(45.50)
    })

    it('total returns 0 when cart items are empty', async () => {
      mockGet.mockResolvedValue(makeCartData([]))
      const { total, refresh } = useCart()
      await refresh(true)
      expect(total.value).toBe(0)
    })
  })
})
