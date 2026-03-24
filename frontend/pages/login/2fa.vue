<script setup lang="ts">
import type { ApiError } from '~/types/api'

definePageMeta({ layout: 'auth' })

const authStore = useAuthStore()

// If no challenge token, user arrived here directly — send back to login
onMounted(() => {
  if (!authStore.mfaChallengeToken) navigateTo('/login')
})

const code = ref('')
const useRecovery = ref(false)
const error = ref('')
const attempts = ref(0)

async function submit() {
  error.value = ''
  try {
    await authStore.verify2fa(code.value)
    await navigateTo('/dashboard')
  }
  catch (err) {
    attempts.value++
    error.value = (err as ApiError).message ?? 'Invalid code'
  }
}
</script>

<template>
  <div>
    <div class="mb-8">
      <h1 class="text-2xl font-bold text-gray-900">Two-factor authentication</h1>
      <p class="mt-1.5 text-sm text-gray-500">
        {{ useRecovery ? 'Enter one of your recovery codes.' : 'Enter the 6-digit code from your authenticator app.' }}
      </p>
    </div>

    <div class="bg-white rounded-2xl border border-gray-100 shadow-sm p-8">
      <form class="space-y-5" @submit.prevent="submit">
        <div v-if="error" class="rounded-xl border border-rose-200 bg-rose-50 p-3.5 text-sm text-rose-700 space-y-1">
          <div class="flex items-center gap-2">
            <svg class="w-4 h-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            {{ error }}
          </div>
          <p v-if="attempts >= 4" class="font-semibold pl-6">Warning: only a few attempts remain before lockout.</p>
        </div>

        <div v-if="!useRecovery">
          <label class="label">Authentication code</label>
          <input
            v-model="code"
            type="text"
            inputmode="numeric"
            maxlength="6"
            pattern="\d{6}"
            class="input text-center text-3xl tracking-[0.5em] font-mono"
            placeholder="000000"
            autofocus
          />
        </div>
        <div v-else>
          <label class="label">Recovery code</label>
          <input v-model="code" type="text" class="input font-mono" placeholder="XXXXX-XXXXX" autofocus />
        </div>

        <button type="submit" class="btn-primary w-full">Verify</button>

        <div class="text-center space-y-2">
          <button
            type="button"
            class="text-xs font-medium text-emerald-600 hover:text-emerald-700 transition-colors"
            @click="useRecovery = !useRecovery; code = ''"
          >
            {{ useRecovery ? 'Use authenticator app instead' : 'Use a recovery code instead' }}
          </button>
          <div>
            <NuxtLink to="/login" class="text-xs text-gray-400 hover:text-gray-600 transition-colors">← Back to sign in</NuxtLink>
          </div>
        </div>
      </form>
    </div>
  </div>
</template>
