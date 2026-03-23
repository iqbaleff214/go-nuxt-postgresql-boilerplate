<script setup lang="ts">
definePageMeta({ middleware: 'superadmin' })
</script>

<template>
  <div class="min-h-screen bg-gray-100 flex">
    <!-- Sidebar -->
    <aside class="w-56 bg-white border-r border-gray-200 flex flex-col">
      <div class="px-4 py-4 border-b border-gray-200">
        <NuxtLink to="/dashboard" class="text-lg font-semibold text-gray-900">MyApp</NuxtLink>
        <p class="text-xs text-gray-400 mt-0.5">Admin Panel</p>
      </div>
      <nav class="flex-1 px-2 py-4 space-y-1">
        <NuxtLink
          to="/admin/users"
          class="flex items-center gap-2 px-3 py-2 rounded-md text-sm text-gray-700 hover:bg-gray-100"
          active-class="bg-gray-100 font-medium"
        >
          Users
        </NuxtLink>
        <NuxtLink
          to="/admin/announcements"
          class="flex items-center gap-2 px-3 py-2 rounded-md text-sm text-gray-700 hover:bg-gray-100"
          active-class="bg-gray-100 font-medium"
        >
          Announcements
        </NuxtLink>
      </nav>
    </aside>

    <!-- Main -->
    <div class="flex-1 flex flex-col">
      <main class="p-6 flex-1">
        <slot />
      </main>
    </div>
  </div>
</template>
