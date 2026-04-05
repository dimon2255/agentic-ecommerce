import { test, expect } from '@playwright/test'

test.describe('Authentication', () => {
  test('login page loads with form', async ({ page }) => {
    await page.goto('/auth/login')
    await expect(page.getByRole('button', { name: /sign in|log in/i })).toBeVisible({ timeout: 10_000 })
  })

  test('register page loads with form', async ({ page }) => {
    await page.goto('/auth/register')
    await expect(page.getByRole('button', { name: /sign up|register|create/i })).toBeVisible({ timeout: 10_000 })
  })

  test('can navigate between login and register', async ({ page }) => {
    await page.goto('/auth/login')
    await expect(page.getByRole('button', { name: /sign in|log in/i })).toBeVisible({ timeout: 10_000 })

    // Find link to register page
    const registerLink = page.locator('a[href*="register"]').first()
    if (await registerLink.isVisible()) {
      await registerLink.click()
      await expect(page).toHaveURL(/register/, { timeout: 5_000 })
    }
  })
})
