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
    <h1 class="text-2xl font-bold mb-2">Two-factor authentication</h1>
    <p class="text-gray-500 text-sm mb-6">Enter the code from your authenticator app.</p>

    <form class="space-y-4" @submit.prevent="submit">
      <div v-if="error" class="p-3 bg-red-50 border border-red-200 rounded text-red-700 text-sm">
        {{ error }}
        <span v-if="attempts >= 4" class="block mt-1 font-medium">Warning: only a few attempts remain before lockout.</span>
      </div>

      <div v-if="!useRecovery">
        <label class="block text-sm font-medium mb-1">6-digit code</label>
        <input
          v-model="code"
          type="text"
          inputmode="numeric"
          maxlength="6"
          pattern="\d{6}"
          class="input text-center text-2xl tracking-widest"
          placeholder="000000"
          autofocus
        />
      </div>
      <div v-else>
        <label class="block text-sm font-medium mb-1">Recovery code</label>
        <input v-model="code" type="text" class="input font-mono" placeholder="XXXXX-XXXXX" />
      </div>

      <button
        type="button"
        class="text-xs text-gray-500 hover:underline"
        @click="useRecovery = !useRecovery; code = ''"
      >
        {{ useRecovery ? 'Use authenticator app instead' : 'Use a recovery code instead' }}
      </button>

      <button type="submit" class="btn-primary w-full">Verify</button>

      <p class="text-center text-sm">
        <NuxtLink to="/login" class="text-gray-500 hover:underline">← Back to login</NuxtLink>
      </p>
    </form>
  </div>
</template>
