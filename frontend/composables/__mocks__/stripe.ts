import { vi } from 'vitest'

export function createMockStripe() {
  const mockPaymentElement = { mount: vi.fn(), destroy: vi.fn() }
  const mockElements = {
    create: vi.fn().mockReturnValue(mockPaymentElement),
  }
  const mockStripe = {
    elements: vi.fn().mockReturnValue(mockElements),
    confirmPayment: vi.fn().mockResolvedValue({ error: null }),
  }
  return { mockStripe, mockElements, mockPaymentElement }
}
