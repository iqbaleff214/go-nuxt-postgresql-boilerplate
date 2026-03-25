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
    <div class="mb-8">
      <h1 class="text-2xl font-bold text-gray-900 dark:text-gray-100">Create your account</h1>
      <p class="mt-1.5 text-sm text-gray-500 dark:text-gray-400">Join today — it's free</p>
    </div>

    <div class="bg-white dark:bg-gray-800 rounded-2xl border border-gray-100 dark:border-gray-700 shadow-sm p-8">
      <div v-if="success" class="flex items-start gap-3 rounded-xl border border-emerald-200 bg-emerald-50 p-4 text-sm text-emerald-700">
        <svg class="w-5 h-5 shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <div>
          <p class="font-semibold">Check your inbox</p>
          <p class="mt-0.5 text-emerald-600">We sent a verification link to your email address.</p>
        </div>
      </div>

      <form v-else class="space-y-4" @submit.prevent="onSubmit">
        <div v-if="serverError" class="flex items-center gap-2.5 rounded-xl border border-rose-200 bg-rose-50 p-3.5 text-sm text-rose-700">
          <svg class="w-4 h-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          {{ serverError }}
        </div>

        <div class="grid grid-cols-2 gap-3">
          <div>
            <label class="label">First name</label>
            <input v-model="firstName" v-bind="firstNameAttrs" type="text" class="input" placeholder="Jane" />
            <p v-if="errors.firstName" class="mt-1.5 text-xs text-rose-600">{{ errors.firstName }}</p>
          </div>
          <div>
            <label class="label">Last name</label>
            <input v-model="lastName" v-bind="lastNameAttrs" type="text" class="input" placeholder="Doe" />
            <p v-if="errors.lastName" class="mt-1.5 text-xs text-rose-600">{{ errors.lastName }}</p>
          </div>
        </div>

        <div>
          <label class="label">Email address</label>
          <input v-model="email" v-bind="emailAttrs" type="email" class="input" placeholder="jane@example.com" />
          <p v-if="errors.email" class="mt-1.5 text-xs text-rose-600">{{ errors.email }}</p>
        </div>

        <div>
          <label class="label">Password</label>
          <PasswordInput v-model="password" v-bind="passwordAttrs" />
          <p v-if="errors.password" class="mt-1.5 text-xs text-rose-600">{{ errors.password }}</p>
        </div>

        <div>
          <label class="label">Confirm password</label>
          <PasswordInput v-model="confirmPassword" v-bind="confirmPasswordAttrs" />
          <p v-if="errors.confirmPassword" class="mt-1.5 text-xs text-rose-600">{{ errors.confirmPassword }}</p>
        </div>

        <button type="submit" class="btn-primary w-full mt-2">Create account</button>
      </form>
    </div>

    <p class="mt-6 text-center text-sm text-gray-500">
      <span class="text-gray-500 dark:text-gray-400">Already have an account?</span>
      <NuxtLink to="/login" class="font-semibold text-emerald-600 hover:text-emerald-700 transition-colors">Sign in</NuxtLink>
    </p>
  </div>
</template>
