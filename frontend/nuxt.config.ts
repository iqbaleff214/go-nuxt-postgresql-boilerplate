import tailwindcss from '@tailwindcss/vite'

export default defineNuxtConfig({
  compatibilityDate: '2024-11-01',
  ssr: false,
  devtools: { enabled: true },

  modules: [
    '@pinia/nuxt',
    '@vee-validate/nuxt',
    '@nuxt/eslint',
  ],

  css: ['~/assets/css/main.css'],

  vite: {
    plugins: [tailwindcss()],
  },

  runtimeConfig: {
    public: {
      apiBase: '',   // overridden by NUXT_PUBLIC_API_BASE
      wsBase: '',    // overridden by NUXT_PUBLIC_WS_BASE
    },
  },

  pinia: {
    storesDirs: ['./stores/**'],
  },

  typescript: {
    strict: true,
    typeCheck: false,
  },
})
