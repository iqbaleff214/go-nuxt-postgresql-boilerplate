<script setup lang="ts">
import { useForm } from 'vee-validate'
import { toTypedSchema } from '@vee-validate/zod'
import { z } from 'zod'
import type { ApiError } from '~/types/api'

definePageMeta({ layout: 'auth', middleware: 'guest' })

const schema = toTypedSchema(z.object({
  email: z.string().email('Invalid email'),
  password: z.string().min(1, 'Password is required'),
}))

const { handleSubmit, errors, defineField } = useForm({ validationSchema: schema })
const [email, emailAttrs] = defineField('email')
const [password, passwordAttrs] = defineField('password')

const authStore = useAuthStore()
const serverError = ref('')

const onSubmit = handleSubmit(async (values) => {
  serverError.value = ''
  try {
    const result = await authStore.login(values.email, values.password)
    if (result.requires2fa) {
      await navigateTo('/login/2fa')
    }
    else {
      await navigateTo('/dashboard')
    }
  }
  catch (err) {
    const e = err as ApiError
    serverError.value = e.message ?? 'Login failed'
  }
})
</script>

<template>
  <div>
    <div class="mb-8">
      <h1 class="text-2xl font-bold text-gray-900">Welcome back</h1>
      <p class="mt-1.5 text-sm text-gray-500">Sign in to your account to continue</p>
    </div>

    <div class="bg-white rounded-2xl border border-gray-100 shadow-sm p-8 space-y-5">
      <div v-if="serverError" class="flex items-center gap-2.5 rounded-xl border border-rose-200 bg-rose-50 p-3.5 text-sm text-rose-700">
        <svg class="w-4 h-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        {{ serverError }}
      </div>

      <form class="space-y-4" @submit.prevent="onSubmit">
        <div>
          <label class="label">Email address</label>
          <input v-model="email" v-bind="emailAttrs" type="email" class="input" placeholder="you@example.com" autofocus />
          <p v-if="errors.email" class="mt-1.5 text-xs text-rose-600">{{ errors.email }}</p>
        </div>

        <div>
          <div class="flex items-center justify-between mb-1.5">
            <label class="label !mb-0">Password</label>
            <NuxtLink to="/forgot-password" class="text-xs font-medium text-emerald-600 hover:text-emerald-700 transition-colors">
              Forgot password?
            </NuxtLink>
          </div>
          <input v-model="password" v-bind="passwordAttrs" type="password" class="input" placeholder="••••••••" />
          <p v-if="errors.password" class="mt-1.5 text-xs text-rose-600">{{ errors.password }}</p>
        </div>

        <button type="submit" class="btn-primary w-full mt-2">
          Sign in
        </button>
      </form>
    </div>

    <p class="mt-6 text-center text-sm text-gray-500">
      Don't have an account?
      <NuxtLink to="/register" class="font-semibold text-emerald-600 hover:text-emerald-700 transition-colors">Create one</NuxtLink>
    </p>
  </div>
</template>
