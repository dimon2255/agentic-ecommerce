import { test, expect } from '@playwright/test'

test.describe('Cart persistence', () => {
  test('guest cart items persist after navigation', async ({ page }) => {
    await page.goto('/catalog/electronics')
    await page.locator('a[href^="/product/"]').first().click()

    await expect(page.locator('[role="radiogroup"]').first()).toBeVisible({ timeout: 10_000 })
    const radioGroups = page.locator('[role="radiogroup"]')
    const groupCount = await radioGroups.count()
    for (let i = 0; i < groupCount; i++) {
      const options = radioGroups.nth(i).locator('[role="radio"]')
      const optionCount = await options.count()
      for (let j = 0; j < optionCount; j++) {
        const option = options.nth(j)
        if ((await option.getAttribute('disabled')) === null) {
          await option.click()
          break
        }
      }
      await page.waitForTimeout(500)
    }

    await expect(page.getByRole('button', { name: 'Add to Cart' })).toBeEnabled({ timeout: 10_000 })
    await page.getByRole('button', { name: 'Add to Cart' }).click()
    await expect(page.getByText('Added to cart!').first()).toBeVisible({ timeout: 5_000 })

    await page.goto('/cart')
    await expect(page.getByText('$').first()).toBeVisible({ timeout: 10_000 })

    await page.goto('/')
    await page.goto('/cart')
    await expect(page.getByText('$').first()).toBeVisible({ timeout: 10_000 })
  })
})
