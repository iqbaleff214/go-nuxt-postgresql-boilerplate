<script setup lang="ts">
import { useForm } from 'vee-validate'
import { toTypedSchema } from '@vee-validate/zod'
import { z } from 'zod'
import type { ApiError } from '~/types/api'

definePageMeta({ layout: 'auth', middleware: 'guest' })

const schema = toTypedSchema(z.object({
  firstName: z.string().min(1, 'First name is required'),
  lastName: z.string().min(1, 'Last name is required'),
  email: z.string().email('Invalid email'),
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

const { handleSubmit, errors, setFieldError, defineField } = useForm({ validationSchema: schema })

const [firstName, firstNameAttrs] = defineField('firstName')
const [lastName, lastNameAttrs] = defineField('lastName')
const [email, emailAttrs] = defineField('email')
const [password, passwordAttrs] = defineField('password')
const [confirmPassword, confirmPasswordAttrs] = defineField('confirmPassword')

const api = useApi()
const success = ref(false)
const serverError = ref('')

const onSubmit = handleSubmit(async (values) => {
  serverError.value = ''
  try {
    await api.post('/auth/register', {
      first_name: values.firstName,
      last_name: values.lastName,
      email: values.email,
      password: values.password,
    })
    success.value = true
  }
  catch (err) {
    const e = err as ApiError
    if (e.errors?.length) {
      e.errors.forEach(fe => setFieldError(fe.field as any, fe.detail))
    }
    else {
      serverError.value = e.message ?? 'Registration failed'
    }
  }
})
</script>

<template>
  <div>
    <h1 class="text-2xl font-bold mb-6">Create account</h1>

    <div v-if="success" class="p-4 bg-green-50 border border-green-200 rounded-lg text-green-800">
      Check your email to verify your account before logging in.
    </div>

    <form v-else class="space-y-4" @submit.prevent="onSubmit">
      <div v-if="serverError" class="p-3 bg-red-50 border border-red-200 rounded text-red-700 text-sm">
        {{ serverError }}
      </div>

      <div class="grid grid-cols-2 gap-4">
        <div>
          <label class="block text-sm font-medium mb-1">First name</label>
          <input v-model="firstName" v-bind="firstNameAttrs" type="text" class="input" placeholder="Jane" />
          <p v-if="errors.firstName" class="text-red-500 text-xs mt-1">{{ errors.firstName }}</p>
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">Last name</label>
          <input v-model="lastName" v-bind="lastNameAttrs" type="text" class="input" placeholder="Doe" />
          <p v-if="errors.lastName" class="text-red-500 text-xs mt-1">{{ errors.lastName }}</p>
        </div>
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
      </div>

      <div>
        <label class="block text-sm font-medium mb-1">Confirm password</label>
        <input v-model="confirmPassword" v-bind="confirmPasswordAttrs" type="password" class="input" />
        <p v-if="errors.confirmPassword" class="text-red-500 text-xs mt-1">{{ errors.confirmPassword }}</p>
      </div>

      <button type="submit" class="btn-primary w-full">Create account</button>

      <p class="text-center text-sm text-gray-600">
        Already have an account? <NuxtLink to="/login" class="text-primary underline">Sign in</NuxtLink>
      </p>
    </form>
  </div>
</template>
