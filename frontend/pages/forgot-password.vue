<script setup lang="ts">
definePageMeta({ layout: 'auth', middleware: 'guest' })

const api = useApi()
const email = ref('')
const submitted = ref(false)

async function submit() {
  if (!email.value) return
  await api.post('/auth/forgot-password', { email: email.value }).catch(() => {})
  // Always show success — do not reveal whether email exists
  submitted.value = true
}
</script>

<template>
  <div>
    <div class="mb-8">
      <h1 class="text-2xl font-bold text-gray-900 dark:text-gray-100">Forgot your password?</h1>
      <p class="mt-1.5 text-sm text-gray-500 dark:text-gray-400">No worries, we'll send you reset instructions.</p>
    </div>

    <div class="bg-white dark:bg-gray-800 rounded-2xl border border-gray-100 dark:border-gray-700 shadow-sm p-8">
      <div v-if="submitted" class="text-center py-4 space-y-4">
        <div class="mx-auto w-14 h-14 rounded-full bg-emerald-100 flex items-center justify-center">
          <svg class="w-7 h-7 text-emerald-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.75" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
          </svg>
        </div>
        <div>
          <p class="font-semibold text-gray-900 dark:text-gray-100">Check your email</p>
          <p class="text-sm text-gray-500 dark:text-gray-400 mt-1">If an account with that email exists, you'll receive a reset link shortly.</p>
        </div>
        <NuxtLink to="/login" class="btn-secondary inline-flex mx-auto">Back to sign in</NuxtLink>
      </div>

      <form v-else class="space-y-5" @submit.prevent="submit">
        <div>
          <label class="label">Email address</label>
          <input v-model="email" type="email" class="input" placeholder="you@example.com" required autofocus />
        </div>

        <button type="submit" class="btn-primary w-full">Send reset link</button>

        <div class="text-center">
          <NuxtLink to="/login" class="text-sm font-medium text-gray-500 hover:text-gray-700 transition-colors">
            ← Back to sign in
          </NuxtLink>
        </div>
      </form>
    </div>
  </div>
</template>
