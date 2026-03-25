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
  <div class="bg-white dark:bg-gray-800 rounded-2xl border border-gray-100 dark:border-gray-700 shadow-sm p-6 sm:p-10 text-center">
    <!-- Loading -->
    <div v-if="status === 'loading'" class="space-y-4">
      <div class="mx-auto w-14 h-14 rounded-full bg-gray-100 dark:bg-gray-700 flex items-center justify-center">
        <svg class="w-6 h-6 text-gray-400 animate-spin" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
        </svg>
      </div>
      <p class="text-gray-500 text-sm">Verifying your email address…</p>
    </div>

    <!-- Success -->
    <div v-else-if="status === 'success'" class="space-y-5">
      <div class="mx-auto w-14 h-14 rounded-full bg-emerald-100 flex items-center justify-center">
        <svg class="w-7 h-7 text-emerald-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
        </svg>
      </div>
      <div>
        <h1 class="text-xl font-bold text-gray-900 dark:text-gray-100">Email verified!</h1>
        <p class="text-sm text-gray-500 dark:text-gray-400 mt-1.5">Your account is ready. You can now sign in.</p>
      </div>
      <NuxtLink to="/login" class="btn-primary inline-flex">Go to sign in</NuxtLink>
    </div>

    <!-- Error -->
    <div v-else class="space-y-6">
      <div class="mx-auto w-14 h-14 rounded-full bg-rose-100 flex items-center justify-center">
        <svg class="w-7 h-7 text-rose-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
        </svg>
      </div>
      <div>
        <h1 class="text-xl font-bold text-gray-900 dark:text-gray-100">Verification failed</h1>
        <p class="text-sm text-rose-600 dark:text-rose-400 mt-1.5">{{ errorMessage }}</p>
      </div>

      <div class="text-left space-y-3 border-t border-gray-100 dark:border-gray-700 pt-6">
        <p class="text-sm font-medium text-gray-700 dark:text-gray-300">Resend verification email</p>
        <div v-if="resendStatus === 'sent'" class="flex items-center gap-2 rounded-xl border border-emerald-200 bg-emerald-50 p-3.5 text-sm text-emerald-700">
          <svg class="w-4 h-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          Verification email sent. Check your inbox.
        </div>
        <div v-else class="flex gap-2">
          <input v-model="resendEmail" type="email" class="input flex-1" placeholder="your@email.com" />
          <button class="btn-primary" @click="resend">Resend</button>
        </div>
        <p v-if="resendStatus === 'error'" class="text-xs text-rose-600">Failed to resend. Try again later.</p>
      </div>

      <NuxtLink to="/login" class="text-sm font-medium text-gray-500 hover:text-gray-700 transition-colors">
        ← Back to sign in
      </NuxtLink>
    </div>
  </div>
</template>
