import { test, expect } from '@playwright/test'

test.describe('Catalog browsing', () => {
  test('homepage loads and shows categories', async ({ page }) => {
    await page.goto('/')
    await expect(page).toHaveTitle(/e-shop|agentic/i)
    // Should show at least one category card
    await expect(page.locator('a[href^="/catalog/"]').first()).toBeVisible()
  })

  test('catalog page shows products', async ({ page }) => {
    await page.goto('/catalog/electronics')
    // Wait for products to load
    await expect(page.locator('a[href^="/product/"]').first()).toBeVisible({ timeout: 10_000 })
  })

  test('product detail page shows product info', async ({ page }) => {
    // Navigate from catalog to product
    await page.goto('/catalog/electronics')
    const productLink = page.locator('a[href^="/product/"]').first()
    await expect(productLink).toBeVisible({ timeout: 10_000 })
    await productLink.click()

    // Product page should show name, price, and add to cart
    await expect(page.locator('text=$')).toBeVisible({ timeout: 10_000 })
  })
})
