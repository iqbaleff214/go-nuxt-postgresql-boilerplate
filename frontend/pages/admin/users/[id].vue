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
    <div class="flex items-center gap-3 mb-6">
      <NuxtLink to="/admin/users" class="text-gray-500 hover:underline text-sm">← Users</NuxtLink>
      <h1 class="text-2xl font-bold">Edit user</h1>
    </div>

    <div v-if="loading" class="text-gray-400">Loading…</div>
    <div v-else-if="error" class="text-red-600">{{ error }}</div>

    <form v-else class="card space-y-4" @submit.prevent="save">
      <div v-if="saveError" class="text-red-600 text-sm">{{ saveError }}</div>
      <div v-if="saveSuccess" class="text-green-600 text-sm">Saved.</div>

      <div class="text-sm text-gray-400">ID: {{ user?.id }}</div>

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
        <textarea v-model="form.bio" class="input" rows="3" />
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

      <div class="flex gap-2 pt-2">
        <button type="submit" class="btn-primary">Save changes</button>
        <NuxtLink to="/admin/users" class="btn-secondary">Cancel</NuxtLink>
      </div>
    </form>
  </div>
</template>
