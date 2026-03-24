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
    <h1 class="text-2xl font-bold mb-6">Sign in</h1>

    <form class="space-y-4" @submit.prevent="onSubmit">
      <div v-if="serverError" class="p-3 bg-red-50 border border-red-200 rounded text-red-700 text-sm">
        {{ serverError }}
      </div>

      <div>
        <label class="block text-sm font-medium mb-1">Email</label>
        <input v-model="email" v-bind="emailAttrs" type="email" class="input" placeholder="jane@example.com" />
        <p v-if="errors.email" class="text-red-500 text-xs mt-1">{{ errors.email }}</p>
      </div>

      <div>
        <label class="block text-sm font-medium mb-1">Password</label>
        <input v-model="password" v-bind="passwordAttrs" type="password" class="input" />
        <p v-if="errors.password" class="text-red-500 text-xs mt-1">{{ errors.password }}</p>
        <div class="text-right mt-1">
          <NuxtLink to="/forgot-password" class="text-xs text-gray-500 hover:underline">Forgot password?</NuxtLink>
        </div>
      </div>

      <button type="submit" class="btn-primary w-full">Sign in</button>

      <p class="text-center text-sm text-gray-600">
        No account? <NuxtLink to="/register" class="text-primary underline">Create one</NuxtLink>
      </p>
    </form>
  </div>
</template>
