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
    <h1 class="text-2xl font-bold mb-6">System announcement</h1>
    <p class="text-sm text-gray-500 mb-6">
      Broadcast a notification to all active users in real-time via WebSocket.
    </p>

    <form class="card space-y-4" @submit.prevent="send">
      <div v-if="error" class="p-3 bg-red-50 border border-red-200 rounded text-red-700 text-sm">{{ error }}</div>
      <div v-if="success" class="p-3 bg-green-50 border border-green-200 rounded text-green-700 text-sm">
        Announcement sent to all active users.
      </div>

      <div>
        <label class="label">Title <span class="text-red-500">*</span></label>
        <input v-model="form.title" type="text" class="input" placeholder="Scheduled maintenance at 22:00 UTC" required />
      </div>

      <div>
        <label class="label">Body</label>
        <textarea v-model="form.body" class="input" rows="4" placeholder="Optional details…" />
      </div>

      <button type="submit" class="btn-primary" :disabled="sending">
        {{ sending ? 'Sending…' : 'Send announcement' }}
      </button>
    </form>
  </div>
</template>
