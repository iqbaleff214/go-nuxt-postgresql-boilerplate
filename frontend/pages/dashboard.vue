<script setup lang="ts">
definePageMeta({ layout: 'default', middleware: 'auth' })

const authStore = useAuthStore()
const notifStore = useNotificationsStore()
</script>

<template>
  <div class="space-y-8">
    <!-- Skeleton: header -->
    <div v-if="!authStore.user" class="flex items-center justify-between">
      <div class="space-y-2">
        <div class="skeleton h-7 w-64" />
        <div class="skeleton h-4 w-48" />
      </div>
      <div class="skeleton h-9 w-28 rounded-xl" />
    </div>

    <!-- Welcome header -->
    <div v-else class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">
          Welcome back, {{ authStore.user.displayName ?? authStore.user.firstName ?? 'there' }}
        </h1>
        <p class="mt-1 text-sm text-gray-500">Here's what's happening with your account today.</p>
      </div>
      <div class="hidden sm:flex items-center gap-2 rounded-xl bg-emerald-50 border border-emerald-100 px-4 py-2.5">
        <div class="w-2 h-2 rounded-full bg-emerald-500 animate-pulse" />
        <span class="text-sm font-medium text-emerald-700">Connected</span>
      </div>
    </div>

    <!-- Skeleton: stats -->
    <div v-if="!authStore.user" class="grid grid-cols-1 sm:grid-cols-3 gap-4">
      <div v-for="i in 3" :key="i" class="card flex items-center gap-4">
        <div class="skeleton w-12 h-12 rounded-xl shrink-0" />
        <div class="space-y-2 flex-1">
          <div class="skeleton h-3 w-16" />
          <div class="skeleton h-6 w-20" />
        </div>
      </div>
    </div>

    <!-- Stats -->
    <div v-else class="grid grid-cols-1 sm:grid-cols-3 gap-4">
      <div class="card flex items-center gap-4">
        <div class="w-12 h-12 rounded-xl bg-emerald-50 flex items-center justify-center shrink-0">
          <svg class="w-6 h-6 text-emerald-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.75" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
          </svg>
        </div>
        <div>
          <p class="text-xs font-semibold text-gray-400 uppercase tracking-wide">Role</p>
          <p class="text-lg font-bold text-gray-900 capitalize mt-0.5">{{ authStore.user.role }}</p>
        </div>
      </div>

      <div class="card flex items-center gap-4">
        <div class="w-12 h-12 rounded-xl bg-blue-50 flex items-center justify-center shrink-0">
          <svg class="w-6 h-6 text-blue-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.75" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
          </svg>
        </div>
        <div>
          <p class="text-xs font-semibold text-gray-400 uppercase tracking-wide">Unread</p>
          <p class="text-lg font-bold text-gray-900 mt-0.5">{{ notifStore.unreadCount }}</p>
        </div>
      </div>

      <div class="card flex items-center gap-4">
        <div class="w-12 h-12 rounded-xl bg-violet-50 flex items-center justify-center shrink-0">
          <svg class="w-6 h-6 text-violet-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.75" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
          </svg>
        </div>
        <div>
          <p class="text-xs font-semibold text-gray-400 uppercase tracking-wide">2FA</p>
          <p class="text-lg font-bold text-gray-900 mt-0.5">{{ authStore.user.is2faEnabled ? 'Enabled' : 'Disabled' }}</p>
        </div>
      </div>
    </div>

    <!-- Skeleton: quick actions -->
    <div v-if="!authStore.user" class="card">
      <div class="skeleton h-4 w-28 mb-4" />
      <div class="grid grid-cols-2 sm:grid-cols-4 gap-3">
        <div v-for="i in 4" :key="i" class="skeleton h-20 rounded-xl" />
      </div>
    </div>

    <!-- Quick links -->
    <div v-else class="card">
      <h2 class="text-sm font-semibold text-gray-700 mb-4">Quick actions</h2>
      <div class="grid grid-cols-2 sm:grid-cols-4 gap-3">
        <NuxtLink to="/profile" class="flex flex-col items-center gap-2 rounded-xl border border-gray-100 p-4 text-center hover:border-emerald-200 hover:bg-emerald-50 transition-all duration-150 group">
          <svg class="w-5 h-5 text-gray-400 group-hover:text-emerald-600 transition-colors" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.75" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
          </svg>
          <span class="text-xs font-medium text-gray-600 group-hover:text-emerald-700">Edit profile</span>
        </NuxtLink>
        <NuxtLink to="/profile/security" class="flex flex-col items-center gap-2 rounded-xl border border-gray-100 p-4 text-center hover:border-emerald-200 hover:bg-emerald-50 transition-all duration-150 group">
          <svg class="w-5 h-5 text-gray-400 group-hover:text-emerald-600 transition-colors" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.75" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
          </svg>
          <span class="text-xs font-medium text-gray-600 group-hover:text-emerald-700">Security</span>
        </NuxtLink>
        <NuxtLink to="/notifications" class="flex flex-col items-center gap-2 rounded-xl border border-gray-100 p-4 text-center hover:border-emerald-200 hover:bg-emerald-50 transition-all duration-150 group">
          <svg class="w-5 h-5 text-gray-400 group-hover:text-emerald-600 transition-colors" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.75" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
          </svg>
          <span class="text-xs font-medium text-gray-600 group-hover:text-emerald-700">Notifications</span>
        </NuxtLink>
        <NuxtLink v-if="authStore.user.role === 'superadmin'" to="/admin/users" class="flex flex-col items-center gap-2 rounded-xl border border-gray-100 p-4 text-center hover:border-emerald-200 hover:bg-emerald-50 transition-all duration-150 group">
          <svg class="w-5 h-5 text-gray-400 group-hover:text-emerald-600 transition-colors" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.75" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.75" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
          </svg>
          <span class="text-xs font-medium text-gray-600 group-hover:text-emerald-700">Admin panel</span>
        </NuxtLink>
      </div>
    </div>
  </div>
</template>
