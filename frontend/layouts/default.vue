<script setup lang="ts">
definePageMeta({ middleware: 'auth' })

const authStore = useAuthStore()
const notifStore = useNotificationsStore()
const { wsConnected } = useNotifications()

async function handleLogout() {
  await authStore.logout()
  await navigateTo('/login')
}
</script>

<template>
  <div class="min-h-screen bg-gray-100">
    <!-- Navbar -->
    <nav class="bg-white border-b border-gray-200 px-4 py-3 flex items-center justify-between">
      <NuxtLink to="/dashboard" class="text-lg font-semibold text-gray-900">
        MyApp
      </NuxtLink>

      <div class="flex items-center gap-4">
        <!-- Notification bell -->
        <NuxtLink to="/notifications" class="relative text-gray-600 hover:text-gray-900">
          <span class="text-xl">🔔</span>
          <span
            v-if="notifStore.unreadCount > 0"
            class="absolute -top-1 -right-1 bg-red-500 text-white text-xs rounded-full w-4 h-4 flex items-center justify-center"
          >
            {{ notifStore.unreadCount > 9 ? '9+' : notifStore.unreadCount }}
          </span>
        </NuxtLink>

        <!-- User menu -->
        <div class="flex items-center gap-2 text-sm text-gray-700">
          <NuxtLink to="/profile" class="hover:underline">
            {{ authStore.user?.display_name || authStore.user?.email }}
          </NuxtLink>
          <button class="text-red-500 hover:underline" @click="handleLogout">
            Logout
          </button>
        </div>
      </div>
    </nav>

    <!-- Page content -->
    <main class="p-6">
      <slot />
    </main>
  </div>
</template>
