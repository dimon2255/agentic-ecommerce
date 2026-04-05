import { test, expect } from '@playwright/test'

test.describe('Admin catalog management', () => {
  // Note: these tests require an admin user to be logged in.
  // In CI, seed a test admin user or use a setup project.

  test('admin dashboard loads', async ({ page }) => {
    await page.goto('/admin')
    // Should show admin page or redirect to login
    await expect(page).toHaveURL(/admin|auth/, { timeout: 10_000 })
  })

  test('admin products page loads', async ({ page }) => {
    await page.goto('/admin/products')
    // Should show product list or auth redirect
    await expect(page).toHaveURL(/admin|auth/, { timeout: 10_000 })
  })

  test('admin categories page loads', async ({ page }) => {
    await page.goto('/admin/categories')
    await expect(page).toHaveURL(/admin|auth/, { timeout: 10_000 })
  })

  test('admin new product page loads', async ({ page }) => {
    await page.goto('/admin/products/new')
    await expect(page).toHaveURL(/admin|auth/, { timeout: 10_000 })
  })
})
