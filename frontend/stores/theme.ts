import { defineStore } from 'pinia'

export type ThemeMode = 'light' | 'dark' | 'system'

export const useThemeStore = defineStore('theme', {
  state: () => ({
    mode: 'system' as ThemeMode,
  }),

  actions: {
    setMode(mode: ThemeMode) {
      this.mode = mode
    },
  },

  persist: {
    pick: ['mode'],
  },
})
