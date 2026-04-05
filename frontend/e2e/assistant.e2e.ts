import { test, expect } from '@playwright/test'

test.describe('AI Assistant', () => {
  test('assistant FAB is visible on page', async ({ page }) => {
    await page.goto('/')
    const fab = page.getByLabel('Open shopping assistant')
    await expect(fab).toBeVisible({ timeout: 10_000 })
  })

  test('clicking FAB toggles assistant panel', async ({ page }) => {
    await page.goto('/')
    const fab = page.getByLabel('Open shopping assistant')
    await expect(fab).toBeVisible({ timeout: 10_000 })
    await fab.click()
    // Wait for panel to render — check for aria-expanded change on the FAB
    await expect(page.locator('[aria-expanded="true"]')).toBeVisible({ timeout: 10_000 })
  })
})
