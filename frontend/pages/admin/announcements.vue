<script setup lang="ts">
import type { ApiError } from '~/types/api'

definePageMeta({ layout: 'admin', middleware: 'superadmin' })

const api = useApi()
const form = reactive({ title: '', body: '' })
const error = ref('')
const success = ref(false)
const sending = ref(false)

async function send() {
  error.value = ''
  success.value = false
  sending.value = true
  try {
    await api.post('/admin/announcements', { title: form.title, body: form.body })
    success.value = true
    form.title = ''
    form.body = ''
  }
  catch (err) {
    error.value = (err as ApiError).message ?? 'Failed to send announcement'
  }
  finally {
    sending.value = false
  }
}
</script>

<template>
  <div class="max-w-lg">
    <div class="mb-6">
      <h1 class="text-2xl font-bold text-gray-900 dark:text-gray-100">System announcement</h1>
      <p class="text-sm text-gray-500 dark:text-gray-400 mt-1">Broadcast a real-time notification to all active users.</p>
    </div>

    <form class="card space-y-5" @submit.prevent="send">
      <div v-if="error" class="flex items-center gap-2.5 rounded-xl border border-rose-200 bg-rose-50 p-3.5 text-sm text-rose-700">
        <svg class="w-4 h-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        {{ error }}
      </div>
      <div v-if="success" class="flex items-center gap-2.5 rounded-xl border border-emerald-200 bg-emerald-50 p-3.5 text-sm text-emerald-700">
        <svg class="w-4 h-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        Announcement sent to all active users.
      </div>

      <div>
        <label class="label">Title <span class="text-rose-500 normal-case font-medium">*</span></label>
        <input v-model="form.title" type="text" class="input" placeholder="e.g. Scheduled maintenance at 22:00 UTC" required />
      </div>

      <div>
        <label class="label">Message body</label>
        <textarea v-model="form.body" class="input resize-none" rows="5" placeholder="Optional details about this announcement…" />
      </div>

      <div class="flex items-center gap-3 pt-1">
        <button type="submit" class="btn-primary" :disabled="sending">
          <svg v-if="!sending" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5.882V19.24a1.76 1.76 0 01-3.417.592l-2.147-6.15M18 13a3 3 0 100-6M5.436 13.683A4.001 4.001 0 017 6h1.832c4.1 0 7.625-1.234 9.168-3v14c-1.543-1.766-5.067-3-9.168-3H7a3.988 3.988 0 01-1.564-.317z" />
          </svg>
          <svg v-else class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
          </svg>
          {{ sending ? 'Sending…' : 'Send announcement' }}
        </button>
        <p class="text-xs text-gray-400 dark:text-gray-500">Delivered via WebSocket to all online users</p>
      </div>
    </form>
  </div>
</template>
