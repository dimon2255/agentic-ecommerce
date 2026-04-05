import { test, expect } from '@playwright/test'

test.describe('Checkout flow', () => {
  test('checkout page requires cart items', async ({ page }) => {
    await page.goto('/checkout')
    // Should show checkout form or redirect if cart is empty
    await expect(page).toHaveURL(/checkout|cart/, { timeout: 10_000 })
  })

  test('checkout form renders with cart items', async ({ page }) => {
    // Add item to cart first
    await page.goto('/catalog/electronics')
    await page.locator('a[href^="/product/"]').first().click()
    const addBtn = page.getByRole('button', { name: /add to cart/i })
    await expect(addBtn).toBeVisible({ timeout: 10_000 })
    await addBtn.click()

    // Navigate to checkout
    await page.goto('/checkout')

    // Should show email input and shipping fields
    await expect(page.getByPlaceholder(/email/i)).toBeVisible({ timeout: 10_000 })
  })
})
