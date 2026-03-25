export default defineNuxtPlugin(() => {
  const themeStore = useThemeStore()

  function applyTheme(mode: string) {
    const isDark
      = mode === 'dark'
      || (mode === 'system' && window.matchMedia('(prefers-color-scheme: dark)').matches)
    document.documentElement.classList.toggle('dark', isDark)
  }

  applyTheme(themeStore.mode)

  watch(() => themeStore.mode, applyTheme)

  const mq = window.matchMedia('(prefers-color-scheme: dark)')
  mq.addEventListener('change', () => {
    if (themeStore.mode === 'system') applyTheme('system')
  })
})
