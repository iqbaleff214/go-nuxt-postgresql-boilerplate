<script setup lang="ts">
definePageMeta({ middleware: 'auth' })

const authStore = useAuthStore()
const notifStore = useNotificationsStore()
useNotifications()

async function handleLogout() {
  await authStore.logout()
  await navigateTo('/login')
}

const initials = computed(() => {
  const u = authStore.user
  if (!u) return '?'
  return ((u.first_name?.[0] ?? '') + (u.last_name?.[0] ?? '')).toUpperCase() || u.email[0].toUpperCase()
})

const sidebarOpen = ref(false)
const route = useRoute()
watch(() => route.path, () => { sidebarOpen.value = false })
</script>

<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-900 lg:flex">

    <!-- Mobile header -->
    <header class="lg:hidden fixed top-0 inset-x-0 z-30 h-14 bg-white dark:bg-gray-800 border-b border-gray-100 dark:border-gray-700 flex items-center justify-between px-4">
      <button
        class="p-1.5 rounded-lg text-gray-500 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
        @click="sidebarOpen = true"
      >
        <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
        </svg>
      </button>
      <div class="flex items-center gap-2">
        <div class="w-6 h-6 rounded-md bg-emerald-500 flex items-center justify-center">
          <svg class="w-3 h-3 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 13l4 4L19 7" />
          </svg>
        </div>
        <span class="font-bold text-gray-900 dark:text-gray-100 tracking-tight">MyApp</span>
      </div>
      <ThemeSwitcher />
    </header>

    <!-- Sidebar backdrop -->
    <Transition
      enter-active-class="transition-opacity duration-200"
      leave-active-class="transition-opacity duration-200"
      enter-from-class="opacity-0"
      leave-to-class="opacity-0"
    >
      <div
        v-if="sidebarOpen"
        class="lg:hidden fixed inset-0 z-40 bg-black/40 backdrop-blur-sm"
        @click="sidebarOpen = false"
      />
    </Transition>

    <!-- Sidebar -->
    <aside
      class="fixed lg:sticky top-0 inset-y-0 left-0 z-50 w-64 lg:w-60 bg-white dark:bg-gray-800 border-r border-gray-100 dark:border-gray-700 flex flex-col h-screen transition-transform duration-300 ease-in-out lg:translate-x-0"
      :class="sidebarOpen ? 'translate-x-0' : '-translate-x-full'"
    >
      <!-- Logo -->
      <div class="h-16 flex items-center justify-between gap-2.5 px-5 border-b border-gray-100 dark:border-gray-700">
        <div class="flex items-center gap-2.5">
          <div class="w-7 h-7 rounded-lg bg-emerald-500 flex items-center justify-center shrink-0">
            <svg class="w-3.5 h-3.5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 13l4 4L19 7" />
            </svg>
          </div>
          <span class="font-bold text-gray-900 dark:text-gray-100 tracking-tight">MyApp</span>
        </div>
        <!-- Close button (mobile only) -->
        <button
          class="lg:hidden p-1.5 rounded-lg text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
          @click="sidebarOpen = false"
        >
          <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <!-- Nav -->
      <nav class="flex-1 px-3 py-4 space-y-0.5 overflow-y-auto">
        <p class="text-[10px] font-semibold text-gray-400 uppercase tracking-widest px-3 mb-2">Menu</p>

        <NuxtLink
          to="/dashboard"
          class="flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium text-gray-600 dark:text-gray-400 transition-all duration-150 hover:bg-gray-50 dark:hover:bg-gray-700 hover:text-gray-900 dark:hover:text-gray-100"
          active-class="bg-emerald-50 dark:bg-emerald-900/30 text-emerald-700 dark:text-emerald-400 font-semibold"
        >
          <svg class="w-4 h-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.75" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" />
          </svg>
          Dashboard
        </NuxtLink>

        <NuxtLink
          to="/notifications"
          class="flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium text-gray-600 dark:text-gray-400 transition-all duration-150 hover:bg-gray-50 dark:hover:bg-gray-700 hover:text-gray-900 dark:hover:text-gray-100"
          active-class="bg-emerald-50 dark:bg-emerald-900/30 text-emerald-700 dark:text-emerald-400 font-semibold"
        >
          <div class="relative shrink-0">
            <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.75" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
            </svg>
            <span
              v-if="notifStore.unreadCount > 0"
              class="absolute -top-1.5 -right-1.5 w-3.5 h-3.5 bg-rose-500 text-white text-[9px] font-bold rounded-full flex items-center justify-center"
            >{{ notifStore.unreadCount > 9 ? '9+' : notifStore.unreadCount }}</span>
          </div>
          Notifications
          <span v-if="notifStore.unreadCount > 0" class="ml-auto text-[10px] font-bold bg-rose-100 dark:bg-rose-900/40 text-rose-600 dark:text-rose-400 rounded-full px-1.5 py-0.5">
            {{ notifStore.unreadCount }}
          </span>
        </NuxtLink>

        <p class="text-[10px] font-semibold text-gray-400 uppercase tracking-widest px-3 pt-4 mb-2">Account</p>

        <NuxtLink
          to="/profile"
          class="flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium text-gray-600 dark:text-gray-400 transition-all duration-150 hover:bg-gray-50 dark:hover:bg-gray-700 hover:text-gray-900 dark:hover:text-gray-100"
          active-class="bg-emerald-50 dark:bg-emerald-900/30 text-emerald-700 dark:text-emerald-400 font-semibold"
        >
          <svg class="w-4 h-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.75" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
          </svg>
          Profile
        </NuxtLink>

        <NuxtLink
          to="/profile/security"
          class="flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium text-gray-600 dark:text-gray-400 transition-all duration-150 hover:bg-gray-50 dark:hover:bg-gray-700 hover:text-gray-900 dark:hover:text-gray-100"
          active-class="bg-emerald-50 dark:bg-emerald-900/30 text-emerald-700 dark:text-emerald-400 font-semibold"
        >
          <svg class="w-4 h-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.75" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
          </svg>
          Security
        </NuxtLink>

        <!-- Admin link — superadmin only -->
        <template v-if="authStore.user?.role === 'superadmin'">
          <p class="text-[10px] font-semibold text-gray-400 uppercase tracking-widest px-3 pt-4 mb-2">Admin</p>
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
        </template>
      </nav>

      <!-- User footer -->
      <div class="px-3 py-3 border-t border-gray-100 dark:border-gray-700">
        <!-- Skeleton while user loads -->
        <div v-if="!authStore.user" class="flex items-center gap-3 px-2">
          <div class="skeleton w-8 h-8 rounded-full shrink-0" />
          <div class="flex-1 min-w-0 space-y-1.5">
            <div class="skeleton h-3 w-24" />
            <div class="skeleton h-2.5 w-14" />
          </div>
        </div>
        <!-- Loaded -->
        <div v-else class="flex items-center gap-3 px-2">
          <div v-if="authStore.user.avatar_url" class="shrink-0">
            <img :src="authStore.user.avatar_url" class="w-8 h-8 rounded-full object-cover" />
          </div>
          <div v-else class="w-8 h-8 rounded-full bg-emerald-100 dark:bg-emerald-900/40 text-emerald-700 dark:text-emerald-400 flex items-center justify-center text-xs font-bold shrink-0">
            {{ initials }}
          </div>
          <div class="flex-1 min-w-0">
            <p class="text-xs font-semibold text-gray-900 dark:text-gray-100 truncate">
              {{ authStore.user.display_name || authStore.user.first_name || authStore.user.email }}
            </p>
            <p class="text-[11px] text-gray-400 truncate capitalize">{{ authStore.user.role }}</p>
          </div>
          <ThemeSwitcher />
          <button
            class="p-1.5 rounded-lg text-gray-400 hover:text-rose-500 hover:bg-rose-50 dark:hover:bg-rose-900/30 transition-colors shrink-0"
            title="Sign out"
            @click="handleLogout"
          >
            <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.75" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
            </svg>
          </button>
        </div>
      </div>
    </aside>

    <!-- Main content -->
    <div class="flex-1 min-w-0 pt-14 lg:pt-0">
      <main class="p-4 sm:p-6 lg:p-8">
        <slot />
      </main>
    </div>
  </div>
</template>
