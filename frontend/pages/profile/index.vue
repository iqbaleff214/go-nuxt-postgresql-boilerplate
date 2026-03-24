<script setup lang="ts">
import type { ApiError } from '~/types/api'

definePageMeta({ layout: 'default', middleware: 'auth' })

const api = useApi()
const authStore = useAuthStore()

// ── Profile form ─────────────────────────────────────────────────────────────
const profileForm = reactive({
  firstName: '',
  lastName: '',
  displayName: '',
  bio: '',
})
const profileError = ref('')
const profileSuccess = ref(false)

watch(() => authStore.user, (u) => {
  if (!u) return
  profileForm.firstName = u.firstName
  profileForm.lastName = u.lastName
  profileForm.displayName = u.displayName
  profileForm.bio = u.bio ?? ''
}, { immediate: true })

async function saveProfile() {
  profileError.value = ''
  profileSuccess.value = false
  try {
    const res = await api.patch('/profile', {
      first_name: profileForm.firstName,
      last_name: profileForm.lastName,
      display_name: profileForm.displayName,
      bio: profileForm.bio,
    })
    authStore.user = res.data as any
    profileSuccess.value = true
  }
  catch (err) {
    profileError.value = (err as ApiError).message ?? 'Update failed'
  }
}

// ── Avatar ────────────────────────────────────────────────────────────────────
const avatarInput = ref<HTMLInputElement>()
const avatarPreview = ref<string>()
const avatarError = ref('')

function pickAvatar() {
  avatarInput.value?.click()
}

async function onAvatarChange(e: Event) {
  const file = (e.target as HTMLInputElement).files?.[0]
  if (!file) return
  avatarError.value = ''
  if (file.size > 2 * 1024 * 1024) {
    avatarError.value = 'File must be 2 MB or smaller'
    return
  }
  avatarPreview.value = URL.createObjectURL(file)
  const fd = new FormData()
  fd.append('avatar', file)
  try {
    const res = await api.upload<{ avatar_url: string }>('/profile/avatar', fd)
    if (authStore.user) authStore.user.avatarUrl = res.data.avatar_url
  }
  catch (err) {
    avatarError.value = (err as ApiError).message ?? 'Upload failed'
  }
}

// ── Email change ──────────────────────────────────────────────────────────────
const showEmailModal = ref(false)
const newEmail = ref('')
const emailChangeStatus = ref<'' | 'sent' | 'error'>('')
const emailChangeError = ref('')

async function requestEmailChange() {
  emailChangeError.value = ''
  try {
    await api.post('/profile/email', { new_email: newEmail.value })
    emailChangeStatus.value = 'sent'
  }
  catch (err) {
    emailChangeError.value = (err as ApiError).message ?? 'Failed'
  }
}

// ── Account deletion ──────────────────────────────────────────────────────────
const showDeleteModal = ref(false)

async function requestDeletion() {
  try {
    await api.post('/profile/delete')
    await authStore.logout()
  }
  catch (err) {
    alert((err as ApiError).message ?? 'Failed to request deletion')
  }
}

const initials = computed(() => {
  const u = authStore.user
  if (!u) return '?'
  return ((u.firstName?.[0] ?? '') + (u.lastName?.[0] ?? '')).toUpperCase() || u.email[0].toUpperCase()
})
</script>

