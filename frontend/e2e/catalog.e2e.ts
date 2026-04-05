import { test, expect } from '@playwright/test'

test.describe('Catalog browsing', () => {
  test('homepage loads and shows categories', async ({ page }) => {
    await page.goto('/')
    // Should show at least one category card
    await expect(page.locator('a[href^="/catalog/"]').first()).toBeVisible({ timeout: 10_000 })
  })

  test('catalog page shows products', async ({ page }) => {
    await page.goto('/catalog/electronics')
    await expect(page.locator('a[href^="/product/"]').first()).toBeVisible({ timeout: 10_000 })
  })

  test('product detail page shows product info', async ({ page }) => {
    await page.goto('/catalog/electronics')
    const productLink = page.locator('a[href^="/product/"]').first()
    await expect(productLink).toBeVisible({ timeout: 10_000 })
    await productLink.click()

    // Product page should show price and the add-to-cart/select-options button
    await expect(page.locator('text=$')).toBeVisible({ timeout: 10_000 })
    await expect(page.locator('button:has-text("Select options"), button:has-text("Add to Cart")').first()).toBeVisible()
  })
})
