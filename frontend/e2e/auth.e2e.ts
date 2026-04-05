import { test, expect } from '@playwright/test'

test.describe('Authentication', () => {
  test('login page loads', async ({ page }) => {
    await page.goto('/auth/login')
    await expect(page.getByRole('button', { name: /sign in|log in/i })).toBeVisible({ timeout: 10_000 })
  })

  test('register page loads', async ({ page }) => {
    await page.goto('/auth/register')
    await expect(page.getByRole('button', { name: /sign up|register|create/i })).toBeVisible({ timeout: 10_000 })
  })

  test('login with invalid credentials shows error', async ({ page }) => {
    await page.goto('/auth/login')

    await page.getByPlaceholder(/email/i).fill('invalid@test.com')
    await page.getByPlaceholder(/password/i).fill('wrongpassword')
    await page.getByRole('button', { name: /sign in|log in/i }).click()

    // Should show an error message
    await expect(page.locator('text=/error|invalid|incorrect/i')).toBeVisible({ timeout: 10_000 })
  })
})
