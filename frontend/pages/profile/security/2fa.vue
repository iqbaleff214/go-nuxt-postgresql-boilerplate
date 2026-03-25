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
    if (authStore.user) authStore.user.is_2fa_enabled = true
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
    <div class="mb-8">
      <h1 class="text-2xl font-bold text-gray-900">Enable two-factor authentication</h1>
      <p class="text-sm text-gray-500 mt-1">Protect your account with an authenticator app</p>
    </div>

    <!-- Progress steps -->
    <div class="flex items-center gap-2 mb-8">
      <div v-for="i in 3" :key="i" class="flex items-center gap-2">
        <div :class="[
          'w-7 h-7 rounded-full flex items-center justify-center text-xs font-bold transition-all',
          step > i
            ? 'bg-emerald-500 text-white'
            : step === i
              ? 'bg-emerald-100 text-emerald-700 ring-2 ring-emerald-500'
              : 'bg-gray-100 text-gray-400',
        ]">
          <svg v-if="step > i" class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 13l4 4L19 7" />
          </svg>
          <span v-else>{{ i }}</span>
        </div>
        <div v-if="i < 3" :class="['h-px w-8 transition-all', step > i ? 'bg-emerald-500' : 'bg-gray-200']" />
      </div>
      <span class="ml-2 text-xs font-medium text-gray-500">
        {{ step === 1 ? 'Scan QR code' : step === 2 ? 'Verify code' : 'Save recovery codes' }}
      </span>
    </div>

    <!-- Step 1: QR Code -->
    <div v-if="step === 1" class="card space-y-6">
      <div>
        <h2 class="font-semibold text-gray-900">Scan with your authenticator app</h2>
        <p class="text-sm text-gray-500 mt-1">Use Google Authenticator, Authy, or any TOTP app.</p>
      </div>
      <div v-if="setupError" class="flex items-center gap-2.5 rounded-xl border border-rose-200 bg-rose-50 p-3.5 text-sm text-rose-700">
        <svg class="w-4 h-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        {{ setupError }}
      </div>
      <div class="flex justify-center">
        <div v-if="qrDataURI" class="p-3 bg-white rounded-xl border border-gray-100 shadow-sm">
          <img :src="qrDataURI" alt="QR code" class="w-48 h-48" />
        </div>
        <div v-else class="w-48 h-48 rounded-xl bg-gray-100 animate-pulse" />
      </div>
      <details class="text-sm">
        <summary class="cursor-pointer text-gray-500 hover:text-gray-700 font-medium transition-colors">Can't scan? Enter key manually</summary>
        <code class="block mt-3 break-all text-xs bg-gray-50 border border-gray-100 p-3 rounded-xl text-gray-600">{{ otpauthURL }}</code>
      </details>
      <button class="btn-primary w-full" :disabled="!qrDataURI" @click="step = 2">Continue</button>
    </div>

    <!-- Step 2: Confirm code -->
    <div v-else-if="step === 2" class="card space-y-6">
      <div>
        <h2 class="font-semibold text-gray-900">Enter the verification code</h2>
        <p class="text-sm text-gray-500 mt-1">Enter the 6-digit code shown in your authenticator app.</p>
      </div>
      <div v-if="setupError" class="flex items-center gap-2.5 rounded-xl border border-rose-200 bg-rose-50 p-3.5 text-sm text-rose-700">
        <svg class="w-4 h-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        {{ setupError }}
      </div>
      <input
        v-model="totpCode"
        type="text"
        inputmode="numeric"
        maxlength="6"
        class="input text-center text-3xl tracking-[0.5em] font-mono"
        placeholder="000000"
        autofocus
      />
      <div class="flex gap-3">
        <button class="btn-secondary flex-1" @click="step = 1">Back</button>
        <button class="btn-primary flex-1" @click="confirmSetup">Verify</button>
      </div>
    </div>

    <!-- Step 3: Recovery codes -->
    <div v-else class="card space-y-5">
      <div>
        <h2 class="font-semibold text-gray-900">Save your recovery codes</h2>
        <p class="text-sm text-gray-500 mt-1">These codes can be used to access your account if you lose your authenticator.</p>
      </div>
      <div class="flex items-start gap-3 rounded-xl border border-amber-200 bg-amber-50 p-4 text-sm text-amber-800">
        <svg class="w-5 h-5 shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
        </svg>
        <div>
          <p class="font-semibold">Store these in a safe place</p>
          <p class="mt-0.5 text-amber-700">These codes will not be shown again.</p>
        </div>
      </div>
      <div class="bg-gray-50 rounded-xl border border-gray-100 p-4 font-mono text-sm grid grid-cols-2 gap-2">
        <span v-for="c in recoveryCodes" :key="c" class="text-gray-700">{{ c }}</span>
      </div>
      <div class="flex gap-3">
        <button class="btn-secondary flex-1" @click="copyAll">Copy all</button>
        <button class="btn-secondary flex-1" @click="downloadCodes">Download .txt</button>
      </div>
      <NuxtLink to="/profile/security" class="btn-primary w-full text-center">Done — go to security settings</NuxtLink>
    </div>
  </div>
</template>
