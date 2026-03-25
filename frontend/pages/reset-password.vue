<script setup lang="ts">
import { useForm } from 'vee-validate'
import { toTypedSchema } from '@vee-validate/zod'
import { z } from 'zod'
import type { ApiError } from '~/types/api'

definePageMeta({ layout: 'auth' })

const route = useRoute()
const api = useApi()
const token = computed(() => route.query.token as string)

const schema = toTypedSchema(z.object({
  password: z.string()
    .min(8, 'At least 8 characters')
    .regex(/[A-Z]/, 'Must contain an uppercase letter')
    .regex(/[0-9]/, 'Must contain a digit')
    .regex(/[^A-Za-z0-9]/, 'Must contain a special character'),
  confirmPassword: z.string(),
}).refine(d => d.password === d.confirmPassword, {
  message: 'Passwords do not match',
  path: ['confirmPassword'],
}))

const { handleSubmit, errors, defineField } = useForm({ validationSchema: schema })
const [password, passwordAttrs] = defineField('password')
const [confirmPassword, confirmPasswordAttrs] = defineField('confirmPassword')

const serverError = ref('')
const success = ref(false)

const onSubmit = handleSubmit(async (values) => {
  serverError.value = ''
  if (!token.value) {
    serverError.value = 'Invalid reset link.'
    return
  }
  try {
    await api.post('/auth/reset-password', { token: token.value, new_password: values.password })
    success.value = true
    setTimeout(() => navigateTo('/login'), 2000)
  }
  catch (err) {
    serverError.value = (err as ApiError).message ?? 'Reset failed'
  }
})
</script>

<template>
  <div>
    <div class="mb-8">
      <h1 class="text-2xl font-bold text-gray-900 dark:text-gray-100">Set new password</h1>
      <p class="mt-1.5 text-sm text-gray-500 dark:text-gray-400">Choose a strong password for your account.</p>
    </div>

    <div class="bg-white dark:bg-gray-800 rounded-2xl border border-gray-100 dark:border-gray-700 shadow-sm p-5 sm:p-8">
      <div v-if="success" class="text-center py-4 space-y-4">
        <div class="mx-auto w-14 h-14 rounded-full bg-emerald-100 flex items-center justify-center">
          <svg class="w-7 h-7 text-emerald-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.75" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
        </div>
        <div>
          <p class="font-semibold text-gray-900 dark:text-gray-100">Password updated!</p>
          <p class="text-sm text-gray-500 dark:text-gray-400 mt-1">Redirecting you to sign in…</p>
        </div>
      </div>

      <form v-else class="space-y-4" @submit.prevent="onSubmit">
        <div v-if="serverError" class="flex items-center gap-2.5 rounded-xl border border-rose-200 bg-rose-50 p-3.5 text-sm text-rose-700">
          <svg class="w-4 h-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          {{ serverError }}
        </div>

        <div>
          <label class="label">New password</label>
          <PasswordInput v-model="password" v-bind="passwordAttrs" :autofocus="true" />
          <p v-if="errors.password" class="mt-1.5 text-xs text-rose-600">{{ errors.password }}</p>
        </div>

        <div>
          <label class="label">Confirm password</label>
          <PasswordInput v-model="confirmPassword" v-bind="confirmPasswordAttrs" />
          <p v-if="errors.confirmPassword" class="mt-1.5 text-xs text-rose-600">{{ errors.confirmPassword }}</p>
        </div>

        <button type="submit" class="btn-primary w-full mt-2">Update password</button>
      </form>
    </div>
  </div>
</template>
