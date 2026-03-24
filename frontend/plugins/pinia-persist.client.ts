import piniaPluginPersistedstate from 'pinia-plugin-persistedstate'

export default defineNuxtPlugin(async (nuxtApp) => {
  // @ts-expect-error pinia is attached by @pinia/nuxt
  nuxtApp.$pinia.use(piniaPluginPersistedstate)

  // After persistence is set up, restore user profile if we have a token but no user
  const authStore = useAuthStore()
  if (authStore.accessToken && !authStore.user) {
    try {
      await authStore.fetchMe()
    }
    catch {
      authStore.accessToken = null
    }
  }
})
