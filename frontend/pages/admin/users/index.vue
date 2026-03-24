<script setup lang="ts">
import type { User } from '~/types/user'
import type { ApiError } from '~/types/api'

definePageMeta({ layout: 'admin', middleware: 'superadmin' })

const api = useApi()

// ── Filters & pagination ──────────────────────────────────────────────────────
const search = ref('')
const role = ref('')
const status = ref('')
const page = ref(1)
const pageSize = 20
const total = ref(0)

const users = ref<User[]>([])
const loading = ref(false)
const error = ref('')

// Debounce search
let searchTimer: ReturnType<typeof setTimeout>
watch(search, () => {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(() => { page.value = 1; fetchUsers() }, 300)
})
watch([role, status], () => { page.value = 1; fetchUsers() })
watch(page, fetchUsers)

async function fetchUsers() {
  loading.value = true
  error.value = ''
  const params = new URLSearchParams({
    page: String(page.value),
    page_size: String(pageSize),
    ...(search.value && { search: search.value }),
    ...(role.value && { role: role.value }),
    ...(status.value && { status: status.value }),
  })
  try {
    const res = await api.get<User[]>(`/admin/users?${params}`)
    users.value = res.data
    total.value = res.meta?.total ?? 0
  }
  catch (err) {
    error.value = (err as ApiError).message ?? 'Failed to load users'
  }
  finally {
    loading.value = false
  }
}

onMounted(fetchUsers)

const totalPages = computed(() => Math.ceil(total.value / pageSize))

// ── Actions ───────────────────────────────────────────────────────────────────
async function setStatus(userId: string, action: 'activate' | 'deactivate' | 'ban' | 'unban') {
  if (!confirm(`${action} this user?`)) return
  try {
    await api.post(`/admin/users/${userId}/${action}`)
    fetchUsers()
  }
  catch (err) {
    alert((err as ApiError).message ?? 'Action failed')
  }
}

async function deleteUser(userId: string) {
  if (!confirm('Permanently delete this user?')) return
  try {
    await api.delete(`/admin/users/${userId}`)
    fetchUsers()
  }
  catch (err) {
    alert((err as ApiError).message ?? 'Delete failed')
  }
}

function statusBadge(s: string) {
  return s === 'active' ? 'badge-success' : s === 'banned' ? 'badge-danger' : 'badge-neutral'
}
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold">Users</h1>
      <NuxtLink to="/admin/users/create" class="btn-primary text-sm">+ New user</NuxtLink>
    </div>

    <!-- Filters -->
    <div class="flex flex-wrap gap-3 mb-4">
      <input v-model="search" type="text" class="input w-60" placeholder="Search name or email…" />
      <select v-model="role" class="input w-36">
        <option value="">All roles</option>
        <option value="user">User</option>
        <option value="superadmin">Superadmin</option>
      </select>
      <select v-model="status" class="input w-36">
        <option value="">All statuses</option>
        <option value="active">Active</option>
        <option value="inactive">Inactive</option>
        <option value="banned">Banned</option>
      </select>
    </div>

    <div v-if="error" class="text-red-600 text-sm mb-4">{{ error }}</div>

    <!-- Table -->
    <div class="overflow-x-auto rounded-lg border">
      <table class="w-full text-sm">
        <thead class="bg-gray-50 text-left">
          <tr>
            <th class="px-4 py-3 font-medium">Name</th>
            <th class="px-4 py-3 font-medium">Email</th>
            <th class="px-4 py-3 font-medium">Role</th>
            <th class="px-4 py-3 font-medium">Status</th>
            <th class="px-4 py-3 font-medium">Verified</th>
            <th class="px-4 py-3 font-medium">Actions</th>
          </tr>
        </thead>
        <tbody class="divide-y">
          <tr v-if="loading">
            <td colspan="6" class="px-4 py-8 text-center text-gray-400">Loading…</td>
          </tr>
          <tr v-else-if="users.length === 0">
            <td colspan="6" class="px-4 py-8 text-center text-gray-400">No users found</td>
          </tr>
          <tr v-for="u in users" :key="u.id" class="hover:bg-gray-50">
            <td class="px-4 py-3">{{ u.displayName || `${u.firstName} ${u.lastName}` }}</td>
            <td class="px-4 py-3">{{ u.email }}</td>
            <td class="px-4 py-3">
              <span class="badge-neutral">{{ u.role }}</span>
            </td>
            <td class="px-4 py-3">
              <span :class="statusBadge(u.status)">{{ u.status }}</span>
            </td>
            <td class="px-4 py-3">{{ u.isEmailVerified ? '✓' : '—' }}</td>
            <td class="px-4 py-3">
              <div class="flex gap-2">
                <NuxtLink :to="`/admin/users/${u.id}`" class="text-blue-600 hover:underline text-xs">Edit</NuxtLink>
                <button
                  v-if="u.status !== 'active'"
                  class="text-green-600 hover:underline text-xs"
                  @click="setStatus(u.id, 'activate')"
                >Activate</button>
                <button
                  v-if="u.status === 'active'"
                  class="text-yellow-600 hover:underline text-xs"
                  @click="setStatus(u.id, 'deactivate')"
                >Deactivate</button>
                <button
                  v-if="u.status !== 'banned'"
                  class="text-red-600 hover:underline text-xs"
                  @click="setStatus(u.id, 'ban')"
                >Ban</button>
                <button
                  v-if="u.status === 'banned'"
                  class="text-gray-600 hover:underline text-xs"
                  @click="setStatus(u.id, 'unban')"
                >Unban</button>
                <button class="text-red-700 hover:underline text-xs" @click="deleteUser(u.id)">Delete</button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Pagination -->
    <div v-if="totalPages > 1" class="flex items-center justify-between mt-4 text-sm">
      <p class="text-gray-500">{{ total }} users total</p>
      <div class="flex gap-2">
        <button class="btn-secondary" :disabled="page <= 1" @click="page--">← Prev</button>
        <span class="px-3 py-1">{{ page }} / {{ totalPages }}</span>
        <button class="btn-secondary" :disabled="page >= totalPages" @click="page++">Next →</button>
      </div>
    </div>
  </div>
</template>
