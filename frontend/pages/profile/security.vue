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
  <div class="max-w-2xl space-y-6">
    <div>
      <h1 class="text-2xl font-bold text-gray-900">Security</h1>
      <p class="text-sm text-gray-500 mt-1">Manage your password and two-factor authentication</p>
    </div>

    <!-- Change password -->
    <div class="card">
      <h2 class="text-base font-semibold text-gray-900 mb-5">Change password</h2>
      <form class="space-y-4" @submit.prevent="changePassword">
        <div v-if="pwError" class="flex items-center gap-2.5 rounded-xl border border-rose-200 bg-rose-50 p-3.5 text-sm text-rose-700">
          <svg class="w-4 h-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          {{ pwError }}
        </div>
        <div v-if="pwSuccess" class="flex items-center gap-2.5 rounded-xl border border-emerald-200 bg-emerald-50 p-3.5 text-sm text-emerald-700">
          <svg class="w-4 h-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          Password updated successfully.
        </div>
        <div>
          <label class="label">Current password</label>
          <input v-model="pwForm.currentPassword" type="password" class="input" placeholder="••••••••" />
        </div>
        <div>
          <label class="label">New password</label>
          <input v-model="pwForm.newPassword" type="password" class="input" placeholder="••••••••" />
        </div>
        <div>
          <label class="label">Confirm new password</label>
          <input v-model="pwForm.confirmNewPassword" type="password" class="input" placeholder="••••••••" />
        </div>
        <button type="submit" class="btn-primary">Update password</button>
      </form>
    </div>

    <!-- 2FA -->
    <div class="card">
      <!-- Skeleton while user loads -->
      <template v-if="!authStore.user">
        <div class="flex items-start justify-between mb-1">
          <div class="space-y-2">
            <div class="skeleton h-5 w-48" />
            <div class="skeleton h-4 w-64" />
          </div>
          <div class="skeleton h-6 w-16 rounded-full" />
        </div>
        <div class="mt-5">
          <div class="skeleton h-10 w-28" />
        </div>
      </template>

      <template v-else>
      <div class="flex items-start justify-between mb-1">
        <div>
          <h2 class="text-base font-semibold text-gray-900">Two-factor authentication</h2>
          <p class="text-sm text-gray-500 mt-0.5">Add an extra layer of security to your account</p>
        </div>
        <span :class="authStore.user.is2faEnabled ? 'badge-success' : 'badge-neutral'">
          {{ authStore.user.is2faEnabled ? 'Enabled' : 'Disabled' }}
        </span>
      </div>

      <div v-if="!authStore.user.is2faEnabled" class="mt-5">
        <NuxtLink to="/profile/security/2fa" class="btn-primary">Enable 2FA</NuxtLink>
      </div>

      <div v-else class="mt-5 space-y-5">
        <div>
          <h3 class="text-sm font-semibold text-gray-700 mb-2.5">Recovery codes</h3>
          <div v-if="recoveryCodes.length" class="bg-gray-50 rounded-xl border border-gray-100 p-4 font-mono text-xs grid grid-cols-2 gap-2 mb-3">
            <span v-for="c in recoveryCodes" :key="c" class="text-gray-700">{{ c }}</span>
          </div>
          <div class="flex flex-wrap gap-2">
            <button class="btn-secondary text-sm" @click="showRegenModal = true">Regenerate codes</button>
            <button v-if="recoveryCodes.length" class="btn-secondary text-sm" @click="copyRecoveryCodes">Copy codes</button>
          </div>
        </div>
        <div class="border-t border-gray-100 pt-5">
          <button class="btn-danger text-sm" @click="showDisableModal = true">Disable 2FA</button>
        </div>
      </div>
      </template>
    </div>

    <!-- Disable 2FA modal -->
    <div v-if="showDisableModal" class="modal-backdrop">
      <div class="modal">
        <h3 class="text-lg font-bold text-gray-900 mb-1">Disable 2FA</h3>
        <p class="text-sm text-gray-500 mb-5">Confirm your identity to disable two-factor authentication.</p>
        <div class="space-y-4">
          <div v-if="disableError" class="flex items-center gap-2.5 rounded-xl border border-rose-200 bg-rose-50 p-3.5 text-sm text-rose-700">
            <svg class="w-4 h-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            {{ disableError }}
          </div>
          <div>
            <label class="label">Password</label>
            <input v-model="disableForm.password" type="password" class="input" placeholder="••••••••" autofocus />
          </div>
          <div>
            <label class="label">Authenticator code</label>
            <input v-model="disableForm.code" type="text" inputmode="numeric" class="input font-mono tracking-widest text-center" maxlength="6" placeholder="000000" />
          </div>
          <div class="flex gap-3 pt-1">
            <button class="btn-secondary flex-1" @click="showDisableModal = false">Cancel</button>
            <button class="btn-danger flex-1" @click="disable2fa">Disable 2FA</button>
          </div>
        </div>
      </div>
    </div>

    <!-- Regen codes modal -->
    <div v-if="showRegenModal" class="modal-backdrop">
      <div class="modal">
        <h3 class="text-lg font-bold text-gray-900 mb-1">Regenerate recovery codes</h3>
        <p class="text-sm text-gray-500 mb-5">Old codes will be invalidated immediately.</p>
        <div class="space-y-4">
          <div v-if="regenError" class="flex items-center gap-2.5 rounded-xl border border-rose-200 bg-rose-50 p-3.5 text-sm text-rose-700">
            <svg class="w-4 h-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            {{ regenError }}
          </div>
          <div>
            <label class="label">Password</label>
            <input v-model="regenForm.password" type="password" class="input" placeholder="••••••••" autofocus />
          </div>
          <div>
            <label class="label">Authenticator code</label>
            <input v-model="regenForm.code" type="text" inputmode="numeric" class="input font-mono tracking-widest text-center" maxlength="6" placeholder="000000" />
          </div>
          <div class="flex gap-3 pt-1">
            <button class="btn-secondary flex-1" @click="showRegenModal = false">Cancel</button>
            <button class="btn-primary flex-1" @click="regenerateCodes">Regenerate</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
