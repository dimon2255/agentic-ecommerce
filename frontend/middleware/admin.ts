export default defineNuxtRouteMiddleware(async (to) => {
  const user = useSupabaseUser()

  // On hard refresh, useSupabaseUser() may not be populated yet.
  // Wait for the Supabase client to restore the session from the cookie.
  if (!user.value) {
    const client = useSupabaseClient()
    const { data: { session } } = await client.auth.getSession()
    if (!session) {
      return navigateTo(`/auth/login?redirect=${encodeURIComponent(to.fullPath)}`)
    }
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
