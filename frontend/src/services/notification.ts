import { apiClient } from './api'
import type { ApiResponse } from '@/types'

export interface NotificationItem {
  id: string
  title: string
  message: string
  type: string
  reference_id?: string
  reference_type?: string
  action_url?: string
  is_read: boolean
  read_at?: string
  created_at: string
}

export interface PaginatedNotifications {
  items: NotificationItem[]
  pagination: {
    page: number
    limit: number
    total_items: number
    total_pages: number
  }
}

export interface UnreadCountData {
  unread_count: number
}

export interface NotificationFilterParams {
  page?: number
  limit?: number
  unread_only?: boolean
  type?: string
}

export const notificationService = {
  // Get notifications
  async getNotifications(params?: NotificationFilterParams) {
    const res = await apiClient.get<ApiResponse<PaginatedNotifications>>('/notifications', { params })
    return res.data
  },

  // Get unread count
  async getUnreadCount() {
    const res = await apiClient.get<ApiResponse<UnreadCountData>>('/notifications/unread-count')
    return res.data
  },

  // Mark single as read
  async markAsRead(id: string) {
    const res = await apiClient.put<ApiResponse<NotificationItem>>(`/notifications/${id}/read`)
    return res.data
  },

  // Mark all as read
  async markAllAsRead() {
    const res = await apiClient.put<ApiResponse<null>>('/notifications/read-all')
    return res.data
  },
}
