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
    <h1 class="text-2xl font-bold mb-6">Set new password</h1>

    <div v-if="success" class="p-4 bg-green-50 border border-green-200 rounded-lg text-green-800">
      Password updated! Redirecting to login…
    </div>

    <form v-else class="space-y-4" @submit.prevent="onSubmit">
      <div v-if="serverError" class="p-3 bg-red-50 border border-red-200 rounded text-red-700 text-sm">
        {{ serverError }}
      </div>

      <div>
        <label class="block text-sm font-medium mb-1">New password</label>
        <input v-model="password" v-bind="passwordAttrs" type="password" class="input" />
        <p v-if="errors.password" class="text-red-500 text-xs mt-1">{{ errors.password }}</p>
      </div>

      <div>
        <label class="block text-sm font-medium mb-1">Confirm password</label>
        <input v-model="confirmPassword" v-bind="confirmPasswordAttrs" type="password" class="input" />
        <p v-if="errors.confirmPassword" class="text-red-500 text-xs mt-1">{{ errors.confirmPassword }}</p>
      </div>

      <button type="submit" class="btn-primary w-full">Update password</button>
    </form>
  </div>
</template>
