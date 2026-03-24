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
</script>

<template>
  <div class="max-w-2xl space-y-8">
    <h1 class="text-2xl font-bold">Profile</h1>

    <!-- Avatar -->
    <section class="card">
      <h2 class="text-lg font-semibold mb-4">Avatar</h2>
      <div class="flex items-center gap-4">
        <img
          :src="avatarPreview ?? authStore.user?.avatarUrl ?? '/default-avatar.png'"
          class="w-20 h-20 rounded-full object-cover bg-gray-100"
        />
        <div>
          <button class="btn-secondary" @click="pickAvatar">Change photo</button>
          <p class="text-xs text-gray-400 mt-1">JPEG, PNG, WebP — max 2 MB</p>
          <p v-if="avatarError" class="text-red-500 text-xs mt-1">{{ avatarError }}</p>
          <input ref="avatarInput" type="file" accept="image/jpeg,image/png,image/webp" class="hidden" @change="onAvatarChange" />
        </div>
      </div>
    </section>

    <!-- Profile info -->
    <section class="card">
      <h2 class="text-lg font-semibold mb-4">Profile info</h2>
      <form class="space-y-4" @submit.prevent="saveProfile">
        <div v-if="profileError" class="text-red-600 text-sm">{{ profileError }}</div>
        <div v-if="profileSuccess" class="text-green-600 text-sm">Profile updated.</div>
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
          <input v-model="profileForm.displayName" type="text" class="input" />
        </div>
        <div>
          <label class="label">Bio</label>
          <textarea v-model="profileForm.bio" class="input" rows="3" maxlength="500" />
        </div>
        <button type="submit" class="btn-primary">Save changes</button>
      </form>
    </section>

    <!-- Email -->
    <section class="card">
      <h2 class="text-lg font-semibold mb-1">Email address</h2>
      <p class="text-gray-500 text-sm mb-4">{{ authStore.user?.email }}</p>
      <button class="btn-secondary" @click="showEmailModal = true">Change email</button>
    </section>

    <!-- Danger zone -->
    <section class="card border-red-200">
      <h2 class="text-lg font-semibold text-red-600 mb-1">Danger zone</h2>
      <p class="text-sm text-gray-500 mb-4">Permanently delete your account and all associated data.</p>
      <button class="btn-danger" @click="showDeleteModal = true">Delete account</button>
    </section>

    <!-- Email change modal -->
    <div v-if="showEmailModal" class="modal-backdrop">
      <div class="modal">
        <h3 class="text-lg font-semibold mb-4">Change email address</h3>
        <div v-if="emailChangeStatus === 'sent'" class="text-green-600 text-sm">
          Check your new email inbox to confirm the change.
        </div>
        <div v-else class="space-y-3">
          <input v-model="newEmail" type="email" class="input" placeholder="new@email.com" />
          <p v-if="emailChangeError" class="text-red-500 text-xs">{{ emailChangeError }}</p>
          <div class="flex gap-2 justify-end">
            <button class="btn-secondary" @click="showEmailModal = false">Cancel</button>
            <button class="btn-primary" @click="requestEmailChange">Send confirmation</button>
          </div>
        </div>
      </div>
    </div>

    <!-- Delete modal -->
    <div v-if="showDeleteModal" class="modal-backdrop">
      <div class="modal">
        <h3 class="text-lg font-semibold mb-2">Delete account?</h3>
        <p class="text-sm text-gray-500 mb-4">You have 7 days to cancel. After that, your account is permanently removed.</p>
        <div class="flex gap-2 justify-end">
          <button class="btn-secondary" @click="showDeleteModal = false">Cancel</button>
          <button class="btn-danger" @click="requestDeletion">Yes, delete my account</button>
        </div>
      </div>
    </div>
  </div>
</template>
