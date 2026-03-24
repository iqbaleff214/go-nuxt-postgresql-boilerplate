export default defineNuxtRouteMiddleware(() => {
  if (import.meta.server) return
  const authStore = useAuthStore()
  if (!authStore.accessToken) {
    return navigateTo('/login')
  }
  if (authStore.user?.role !== 'superadmin') {
    return navigateTo('/dashboard')
  }
})
