import { test, expect } from '@playwright/test'

test.describe('Guest cart', () => {
  test('add item to cart and view cart', async ({ page }) => {
    // Browse to a product
    await page.goto('/catalog/electronics')
    await page.locator('a[href^="/product/"]').first().click()

    // Add to cart
    const addBtn = page.getByRole('button', { name: /add to cart/i })
    await expect(addBtn).toBeVisible({ timeout: 10_000 })
    await addBtn.click()

    // Navigate to cart
    await page.goto('/cart')
    // Should show at least one cart item
    await expect(page.locator('text=$')).toBeVisible({ timeout: 10_000 })
  })

  test('update quantity in cart', async ({ page }) => {
    // First add an item
    await page.goto('/catalog/electronics')
    await page.locator('a[href^="/product/"]').first().click()
    const addBtn = page.getByRole('button', { name: /add to cart/i })
    await expect(addBtn).toBeVisible({ timeout: 10_000 })
    await addBtn.click()

    // Go to cart and increase quantity
    await page.goto('/cart')
    const increaseBtn = page.getByRole('button', { name: /increase/i }).first()
    await expect(increaseBtn).toBeVisible({ timeout: 10_000 })
    await increaseBtn.click()

    // Quantity should update
    await expect(page.locator('[aria-live="polite"]').first()).toContainText('2', { timeout: 5_000 })
  })
})
