import { test, expect } from '@playwright/test'

test.describe('Cart merge on login', () => {
  test('guest cart items persist after navigation', async ({ page }) => {
    // Add item as guest
    await page.goto('/catalog/electronics')
    await page.locator('a[href^="/product/"]').first().click()
    const addBtn = page.getByRole('button', { name: /add to cart/i })
    await expect(addBtn).toBeVisible({ timeout: 10_000 })
    await addBtn.click()

    // Verify cart has items
    await page.goto('/cart')
    await expect(page.locator('text=$')).toBeVisible({ timeout: 10_000 })

    // Navigate away and back — cart should still have items
    await page.goto('/')
    await page.goto('/cart')
    await expect(page.locator('text=$')).toBeVisible({ timeout: 10_000 })
  })
})
