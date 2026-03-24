<script setup lang="ts">
import type { ApiError } from '~/types/api'

definePageMeta({ layout: 'default', middleware: 'auth' })

const api = useApi()
const authStore = useAuthStore()

// ── Change password ───────────────────────────────────────────────────────────
const pwForm = reactive({ currentPassword: '', newPassword: '', confirmNewPassword: '' })
const pwError = ref('')
const pwSuccess = ref(false)

async function changePassword() {
  pwError.value = ''
  pwSuccess.value = false
  if (pwForm.newPassword !== pwForm.confirmNewPassword) {
    pwError.value = 'Passwords do not match'
    return
  }
  try {
    await api.post('/auth/change-password', {
      current_password: pwForm.currentPassword,
      new_password: pwForm.newPassword,
    })
    pwSuccess.value = true
    pwForm.currentPassword = ''
    pwForm.newPassword = ''
    pwForm.confirmNewPassword = ''
  }
  catch (err) {
    pwError.value = (err as ApiError).message ?? 'Failed'
  }
}

// ── 2FA ──────────────────────────────────────────────────────────────────────
const showDisableModal = ref(false)
const disableForm = reactive({ password: '', code: '' })
const disableError = ref('')

async function disable2fa() {
  disableError.value = ''
  try {
    await api.post('/auth/2fa/disable', {
      password: disableForm.password,
      code: disableForm.code,
    })
    if (authStore.user) authStore.user.is2faEnabled = false
    showDisableModal.value = false
  }
  catch (err) {
    disableError.value = (err as ApiError).message ?? 'Failed'
  }
}

// ── Recovery codes ────────────────────────────────────────────────────────────
const showRegenModal = ref(false)
const regenForm = reactive({ password: '', code: '' })
const regenError = ref('')
const recoveryCodes = ref<string[]>([])

async function regenerateCodes() {
  regenError.value = ''
  try {
    const res = await api.post<{ recovery_codes: string[] }>('/auth/2fa/recovery-codes/regenerate', {
      password: regenForm.password,
      code: regenForm.code,
    })
    recoveryCodes.value = res.data.recovery_codes
    showRegenModal.value = false
  }
  catch (err) {
    regenError.value = (err as ApiError).message ?? 'Failed'
  }
}

function copyRecoveryCodes() {
  navigator.clipboard.writeText(recoveryCodes.value.join('\n'))
}
</script>

<template>
  <div class="max-w-2xl space-y-8">
    <h1 class="text-2xl font-bold">Security</h1>

    <!-- Change password -->
    <section class="card">
      <h2 class="text-lg font-semibold mb-4">Change password</h2>
      <form class="space-y-4" @submit.prevent="changePassword">
        <div v-if="pwError" class="text-red-600 text-sm">{{ pwError }}</div>
        <div v-if="pwSuccess" class="text-green-600 text-sm">Password updated.</div>
        <div>
          <label class="label">Current password</label>
          <input v-model="pwForm.currentPassword" type="password" class="input" />
        </div>
        <div>
          <label class="label">New password</label>
          <input v-model="pwForm.newPassword" type="password" class="input" />
        </div>
        <div>
          <label class="label">Confirm new password</label>
          <input v-model="pwForm.confirmNewPassword" type="password" class="input" />
        </div>
        <button type="submit" class="btn-primary">Update password</button>
      </form>
    </section>

    <!-- 2FA -->
    <section class="card">
      <div class="flex items-center justify-between mb-4">
        <h2 class="text-lg font-semibold">Two-factor authentication</h2>
        <span :class="authStore.user?.is2faEnabled ? 'badge-success' : 'badge-neutral'">
          {{ authStore.user?.is2faEnabled ? 'Enabled' : 'Disabled' }}
        </span>
      </div>

      <div v-if="!authStore.user?.is2faEnabled">
        <p class="text-sm text-gray-500 mb-3">Add an extra layer of security to your account.</p>
        <NuxtLink to="/profile/security/2fa" class="btn-primary">Enable 2FA</NuxtLink>
      </div>
      <div v-else class="space-y-3">
        <button class="btn-secondary" @click="showDisableModal = true">Disable 2FA</button>
        <div>
          <h3 class="font-medium text-sm mb-2">Recovery codes</h3>
          <div v-if="recoveryCodes.length" class="bg-gray-50 rounded p-3 font-mono text-sm grid grid-cols-2 gap-1 mb-2">
            <span v-for="c in recoveryCodes" :key="c">{{ c }}</span>
          </div>
          <button class="btn-secondary text-sm" @click="showRegenModal = true">Regenerate recovery codes</button>
          <button v-if="recoveryCodes.length" class="btn-secondary text-sm ml-2" @click="copyRecoveryCodes">Copy codes</button>
        </div>
      </div>
    </section>

    <!-- Disable 2FA modal -->
    <div v-if="showDisableModal" class="modal-backdrop">
      <div class="modal">
        <h3 class="text-lg font-semibold mb-4">Disable 2FA</h3>
        <div class="space-y-3">
          <div v-if="disableError" class="text-red-600 text-sm">{{ disableError }}</div>
          <div>
            <label class="label">Password</label>
            <input v-model="disableForm.password" type="password" class="input" />
          </div>
          <div>
            <label class="label">Authenticator code</label>
            <input v-model="disableForm.code" type="text" class="input" maxlength="6" />
          </div>
          <div class="flex gap-2 justify-end">
            <button class="btn-secondary" @click="showDisableModal = false">Cancel</button>
            <button class="btn-danger" @click="disable2fa">Disable</button>
          </div>
        </div>
      </div>
    </div>

    <!-- Regen codes modal -->
    <div v-if="showRegenModal" class="modal-backdrop">
      <div class="modal">
        <h3 class="text-lg font-semibold mb-4">Regenerate recovery codes</h3>
        <div class="space-y-3">
          <div v-if="regenError" class="text-red-600 text-sm">{{ regenError }}</div>
          <div>
            <label class="label">Password</label>
            <input v-model="regenForm.password" type="password" class="input" />
          </div>
          <div>
            <label class="label">Authenticator code</label>
            <input v-model="regenForm.code" type="text" class="input" maxlength="6" />
          </div>
          <div class="flex gap-2 justify-end">
            <button class="btn-secondary" @click="showRegenModal = false">Cancel</button>
            <button class="btn-primary" @click="regenerateCodes">Regenerate</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
