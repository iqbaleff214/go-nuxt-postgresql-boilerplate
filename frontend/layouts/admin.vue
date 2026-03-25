<script setup lang="ts">
definePageMeta({ middleware: 'superadmin' })
</script>

<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-900 flex">
    <!-- Sidebar -->
    <aside class="w-60 bg-white dark:bg-gray-800 border-r border-gray-100 dark:border-gray-700 flex flex-col shrink-0 sticky top-0 h-screen">
      <!-- Logo -->
      <div class="h-16 flex items-center gap-2.5 px-5 border-b border-gray-100 dark:border-gray-700">
        <div class="w-7 h-7 rounded-lg bg-emerald-500 flex items-center justify-center">
          <svg class="w-3.5 h-3.5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 13l4 4L19 7" />
          </svg>
        </div>
        <div>
          <span class="font-bold text-gray-900 dark:text-gray-100 text-sm tracking-tight leading-none block">MyApp</span>
          <span class="text-[10px] font-semibold text-emerald-600 dark:text-emerald-400 uppercase tracking-widest">Admin</span>
        </div>
      </div>

      <!-- Nav -->
      <nav class="flex-1 px-3 py-4 space-y-0.5">
        <p class="text-[10px] font-semibold text-gray-400 uppercase tracking-widest px-3 mb-2">Management</p>
        <NuxtLink
          to="/admin/users"
          class="flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium text-gray-600 dark:text-gray-400 transition-all duration-150 hover:bg-gray-50 dark:hover:bg-gray-700 hover:text-gray-900 dark:hover:text-gray-100"
          active-class="bg-emerald-50 dark:bg-emerald-900/30 text-emerald-700 dark:text-emerald-400 font-semibold"
        >
          <svg class="w-4 h-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.75" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z" />
          </svg>
          Users
        </NuxtLink>
        <NuxtLink
          to="/admin/announcements"
          class="flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium text-gray-600 dark:text-gray-400 transition-all duration-150 hover:bg-gray-50 dark:hover:bg-gray-700 hover:text-gray-900 dark:hover:text-gray-100"
          active-class="bg-emerald-50 dark:bg-emerald-900/30 text-emerald-700 dark:text-emerald-400 font-semibold"
        >
          <svg class="w-4 h-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.75" d="M11 5.882V19.24a1.76 1.76 0 01-3.417.592l-2.147-6.15M18 13a3 3 0 100-6M5.436 13.683A4.001 4.001 0 017 6h1.832c4.1 0 7.625-1.234 9.168-3v14c-1.543-1.766-5.067-3-9.168-3H7a3.988 3.988 0 01-1.564-.317z" />
          </svg>
          Announcements
        </NuxtLink>
      </nav>

      <!-- Theme switcher + Back to app -->
      <div class="px-3 py-3 border-t border-gray-100 dark:border-gray-700 space-y-2">
        <div class="flex items-center justify-between px-2">
          <span class="text-[10px] font-semibold text-gray-400 uppercase tracking-widest">Theme</span>
          <ThemeSwitcher />
        </div>
        <NuxtLink
          to="/dashboard"
          class="flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium text-gray-500 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-700 hover:text-gray-700 dark:hover:text-gray-200 transition-colors"
        >
          <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.75" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
          </svg>
          Back to app
        </NuxtLink>
      </div>
    </aside>

    <!-- Main -->
    <div class="flex-1 flex flex-col min-w-0">
      <main class="flex-1 p-8">
        <slot />
      </main>
    </div>
  </div>
</template>
