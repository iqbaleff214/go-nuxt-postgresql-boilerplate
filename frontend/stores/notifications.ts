import { defineStore } from 'pinia'
import type { Notification } from '~/types/user'

export const useNotificationsStore = defineStore('notifications', {
  state: () => ({
    notifications: [] as Notification[],
    unreadCount: 0,
    wsConnected: false,
  }),

  actions: {
    async fetchUnreadCount() {
      const api = useApi()
      const res = await api.get<{ count: number }>('/notifications/unread-count')
      this.unreadCount = res.data.count
    },

    async fetchNotifications(page = 1) {
      const api = useApi()
      const res = await api.get<Notification[]>(`/notifications?page=${page}&page_size=20`)
      if (page === 1) {
        this.notifications = res.data
      }
      else {
        this.notifications.push(...res.data)
      }
    },

    async markRead(id: string) {
      const api = useApi()
      await api.patch(`/notifications/${id}/read`)
      const n = this.notifications.find(n => n.id === id)
      if (n && !n.read_at) {
        n.read_at = new Date().toISOString()
        this.unreadCount = Math.max(0, this.unreadCount - 1)
      }
    },

    async markAllRead() {
      const api = useApi()
      await api.patch('/notifications/read-all')
      const now = new Date().toISOString()
      this.notifications.forEach(n => { if (!n.read_at) n.read_at = now })
      this.unreadCount = 0
    },

    addFromWebSocket(notification: Notification) {
      this.notifications.unshift(notification)
      if (!notification.read_at) {
        this.unreadCount++
      }
    },
  },
})
