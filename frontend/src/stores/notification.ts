import { defineStore } from 'pinia'
import { ref } from 'vue'
import { notificationService, type NotificationItem } from '@/services/notification'

export const useNotificationStore = defineStore('notification', () => {
  const unreadCount = ref<number>(0)
  const recentNotifications = ref<NotificationItem[]>([])
  const loading = ref<boolean>(false)
  const wsConnected = ref<boolean>(false)

  let socket: WebSocket | null = null
  let reconnectTimer: any = null

  async function fetchUnreadCount() {
    try {
      const res = await notificationService.getUnreadCount()
      if (res.data) {
        unreadCount.value = res.data.unread_count
      }
    } catch (err) {
      console.error('Failed to fetch unread notification count:', err)
    }
  }

  async function fetchRecent() {
    loading.value = true
    try {
      const res = await notificationService.getNotifications({ limit: 5 })
      if (res.data && res.data.items) {
        recentNotifications.value = res.data.items
      }
    } catch (err) {
      console.error('Failed to fetch recent notifications:', err)
    } finally {
      loading.value = false
    }
  }

  async function markAsRead(id: string) {
    try {
      await notificationService.markAsRead(id)
      const found = recentNotifications.value.find((n) => n.id === id)
      if (found && !found.is_read) {
        found.is_read = true
        if (unreadCount.value > 0) unreadCount.value--
      }
    } catch (err) {
      console.error('Failed to mark notification as read:', err)
    }
  }

  async function markAllAsRead() {
    try {
      await notificationService.markAllAsRead()
      unreadCount.value = 0
      recentNotifications.value.forEach((n) => (n.is_read = true))
    } catch (err) {
      console.error('Failed to mark all notifications as read:', err)
    }
  }

  function connectWebSocket(token?: string) {
    if (socket && (socket.readyState === WebSocket.OPEN || socket.readyState === WebSocket.CONNECTING)) {
      return
    }

    const authToken = token || localStorage.getItem('simak_token') || ''
    if (!authToken) return

    const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const wsHost = '127.0.0.1:8080'
    const wsUrl = `${wsProtocol}//${wsHost}/api/v1/ws?token=${authToken}`

    try {
      socket = new WebSocket(wsUrl)

      socket.onopen = () => {
        wsConnected.value = true
        clearTimeout(reconnectTimer)
        console.log('⚡ Connected to SIMAK Real-time Notification Stream')
      }

      socket.onmessage = (event) => {
        try {
          const notif = JSON.parse(event.data) as NotificationItem
          if (notif && notif.id) {
            unreadCount.value++
            recentNotifications.value.unshift(notif)
            if (recentNotifications.value.length > 10) {
              recentNotifications.value.pop()
            }

            // Browser notification if permitted
            if ('Notification' in window && Notification.permission === 'granted') {
              new Notification(notif.title, {
                body: notif.message,
                icon: '/favicon.ico',
              })
            }
          }
        } catch (err) {
          console.warn('Error parsing incoming WS message:', err)
        }
      }

      socket.onclose = () => {
        wsConnected.value = false
        // Try reconnecting after 5 seconds
        reconnectTimer = setTimeout(() => {
          connectWebSocket(token)
        }, 5000)
      }

      socket.onerror = (err) => {
        console.warn('WebSocket connection error:', err)
        wsConnected.value = false
      }
    } catch (err) {
      console.error('Failed to initialize WebSocket:', err)
    }
  }

  function disconnectWebSocket() {
    clearTimeout(reconnectTimer)
    if (socket) {
      socket.close()
      socket = null
    }
    wsConnected.value = false
  }

  return {
    unreadCount,
    recentNotifications,
    loading,
    wsConnected,
    fetchUnreadCount,
    fetchRecent,
    markAsRead,
    markAllAsRead,
    connectWebSocket,
    disconnectWebSocket,
  }
})
