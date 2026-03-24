<script setup lang="ts">
import type { User } from '~/types/user'
import type { ApiError } from '~/types/api'

definePageMeta({ layout: 'admin', middleware: 'superadmin' })

const route = useRoute()
const api = useApi()
const userId = computed(() => route.params.id as string)

const user = ref<User | null>(null)
const loading = ref(true)
const error = ref('')
const saveSuccess = ref(false)
const saveError = ref('')

const form = reactive({
  firstName: '',
  lastName: '',
  displayName: '',
  bio: '',
  role: '',
  status: '',
})

onMounted(async () => {
  try {
    const res = await api.get<User>(`/admin/users/${userId.value}`)
    user.value = res.data
    Object.assign(form, {
      firstName: res.data.firstName,
      lastName: res.data.lastName,
      displayName: res.data.displayName,
      bio: res.data.bio ?? '',
      role: res.data.role,
      status: res.data.status,
    })
  }
  catch (err) {
    error.value = (err as ApiError).message ?? 'User not found'
  }
  finally {
    loading.value = false
  }
})

async function save() {
  saveError.value = ''
  saveSuccess.value = false
  try {
    const res = await api.patch<User>(`/admin/users/${userId.value}`, {
      first_name: form.firstName,
      last_name: form.lastName,
      display_name: form.displayName,
      bio: form.bio,
      role: form.role,
      status: form.status,
    })
    user.value = res.data
    saveSuccess.value = true
  }
  catch (err) {
    saveError.value = (err as ApiError).message ?? 'Save failed'
  }
}
</script>

<template>
  <div class="max-w-2xl">
    <!-- Breadcrumb -->
    <div class="flex items-center gap-2 text-sm mb-6">
      <NuxtLink to="/admin/users" class="text-gray-400 hover:text-gray-600 transition-colors font-medium">Users</NuxtLink>
      <svg class="w-3.5 h-3.5 text-gray-300" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
      </svg>
      <span class="text-gray-700 font-semibold">Edit user</span>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="card flex items-center justify-center py-16">
      <svg class="w-6 h-6 text-gray-300 animate-spin" fill="none" viewBox="0 0 24 24">
        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
      </svg>
    </div>

    <!-- Error -->
    <div v-else-if="error" class="flex items-center gap-2.5 rounded-xl border border-rose-200 bg-rose-50 p-4 text-sm text-rose-700">
      <svg class="w-4 h-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
      </svg>
      {{ error }}
    </div>

    <!-- Form -->
    <form v-else class="card space-y-5" @submit.prevent="save">
      <div class="flex items-center justify-between pb-4 border-b border-gray-100">
        <div>
          <h1 class="text-lg font-bold text-gray-900">{{ user?.displayName || `${user?.firstName} ${user?.lastName}` }}</h1>
          <p class="text-xs text-gray-400 mt-0.5 font-mono">{{ user?.id }}</p>
        </div>
      </div>

      <div v-if="saveError" class="flex items-center gap-2.5 rounded-xl border border-rose-200 bg-rose-50 p-3.5 text-sm text-rose-700">
        <svg class="w-4 h-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        {{ saveError }}
      </div>
      <div v-if="saveSuccess" class="flex items-center gap-2.5 rounded-xl border border-emerald-200 bg-emerald-50 p-3.5 text-sm text-emerald-700">
        <svg class="w-4 h-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        Changes saved successfully.
      </div>

      <div class="grid grid-cols-2 gap-4">
        <div>
          <label class="label">First name</label>
          <input v-model="form.firstName" type="text" class="input" />
        </div>
        <div>
          <label class="label">Last name</label>
          <input v-model="form.lastName" type="text" class="input" />
        </div>
      </div>

      <div>
        <label class="label">Display name</label>
        <input v-model="form.displayName" type="text" class="input" />
      </div>

      <div>
        <label class="label">Bio</label>
        <textarea v-model="form.bio" class="input resize-none" rows="3" />
      </div>

      <div class="grid grid-cols-2 gap-4">
        <div>
          <label class="label">Role</label>
          <select v-model="form.role" class="input">
            <option value="user">User</option>
            <option value="superadmin">Superadmin</option>
          </select>
        </div>
        <div>
          <label class="label">Status</label>
          <select v-model="form.status" class="input">
            <option value="active">Active</option>
            <option value="inactive">Inactive</option>
            <option value="banned">Banned</option>
          </select>
        </div>
      </div>

      <div class="flex gap-3 pt-2 border-t border-gray-100">
        <button type="submit" class="btn-primary">Save changes</button>
        <NuxtLink to="/admin/users" class="btn-secondary">Cancel</NuxtLink>
      </div>
    </form>
  </div>
</template>
