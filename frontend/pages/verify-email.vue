<script setup lang="ts">
import type { ApiError } from '~/types/api'

definePageMeta({ layout: 'auth' })

const route = useRoute()
const api = useApi()

const status = ref<'loading' | 'success' | 'error'>('loading')
const errorMessage = ref('')
const resendEmail = ref('')
const resendStatus = ref<'' | 'sent' | 'error'>('')

onMounted(async () => {
  const token = route.query.token as string
  if (!token) {
    status.value = 'error'
    errorMessage.value = 'No verification token found in URL.'
    return
  }
  try {
    await api.post('/auth/verify-email', { token })
    status.value = 'success'
  }
  catch (err) {
    status.value = 'error'
    errorMessage.value = (err as ApiError).message ?? 'Verification failed.'
  }
})

async function resend() {
  if (!resendEmail.value) return
  try {
    await api.post('/auth/resend-verification', { email: resendEmail.value })
    resendStatus.value = 'sent'
  }
  catch {
    resendStatus.value = 'error'
  }
}
</script>

<template>
  <div class="text-center">
    <div v-if="status === 'loading'" class="space-y-3">
      <div class="animate-spin h-8 w-8 border-4 border-primary border-t-transparent rounded-full mx-auto" />
      <p class="text-gray-500">Verifying your email…</p>
    </div>

    <div v-else-if="status === 'success'" class="space-y-4">
      <div class="text-5xl">✓</div>
      <h1 class="text-2xl font-bold">Email verified!</h1>
      <p class="text-gray-500">Your account is ready. You can now log in.</p>
      <NuxtLink to="/login" class="btn-primary inline-block">Go to login</NuxtLink>
    </div>

    <div v-else class="space-y-4">
      <h1 class="text-2xl font-bold">Verification failed</h1>
      <p class="text-red-600">{{ errorMessage }}</p>

      <div class="mt-6 text-left">
        <p class="text-sm text-gray-600 mb-2">Resend verification email:</p>
        <div v-if="resendStatus === 'sent'" class="text-green-600 text-sm">Verification email sent. Check your inbox.</div>
        <div v-else class="flex gap-2">
          <input v-model="resendEmail" type="email" class="input flex-1" placeholder="your@email.com" />
          <button class="btn-primary" @click="resend">Resend</button>
        </div>
        <p v-if="resendStatus === 'error'" class="text-red-500 text-xs mt-1">Failed to resend. Try again later.</p>
      </div>

      <NuxtLink to="/login" class="text-sm text-gray-500 hover:underline block mt-2">Back to login</NuxtLink>
    </div>
  </div>
</template>
