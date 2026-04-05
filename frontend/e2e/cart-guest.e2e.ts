import { test, expect } from '@playwright/test'

async function selectSkuAndAddToCart(page: import('@playwright/test').Page) {
  // Wait for SKU selector
  await expect(page.locator('[role="radiogroup"]').first()).toBeVisible({ timeout: 10_000 })

  // Click first non-disabled radio in each group by scrolling to it and clicking
  const radioGroups = page.locator('[role="radiogroup"]')
  const groupCount = await radioGroups.count()
  for (let i = 0; i < groupCount; i++) {
    const group = radioGroups.nth(i)
    const options = group.locator('[role="radio"]')
    const optionCount = await options.count()
    for (let j = 0; j < optionCount; j++) {
      const option = options.nth(j)
      const isDisabled = await option.getAttribute('disabled')
      if (isDisabled === null) {
        await option.click()
        break
      }
    }
    await page.waitForTimeout(500)
  }

  // Wait for "Add to Cart" to become enabled
  await expect(page.getByRole('button', { name: 'Add to Cart' })).toBeEnabled({ timeout: 10_000 })
  await page.getByRole('button', { name: 'Add to Cart' }).click()
  await expect(page.getByText('Added to cart!').first()).toBeVisible({ timeout: 5_000 })
}

test.describe('Guest cart', () => {
  test('add item to cart and view cart', async ({ page }) => {
    await page.goto('/catalog/electronics')
    await page.locator('a[href^="/product/"]').first().click()
    await selectSkuAndAddToCart(page)

    await page.goto('/cart')
    await expect(page.getByText('$').first()).toBeVisible({ timeout: 10_000 })
  })

  test('update quantity in cart', async ({ page }) => {
    await page.goto('/catalog/electronics')
    await page.locator('a[href^="/product/"]').first().click()
    await selectSkuAndAddToCart(page)

    await page.goto('/cart')
    const increaseBtn = page.getByLabel('Increase quantity').first()
    await expect(increaseBtn).toBeVisible({ timeout: 10_000 })
    await increaseBtn.click()
    await expect(page.locator('[aria-live="polite"]').first()).toContainText('2', { timeout: 5_000 })
  })
})
