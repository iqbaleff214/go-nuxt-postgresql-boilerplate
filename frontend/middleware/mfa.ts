export default defineNuxtRouteMiddleware((to) => {
  const authStore = useAuthStore()
  if (authStore.mfaChallengeToken && to.path !== '/login/2fa') {
    return navigateTo('/login/2fa')
  }
})
