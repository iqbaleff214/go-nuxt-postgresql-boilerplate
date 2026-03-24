<template>
  <div class="min-h-screen flex bg-gray-50">
    <!-- Left branding panel -->
    <div class="hidden lg:flex lg:w-[44%] bg-white border-r border-gray-100 flex-col justify-between p-12">
      <div class="flex items-center gap-2.5">
        <div class="w-8 h-8 rounded-lg bg-emerald-500 flex items-center justify-center">
          <svg class="w-4 h-4 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 13l4 4L19 7" />
          </svg>
        </div>
        <span class="text-gray-900 font-bold text-xl tracking-tight">MyApp</span>
      </div>

      <div class="space-y-8">
        <div class="space-y-3">
          <div v-for="item in features" :key="item.text" class="flex items-start gap-3">
            <div class="mt-0.5 w-5 h-5 rounded-full bg-emerald-100 flex items-center justify-center shrink-0">
              <svg class="w-3 h-3 text-emerald-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 13l4 4L19 7" />
              </svg>
            </div>
            <span class="text-sm text-gray-600">{{ item.text }}</span>
          </div>
        </div>
      </div>

      <p class="text-xs text-gray-400">© 2026 MyApp. All rights reserved.</p>
    </div>

    <!-- Right form panel -->
    <div class="flex-1 flex items-center justify-center p-6">
      <div class="w-full max-w-md">
        <!-- Mobile logo -->
        <div class="mb-10 lg:hidden flex items-center justify-center gap-2.5">
          <div class="w-8 h-8 rounded-lg bg-emerald-500 flex items-center justify-center">
            <svg class="w-4 h-4 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 13l4 4L19 7" />
            </svg>
          </div>
          <span class="text-gray-900 font-bold text-xl tracking-tight">MyApp</span>
        </div>
        <slot />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const features = [
  { text: 'Secure authentication with 2FA support' },
  { text: 'Real-time notifications via WebSocket' },
  { text: 'Role-based access control' },
  { text: 'Cloud storage with S3-compatible backends' },
]
</script>
