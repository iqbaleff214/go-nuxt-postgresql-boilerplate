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
    <h1 class="text-2xl font-bold mb-2">Forgot password</h1>

    <div v-if="submitted" class="p-4 bg-green-50 border border-green-200 rounded-lg text-green-800">
      If an account with that email exists, you'll receive a reset link shortly.
    </div>

    <form v-else class="space-y-4" @submit.prevent="submit">
      <p class="text-gray-500 text-sm">Enter your email and we'll send you a reset link.</p>

      <div>
        <label class="block text-sm font-medium mb-1">Email</label>
        <input v-model="email" type="email" class="input" placeholder="jane@example.com" required />
      </div>

      <button type="submit" class="btn-primary w-full">Send reset link</button>

      <p class="text-center text-sm">
        <NuxtLink to="/login" class="text-gray-500 hover:underline">Back to login</NuxtLink>
      </p>
    </form>
  </div>
</template>
