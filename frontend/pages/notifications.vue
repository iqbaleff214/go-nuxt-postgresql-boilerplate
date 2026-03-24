<script setup lang="ts">
definePageMeta({ layout: 'default', middleware: 'auth' })

const notifStore = useNotificationsStore()

onMounted(() => notifStore.fetchNotifications(1))

async function markRead(id: string) {
  await notifStore.markRead(id)
}

async function markAll() {
  await notifStore.markAllRead()
}

function formatDate(iso: string) {
  return new Date(iso).toLocaleString()
}
</script>

<template>
  <div class="max-w-2xl">
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold">Notifications</h1>
      <button
        v-if="notifStore.unreadCount > 0"
        class="btn-secondary text-sm"
        @click="markAll"
      >
        Mark all read
      </button>
    </div>

    <div v-if="notifStore.notifications.length === 0" class="text-center text-gray-400 py-16">
      No notifications yet.
    </div>

    <div v-else class="space-y-2">
      <div
        v-for="n in notifStore.notifications"
        :key="n.id"
        :class="['card cursor-pointer transition', !n.readAt ? 'bg-blue-50 border-blue-200' : '']"
        @click="!n.readAt && markRead(n.id)"
      >
        <div class="flex justify-between items-start">
          <div>
            <p class="font-medium text-sm">{{ n.title }}</p>
            <p v-if="n.body" class="text-gray-500 text-sm mt-0.5">{{ n.body }}</p>
          </div>
          <div class="text-xs text-gray-400 shrink-0 ml-4">{{ formatDate(n.createdAt) }}</div>
        </div>
        <div v-if="!n.readAt" class="flex justify-end mt-2">
          <span class="inline-block w-2 h-2 bg-blue-500 rounded-full" />
        </div>
      </div>
    </div>
  </div>
</template>
