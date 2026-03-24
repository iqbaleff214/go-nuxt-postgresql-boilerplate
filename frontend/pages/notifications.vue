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
      <div>
        <h1 class="text-2xl font-bold text-gray-900">Notifications</h1>
        <p class="text-sm text-gray-500 mt-0.5">Stay up to date with your account activity</p>
      </div>
      <button
        v-if="notifStore.unreadCount > 0"
        class="btn-secondary text-sm"
        @click="markAll"
      >
        Mark all read
      </button>
    </div>

    <!-- Empty state -->
    <div v-if="notifStore.notifications.length === 0" class="card text-center py-16">
      <div class="mx-auto w-14 h-14 rounded-full bg-gray-100 flex items-center justify-center mb-4">
        <svg class="w-7 h-7 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
        </svg>
      </div>
      <p class="font-semibold text-gray-700">You're all caught up</p>
      <p class="text-sm text-gray-400 mt-1">No notifications yet.</p>
    </div>

    <!-- List -->
    <div v-else class="space-y-2">
      <div
        v-for="n in notifStore.notifications"
        :key="n.id"
        :class="[
          'bg-white rounded-2xl border transition-all duration-150 cursor-pointer overflow-hidden flex',
          !n.readAt
            ? 'border-emerald-200 shadow-sm hover:shadow-md'
            : 'border-gray-100 hover:border-gray-200',
        ]"
        @click="!n.readAt && markRead(n.id)"
      >
        <!-- Unread indicator strip -->
        <div :class="['w-1 shrink-0', !n.readAt ? 'bg-emerald-500' : 'bg-transparent']" />

        <div class="flex-1 px-5 py-4">
          <div class="flex justify-between items-start gap-4">
            <div class="flex-1 min-w-0">
              <p :class="['text-sm', !n.readAt ? 'font-semibold text-gray-900' : 'font-medium text-gray-700']">
                {{ n.title }}
              </p>
              <p v-if="n.body" class="text-sm text-gray-500 mt-0.5 truncate">{{ n.body }}</p>
            </div>
            <div class="flex items-center gap-2 shrink-0">
              <span class="text-xs text-gray-400">{{ formatDate(n.createdAt) }}</span>
              <span v-if="!n.readAt" class="w-2 h-2 rounded-full bg-emerald-500 shrink-0" />
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
