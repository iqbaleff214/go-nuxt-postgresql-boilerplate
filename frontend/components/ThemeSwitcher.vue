<script setup lang="ts">
import type { ThemeMode } from '~/stores/theme'

const themeStore = useThemeStore()

const cycle: ThemeMode[] = ['light', 'dark', 'system']

function toggle() {
  const next = cycle[(cycle.indexOf(themeStore.mode) + 1) % cycle.length]
  themeStore.setMode(next)
}

const label = computed(() => ({ light: 'Light', dark: 'Dark', system: 'System' }[themeStore.mode]))
</script>

<template>
  <button
    type="button"
    :title="`Theme: ${label}`"
    class="p-1.5 rounded-lg text-gray-400 dark:text-gray-500 hover:text-gray-600 dark:hover:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
    @click="toggle"
  >
    <!-- Sun (light) -->
    <svg v-if="themeStore.mode === 'light'" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364-6.364l-.707.707M6.343 17.657l-.707.707M17.657 17.657l-.707-.707M6.343 6.343l-.707-.707M12 7a5 5 0 100 10A5 5 0 0012 7z" />
    </svg>
    <!-- Moon (dark) -->
    <svg v-else-if="themeStore.mode === 'dark'" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z" />
    </svg>
    <!-- Monitor (system) -->
    <svg v-else class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
    </svg>
  </button>
</template>
