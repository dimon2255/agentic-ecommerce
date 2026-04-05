export default defineNuxtRouteMiddleware(async (to) => {
  const user = useSupabaseUser()
  if (!user.value) {
    return navigateTo(`/auth/login?redirect=${encodeURIComponent(to.fullPath)}`)
  }

  const { permissions, fetchPermissions } = useAdminAuth()

  // Load permissions if not yet fetched
  if (permissions.value === null) {
    await fetchPermissions()
  }

  // No admin permissions — redirect to storefront
  if (!permissions.value?.length) {
    return navigateTo('/', { replace: true })
  }
})
