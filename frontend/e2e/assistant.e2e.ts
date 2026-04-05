import { test, expect } from '@playwright/test'

test.describe('AI Assistant', () => {
  test('assistant FAB is visible on page', async ({ page }) => {
    await page.goto('/')
    // The floating action button should be visible
    const fab = page.getByRole('button', { name: /assistant|chat/i })
    await expect(fab).toBeVisible({ timeout: 10_000 })
  })

  test('clicking FAB opens assistant panel', async ({ page }) => {
    await page.goto('/')
    const fab = page.getByRole('button', { name: /assistant|chat/i })
    await expect(fab).toBeVisible({ timeout: 10_000 })
    await fab.click()

    // Panel should open with input field
    await expect(page.getByPlaceholder(/ask|message|type/i)).toBeVisible({ timeout: 5_000 })
  })

  test('can close assistant panel', async ({ page }) => {
    await page.goto('/')
    const fab = page.getByRole('button', { name: /assistant|chat/i })
    await fab.click()

    // Find close button and click it
    const closeBtn = page.getByRole('button', { name: /close/i })
    if (await closeBtn.isVisible()) {
      await closeBtn.click()
    } else {
      // FAB toggles — click again to close
      await page.getByRole('button', { name: /assistant|chat|close/i }).first().click()
    }

    // Input should no longer be visible
    await expect(page.getByPlaceholder(/ask|message|type/i)).not.toBeVisible({ timeout: 3_000 })
  })
})
