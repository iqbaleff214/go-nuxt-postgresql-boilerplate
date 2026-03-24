<script setup lang="ts">
import type { ApiError } from '~/types/api'

definePageMeta({ layout: 'admin', middleware: 'superadmin' })

const api = useApi()
const form = reactive({ firstName: '', lastName: '', email: '', role: 'user' })
const error = ref('')
const fieldErrors = ref<Record<string, string>>({})

async function submit() {
  error.value = ''
  fieldErrors.value = {}
  try {
    await api.post('/admin/users', {
      first_name: form.firstName,
      last_name: form.lastName,
      email: form.email,
      role: form.role,
    })
    await navigateTo('/admin/users')
  }
  catch (err) {
    const e = err as ApiError
    if (e.errors?.length) {
      e.errors.forEach(fe => { fieldErrors.value[fe.field] = fe.detail })
    }
    else {
      error.value = e.message ?? 'Failed to create user'
    }
  }
}
</script>

<template>
  <div class="max-w-lg">
    <div class="flex items-center gap-3 mb-6">
      <NuxtLink to="/admin/users" class="text-gray-500 hover:underline text-sm">← Users</NuxtLink>
      <h1 class="text-2xl font-bold">New user</h1>
    </div>

    <form class="card space-y-4" @submit.prevent="submit">
      <div v-if="error" class="p-3 bg-red-50 border border-red-200 rounded text-red-700 text-sm">{{ error }}</div>

      <div class="grid grid-cols-2 gap-4">
        <div>
          <label class="label">First name</label>
          <input v-model="form.firstName" type="text" class="input" required />
          <p v-if="fieldErrors.first_name" class="text-red-500 text-xs mt-1">{{ fieldErrors.first_name }}</p>
        </div>
        <div>
          <label class="label">Last name</label>
          <input v-model="form.lastName" type="text" class="input" required />
          <p v-if="fieldErrors.last_name" class="text-red-500 text-xs mt-1">{{ fieldErrors.last_name }}</p>
        </div>
      </div>

      <div>
        <label class="label">Email</label>
        <input v-model="form.email" type="email" class="input" required />
        <p v-if="fieldErrors.email" class="text-red-500 text-xs mt-1">{{ fieldErrors.email }}</p>
      </div>

      <div>
        <label class="label">Role</label>
        <select v-model="form.role" class="input">
          <option value="user">User</option>
          <option value="superadmin">Superadmin</option>
        </select>
      </div>

      <p class="text-xs text-gray-400">
        A temporary password will be generated and emailed to the user. They must change it on first login.
      </p>

      <div class="flex gap-2 pt-2">
        <button type="submit" class="btn-primary">Create user</button>
        <NuxtLink to="/admin/users" class="btn-secondary">Cancel</NuxtLink>
      </div>
    </form>
  </div>
</template>
