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
    <!-- Header -->
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-gray-100">Users</h1>
        <p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">{{ total }} total users</p>
      </div>
      <NuxtLink to="/admin/users/create" class="btn-primary">
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
        </svg>
        New user
      </NuxtLink>
    </div>

    <!-- Filters -->
    <div class="flex flex-wrap gap-3 mb-5">
      <div class="relative">
        <svg class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
        </svg>
        <input v-model="search" type="text" class="input pl-9 w-64" placeholder="Search name or email…" />
      </div>
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

    <div v-if="error" class="flex items-center gap-2.5 rounded-xl border border-rose-200 bg-rose-50 p-3.5 text-sm text-rose-700 mb-4">
      <svg class="w-4 h-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
      </svg>
      {{ error }}
    </div>

    <!-- Table -->
    <div class="bg-white dark:bg-gray-800 rounded-2xl border border-gray-100 dark:border-gray-700 shadow-sm overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr class="border-b border-gray-100 dark:border-gray-700 bg-gray-50/50 dark:bg-gray-700/30">
              <th class="text-left px-5 py-3.5 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide">Name</th>
              <th class="text-left px-5 py-3.5 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide">Email</th>
              <th class="text-left px-5 py-3.5 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide">Role</th>
              <th class="text-left px-5 py-3.5 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide">Status</th>
              <th class="text-left px-5 py-3.5 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide">Verified</th>
              <th class="text-left px-5 py-3.5 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide">Actions</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-50 dark:divide-gray-700">
            <template v-if="loading">
              <tr v-for="i in 8" :key="i" class="border-b border-gray-50 dark:border-gray-700">
                <td class="px-5 py-3.5"><div class="skeleton h-4 w-32" /></td>
                <td class="px-5 py-3.5"><div class="skeleton h-4 w-40" /></td>
                <td class="px-5 py-3.5"><div class="skeleton h-5 w-16 rounded-full" /></td>
                <td class="px-5 py-3.5"><div class="skeleton h-5 w-14 rounded-full" /></td>
                <td class="px-5 py-3.5"><div class="skeleton w-5 h-5 rounded-full" /></td>
                <td class="px-5 py-3.5"><div class="skeleton h-4 w-24" /></td>
              </tr>
            </template>
            <tr v-else-if="users.length === 0">
              <td colspan="6" class="px-5 py-12 text-center text-gray-400 dark:text-gray-500 text-sm">No users found</td>
            </tr>
            <tr v-for="u in users" :key="u.id" class="hover:bg-gray-50/50 dark:hover:bg-gray-700/30 transition-colors">
              <td class="px-5 py-3.5 font-medium text-gray-900 dark:text-gray-100">{{ u.display_name || `${u.first_name} ${u.last_name}` }}</td>
              <td class="px-5 py-3.5 text-gray-500 dark:text-gray-400">{{ u.email }}</td>
              <td class="px-5 py-3.5">
                <span class="badge-neutral capitalize">{{ u.role }}</span>
              </td>
              <td class="px-5 py-3.5">
                <span :class="statusBadge(u.status)" class="capitalize">{{ u.status }}</span>
              </td>
              <td class="px-5 py-3.5">
                <div v-if="u.is_email_verified" class="w-5 h-5 rounded-full bg-emerald-100 dark:bg-emerald-900/40 flex items-center justify-center">
                  <svg class="w-3 h-3 text-emerald-600 dark:text-emerald-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 13l4 4L19 7" />
                  </svg>
                </div>
                <div v-else class="w-5 h-5 rounded-full bg-gray-100 dark:bg-gray-700 flex items-center justify-center">
                  <svg class="w-3 h-3 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </div>
              </td>
              <td class="px-5 py-3.5">
                <div class="flex items-center gap-3">
                  <NuxtLink :to="`/admin/users/${u.id}`" class="text-xs font-semibold text-emerald-600 hover:text-emerald-700 transition-colors">Edit</NuxtLink>
                  <button
                    v-if="u.status !== 'active'"
                    class="text-xs font-semibold text-blue-600 hover:text-blue-700 transition-colors"
                    @click="setStatus(u.id, 'activate')"
                  >Activate</button>
                  <button
                    v-if="u.status === 'active'"
                    class="text-xs font-semibold text-amber-600 hover:text-amber-700 transition-colors"
                    @click="setStatus(u.id, 'deactivate')"
                  >Deactivate</button>
                  <button
                    v-if="u.status !== 'banned'"
                    class="text-xs font-semibold text-rose-500 hover:text-rose-600 transition-colors"
                    @click="setStatus(u.id, 'ban')"
                  >Ban</button>
                  <button
                    v-if="u.status === 'banned'"
                    class="text-xs font-semibold text-gray-500 hover:text-gray-700 transition-colors"
                    @click="setStatus(u.id, 'unban')"
                  >Unban</button>
                  <button class="text-xs font-semibold text-rose-600 hover:text-rose-700 transition-colors" @click="deleteUser(u.id)">Delete</button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Pagination -->
    <div v-if="totalPages > 1" class="flex items-center justify-between mt-4">
      <p class="text-sm text-gray-500 dark:text-gray-400">Showing page {{ page }} of {{ totalPages }}</p>
      <div class="flex items-center gap-2">
        <button class="btn-secondary py-1.5 px-3 text-xs" :disabled="page <= 1" @click="page--">← Previous</button>
        <span class="text-sm text-gray-600 dark:text-gray-400 font-medium px-2">{{ page }}</span>
        <button class="btn-secondary py-1.5 px-3 text-xs" :disabled="page >= totalPages" @click="page++">Next →</button>
      </div>
    </div>
  </div>
</template>
