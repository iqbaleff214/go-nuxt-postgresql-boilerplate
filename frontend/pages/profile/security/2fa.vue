<script setup lang="ts">
import type { ApiError } from '~/types/api'

definePageMeta({ layout: 'default', middleware: 'auth' })

const api = useApi()
const authStore = useAuthStore()

const step = ref(1)
const qrDataURI = ref('')
const otpauthURL = ref('')
const totpCode = ref('')
const setupError = ref('')
const recoveryCodes = ref<string[]>([])

onMounted(async () => {
  try {
    const res = await api.post<{ qr_data_uri: string; otpauth_url: string }>('/auth/2fa/setup')
    qrDataURI.value = res.data.qr_data_uri
    otpauthURL.value = res.data.otpauth_url
  }
  catch (err) {
    setupError.value = (err as ApiError).message ?? 'Setup failed'
  }
})

async function confirmSetup() {
  setupError.value = ''
  try {
    const res = await api.post<{ recovery_codes: string[] }>('/auth/2fa/confirm', { code: totpCode.value })
    recoveryCodes.value = res.data.recovery_codes
    if (authStore.user) authStore.user.is2faEnabled = true
    step.value = 3
  }
  catch (err) {
    setupError.value = (err as ApiError).message ?? 'Invalid code'
  }
}

function copyAll() {
  navigator.clipboard.writeText(recoveryCodes.value.join('\n'))
}

function downloadCodes() {
  const blob = new Blob([recoveryCodes.value.join('\n')], { type: 'text/plain' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = 'recovery-codes.txt'
  a.click()
  URL.revokeObjectURL(url)
}
</script>

<template>
  <div class="max-w-lg">
    <h1 class="text-2xl font-bold mb-6">Enable two-factor authentication</h1>

    <!-- Step 1: QR Code -->
    <div v-if="step === 1" class="card space-y-4">
      <h2 class="font-semibold">Step 1 — Scan QR code</h2>
      <div v-if="setupError" class="text-red-600 text-sm">{{ setupError }}</div>
      <div v-if="qrDataURI" class="flex justify-center">
        <img :src="qrDataURI" alt="QR code" class="w-48 h-48" />
      </div>
      <div v-else class="animate-pulse h-48 w-48 bg-gray-200 rounded mx-auto" />
      <details class="text-sm">
        <summary class="cursor-pointer text-gray-500">Enter key manually instead</summary>
        <code class="block mt-2 break-all text-xs bg-gray-50 p-2 rounded">{{ otpauthURL }}</code>
      </details>
      <button class="btn-primary w-full" :disabled="!qrDataURI" @click="step = 2">Next →</button>
    </div>

    <!-- Step 2: Confirm code -->
    <div v-else-if="step === 2" class="card space-y-4">
      <h2 class="font-semibold">Step 2 — Enter code</h2>
      <p class="text-sm text-gray-500">Enter the 6-digit code from your authenticator app.</p>
      <div v-if="setupError" class="text-red-600 text-sm">{{ setupError }}</div>
      <input
        v-model="totpCode"
        type="text"
        inputmode="numeric"
        maxlength="6"
        class="input text-center text-2xl tracking-widest"
        placeholder="000000"
        autofocus
      />
      <div class="flex gap-2">
        <button class="btn-secondary flex-1" @click="step = 1">← Back</button>
        <button class="btn-primary flex-1" @click="confirmSetup">Verify</button>
      </div>
    </div>

    <!-- Step 3: Recovery codes -->
    <div v-else class="card space-y-4">
      <h2 class="font-semibold">Step 3 — Save recovery codes</h2>
      <div class="bg-amber-50 border border-amber-200 rounded p-3 text-sm text-amber-800">
        ⚠️ Save these now — they will not be shown again.
      </div>
      <div class="bg-gray-50 rounded p-3 font-mono text-sm grid grid-cols-2 gap-1">
        <span v-for="c in recoveryCodes" :key="c">{{ c }}</span>
      </div>
      <div class="flex gap-2">
        <button class="btn-secondary flex-1" @click="copyAll">Copy all</button>
        <button class="btn-secondary flex-1" @click="downloadCodes">Download .txt</button>
      </div>
      <NuxtLink to="/profile/security" class="btn-primary block text-center">Done</NuxtLink>
    </div>
  </div>
</template>
