import type { Notification } from '~/types/user'

export function useNotifications() {
  const config = useRuntimeConfig()
  const authStore = useAuthStore()
  const notifStore = useNotificationsStore()

  let ws: WebSocket | null = null
  let reconnectTimer: ReturnType<typeof setTimeout> | null = null
  let reconnectDelay = 1000

  function connect() {
    if (!authStore.accessToken) return

    const url = `${config.public.wsBase}/ws/notifications?token=${authStore.accessToken}`
    ws = new WebSocket(url)

    ws.onopen = () => {
      notifStore.wsConnected = true
      reconnectDelay = 1000
    }

    ws.onmessage = (event) => {
      try {
        const notification = JSON.parse(event.data) as Notification
        notifStore.addFromWebSocket(notification)
      }
      catch {
        // ignore malformed messages
      }
    }

    ws.onclose = () => {
      notifStore.wsConnected = false
      scheduleReconnect()
    }

    ws.onerror = () => {
      ws?.close()
    }
  }

  async function scheduleReconnect() {
    if (reconnectTimer) clearTimeout(reconnectTimer)
    reconnectTimer = setTimeout(async () => {
      if (!authStore.isAuthenticated) return
      await authStore.refresh()
      connect()
      reconnectDelay = Math.min(reconnectDelay * 2, 30000)
    }, reconnectDelay)
  }

  function disconnect() {
    if (reconnectTimer) clearTimeout(reconnectTimer)
    ws?.close()
    ws = null
    notifStore.wsConnected = false
  }

  onMounted(() => {
    if (authStore.isAuthenticated) {
      connect()
      notifStore.fetchUnreadCount()
    }
  })

  onUnmounted(() => {
    disconnect()
  })

  return {
    wsConnected: computed(() => notifStore.wsConnected),
  }
}
