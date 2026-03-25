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
    <!-- Breadcrumb -->
    <div class="flex items-center gap-2 text-sm mb-6">
      <NuxtLink to="/admin/users" class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 transition-colors font-medium">Users</NuxtLink>
      <svg class="w-3.5 h-3.5 text-gray-300" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
      </svg>
      <span class="text-gray-700 font-semibold">New user</span>
    </div>

    <form class="card space-y-5" @submit.prevent="submit">
      <div class="pb-4 border-b border-gray-100">
        <h1 class="text-lg font-bold text-gray-900 dark:text-gray-100">Create user</h1>
        <p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">A temporary password will be emailed to the new user.</p>
      </div>

      <div v-if="error" class="flex items-center gap-2.5 rounded-xl border border-rose-200 bg-rose-50 p-3.5 text-sm text-rose-700">
        <svg class="w-4 h-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        {{ error }}
      </div>

      <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <div>
          <label class="label">First name</label>
          <input v-model="form.firstName" type="text" class="input" required autofocus />
          <p v-if="fieldErrors.first_name" class="mt-1.5 text-xs text-rose-600">{{ fieldErrors.first_name }}</p>
        </div>
        <div>
          <label class="label">Last name</label>
          <input v-model="form.lastName" type="text" class="input" required />
          <p v-if="fieldErrors.last_name" class="mt-1.5 text-xs text-rose-600">{{ fieldErrors.last_name }}</p>
        </div>
      </div>

      <div>
        <label class="label">Email address</label>
        <input v-model="form.email" type="email" class="input" required />
        <p v-if="fieldErrors.email" class="mt-1.5 text-xs text-rose-600">{{ fieldErrors.email }}</p>
      </div>

      <div>
        <label class="label">Role</label>
        <select v-model="form.role" class="input">
          <option value="user">User</option>
          <option value="superadmin">Superadmin</option>
        </select>
      </div>

      <div class="flex gap-3 pt-2 border-t border-gray-100">
        <button type="submit" class="btn-primary">Create user</button>
        <NuxtLink to="/admin/users" class="btn-secondary">Cancel</NuxtLink>
      </div>
    </form>
  </div>
</template>
