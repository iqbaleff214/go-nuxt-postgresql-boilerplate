<script setup lang="ts">
definePageMeta({ layout: 'default', middleware: 'auth' })

const authStore = useAuthStore()
</script>

<template>
  <div>
    <h1 class="text-2xl font-bold mb-2">Dashboard</h1>
    <p class="text-gray-500">Welcome back, {{ authStore.user?.displayName ?? authStore.user?.firstName }}!</p>
  </div>
</template>