<template>
  <div class="max-w-2xl space-y-6">
    <div>
      <h1 class="text-2xl font-bold text-gray-900">Profile</h1>
      <p class="text-sm text-gray-500 mt-1">Manage your personal information and preferences</p>
    </div>

    <!-- Skeleton -->
    <template v-if="!authStore.user">
      <!-- Avatar skeleton -->
      <div class="card flex items-center gap-5">
        <div class="skeleton w-20 h-20 rounded-2xl shrink-0" />
        <div class="space-y-2 flex-1">
          <div class="skeleton h-4 w-36" />
          <div class="skeleton h-3 w-48" />
          <div class="skeleton h-8 w-28 mt-1" />
        </div>
      </div>
      <!-- Form skeleton -->
      <div class="card space-y-5">
        <div class="skeleton h-5 w-40" />
        <div class="grid grid-cols-2 gap-4">
          <div class="space-y-1.5"><div class="skeleton h-3 w-20" /><div class="skeleton h-10" /></div>
          <div class="space-y-1.5"><div class="skeleton h-3 w-20" /><div class="skeleton h-10" /></div>
        </div>
        <div class="space-y-1.5"><div class="skeleton h-3 w-28" /><div class="skeleton h-10" /></div>
        <div class="space-y-1.5"><div class="skeleton h-3 w-12" /><div class="skeleton h-24" /></div>
        <div class="skeleton h-10 w-28" />
      </div>
      <!-- Email skeleton -->
      <div class="card flex items-center justify-between">
        <div class="space-y-1.5">
          <div class="skeleton h-4 w-32" />
          <div class="skeleton h-3 w-44" />
        </div>
        <div class="skeleton h-9 w-20" />
      </div>
    </template>

    <!-- Real content -->
    <template v-else>
      <!-- Avatar -->
      <div class="card">
        <div class="flex items-center gap-5">
          <div class="relative shrink-0">
            <div v-if="avatarPreview || authStore.user.avatarUrl">
              <img
                :src="avatarPreview ?? authStore.user.avatarUrl"
                class="w-20 h-20 rounded-2xl object-cover"
              />
            </div>
            <div v-else class="w-20 h-20 rounded-2xl bg-emerald-100 text-emerald-700 flex items-center justify-center text-2xl font-bold">
              {{ initials }}
            </div>
          </div>
          <div class="flex-1 min-w-0">
            <p class="font-semibold text-gray-900">
              {{ authStore.user.displayName || `${authStore.user.firstName} ${authStore.user.lastName}` }}
            </p>
            <p class="text-sm text-gray-500 mt-0.5 truncate">{{ authStore.user.email }}</p>
            <div class="flex items-center gap-3 mt-3">
              <button class="btn-secondary text-sm py-1.5 px-3" @click="pickAvatar">Change photo</button>
              <span class="text-xs text-gray-400">JPEG, PNG, WebP · max 2 MB</span>
            </div>
            <p v-if="avatarError" class="text-xs text-rose-600 mt-1.5">{{ avatarError }}</p>
            <input ref="avatarInput" type="file" accept="image/jpeg,image/png,image/webp" class="hidden" @change="onAvatarChange" />
          </div>
        </div>
      </div>

      <!-- Profile info -->
      <div class="card">
        <h2 class="text-base font-semibold text-gray-900 mb-5">Personal information</h2>
        <form class="space-y-4" @submit.prevent="saveProfile">
          <div v-if="profileError" class="flex items-center gap-2.5 rounded-xl border border-rose-200 bg-rose-50 p-3.5 text-sm text-rose-700">
            <svg class="w-4 h-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            {{ profileError }}
          </div>
          <div v-if="profileSuccess" class="flex items-center gap-2.5 rounded-xl border border-emerald-200 bg-emerald-50 p-3.5 text-sm text-emerald-700">
            <svg class="w-4 h-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            Profile updated successfully.
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="label">First name</label>
              <input v-model="profileForm.firstName" type="text" class="input" />
            </div>
            <div>
              <label class="label">Last name</label>
              <input v-model="profileForm.lastName" type="text" class="input" />
            </div>
          </div>
          <div>
            <label class="label">Display name</label>
            <input v-model="profileForm.displayName" type="text" class="input" placeholder="How you'll appear to others" />
          </div>
          <div>
            <label class="label">Bio</label>
            <textarea v-model="profileForm.bio" class="input resize-none" rows="3" maxlength="500" placeholder="Tell us a bit about yourself…" />
          </div>
          <div class="flex items-center justify-between pt-1">
            <button type="submit" class="btn-primary">Save changes</button>
            <NuxtLink to="/profile/security" class="text-sm font-medium text-gray-500 hover:text-gray-700 transition-colors">
              Security settings →
            </NuxtLink>
          </div>
        </form>
      </div>

      <!-- Email -->
      <div class="card">
        <div class="flex items-center justify-between">
          <div>
            <h2 class="text-base font-semibold text-gray-900">Email address</h2>
            <p class="text-sm text-gray-500 mt-0.5">{{ authStore.user.email }}</p>
          </div>
          <button class="btn-secondary text-sm" @click="showEmailModal = true">Change</button>
        </div>
      </div>

      <!-- Danger zone -->
      <div class="rounded-2xl border border-rose-200 bg-white p-6">
        <h2 class="text-base font-semibold text-rose-600 mb-1">Danger zone</h2>
        <p class="text-sm text-gray-500 mb-4">Permanently delete your account and all associated data. This cannot be undone.</p>
        <button class="btn-danger text-sm" @click="showDeleteModal = true">Delete account</button>
      </div>
    </template>

    <!-- Email change modal -->
    <div v-if="showEmailModal" class="modal-backdrop">
      <div class="modal">
        <h3 class="text-lg font-bold text-gray-900 mb-1">Change email address</h3>
        <p class="text-sm text-gray-500 mb-5">We'll send a confirmation link to your new email address.</p>
        <div v-if="emailChangeStatus === 'sent'" class="flex items-center gap-3 rounded-xl border border-emerald-200 bg-emerald-50 p-4 text-sm text-emerald-700 mb-4">
          <svg class="w-5 h-5 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          Check your new inbox to confirm the change.
        </div>
        <div v-else class="space-y-4">
          <div>
            <label class="label">New email address</label>
            <input v-model="newEmail" type="email" class="input" placeholder="new@example.com" autofocus />
            <p v-if="emailChangeError" class="mt-1.5 text-xs text-rose-600">{{ emailChangeError }}</p>
          </div>
          <div class="flex gap-3 justify-end">
            <button class="btn-secondary" @click="showEmailModal = false; emailChangeStatus = ''">Cancel</button>
            <button class="btn-primary" @click="requestEmailChange">Send confirmation</button>
          </div>
        </div>
        <button v-if="emailChangeStatus === 'sent'" class="btn-secondary w-full mt-2" @click="showEmailModal = false; emailChangeStatus = ''">Close</button>
      </div>
    </div>

    <!-- Delete modal -->
    <div v-if="showDeleteModal" class="modal-backdrop">
      <div class="modal">
        <div class="mx-auto w-12 h-12 rounded-full bg-rose-100 flex items-center justify-center mb-4">
          <svg class="w-6 h-6 text-rose-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
          </svg>
        </div>
        <h3 class="text-lg font-bold text-gray-900 text-center mb-1">Delete account?</h3>
        <p class="text-sm text-gray-500 text-center mb-6">You have 30 days to cancel. After that, your account and all data are permanently removed.</p>
        <div class="flex gap-3">
          <button class="btn-secondary flex-1" @click="showDeleteModal = false">Cancel</button>
          <button class="btn-danger flex-1" @click="requestDeletion">Yes, delete my account</button>
        </div>
      </div>
    </div>
  </div>
</template>
