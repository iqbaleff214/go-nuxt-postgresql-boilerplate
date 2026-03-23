import { defineStore } from 'pinia'
import type { User } from '~/types/user'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    accessToken: null as string | null,
    user: null as User | null,
    mfaChallengeToken: null as string | null,
  }),

  getters: {
    isAuthenticated: (state) => !!state.accessToken,
    isSuperadmin: (state) => state.user?.role === 'superadmin',
  },

  actions: {
    async login(email: string, password: string): Promise<{ requires2fa: boolean }> {
      const api = useApi()
      const res = await api.post<{ access_token?: string; mfa_challenge_token?: string }>(
        '/auth/login',
        { email, password },
      )

      if (res.data.mfa_challenge_token) {
        this.mfaChallengeToken = res.data.mfa_challenge_token
        return { requires2fa: true }
      }

      this.accessToken = res.data.access_token ?? null
      this.mfaChallengeToken = null
      await this.fetchMe()
      return { requires2fa: false }
    },

    async verify2fa(code: string): Promise<void> {
      const api = useApi()
      const res = await api.post<{ access_token: string }>('/auth/2fa/verify', {
        mfa_challenge_token: this.mfaChallengeToken,
        code,
      })
      this.accessToken = res.data.access_token
      this.mfaChallengeToken = null
      await this.fetchMe()
    },

    async fetchMe(): Promise<void> {
      const api = useApi()
      const res = await api.get<User>('/profile')
      this.user = res.data
    },

    async refresh(): Promise<boolean> {
      try {
        const config = useRuntimeConfig()
        const res = await fetch(`${config.public.apiBase}/auth/refresh`, {
          method: 'POST',
          credentials: 'include',
        })
        if (!res.ok) return false
        const body = await res.json()
        this.accessToken = body.data?.access_token ?? null
        return !!this.accessToken
      }
      catch {
        return false
      }
    },

    async logout(): Promise<void> {
      try {
        const api = useApi()
        await api.post('/auth/logout')
      }
      catch {
        // ignore errors on logout
      }
      finally {
        this.accessToken = null
        this.user = null
        this.mfaChallengeToken = null
      }
    },
  },

  persist: {
    pick: ['accessToken'],
  },
})
