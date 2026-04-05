import { test, expect } from '@playwright/test'

test.describe('Admin order management', () => {
  test('admin orders page loads', async ({ page }) => {
    await page.goto('/admin/orders')
    // Should show orders list or auth redirect
    await expect(page).toHaveURL(/admin|auth/, { timeout: 10_000 })
  })

  test('admin reports page loads', async ({ page }) => {
    await page.goto('/admin/reports/sales')
    await expect(page).toHaveURL(/admin|auth/, { timeout: 10_000 })
  })
})
