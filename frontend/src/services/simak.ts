import { apiClient } from './api'
import type { ApiResponse } from '@/types'
import type { ItemCategory, Worker } from './masterData'

export interface SimakItem {
  id?: string
  simak_id?: string
  category_id?: string
  item_name: string
  brand?: string
  model?: string
  serial_number?: string
  size_spec?: string
  quantity: number
  unit: string
  description?: string
  photo_url?: string
  created_at?: string
  category?: ItemCategory
}

export interface SimakAttachment {
  id?: string
  simak_id?: string
  file_name: string
  file_url: string
  file_type?: string
  file_size?: number
  attachment_type: string
  created_at?: string
}

export interface Simak {
  id: string
  worker_id: string
  simak_number?: string
  purpose: string
  direction: 'MASUK' | 'KELUAR' | 'MASUK_KELUAR'
  destination_area: string
  start_date: string
  end_date: string
  status:
    | 'DRAFT'
    | 'SUBMITTED'
    | 'UNDER_REVIEW'
    | 'REQUIRES_REVISION'
    | 'REJECTED'
    | 'APPROVED'
    | 'ISSUED'
    | 'ACTIVE'
    | 'VERIFIED'
    | 'COMPLETED'
    | 'CANCELLED'
    | 'EXPIRED'
  notes?: string
  revision_notes?: string
  rejection_reason?: string
  qr_token?: string
  issued_at?: string
  issued_by?: string
  created_at: string
  updated_at: string
  worker?: Worker
  items?: SimakItem[]
  attachments?: SimakAttachment[]
}

export interface SimakPayload {
  purpose: string
  direction: string
  destination_area: string
  start_date: string
  end_date: string
  notes?: string
  submit_directly?: boolean
  items: SimakItem[]
  attachments?: SimakAttachment[]
}

export interface SimakListParams {
  status?: string
  direction?: string
  search?: string
  start_date?: string
  end_date?: string
  worker_id?: string
  page?: number
  limit?: number
}

export interface PaginatedSimak {
  items: Simak[]
  pagination: {
    page: number
    limit: number
    total_items: number
    total_pages: number
  }
}

export const simakService = {
  async getSimaks(params?: SimakListParams) {
    const res = await apiClient.get<ApiResponse<PaginatedSimak>>('/simak', { params })
    return res.data
  },

  async getSimak(id: string) {
    const res = await apiClient.get<ApiResponse<Simak>>(`/simak/${id}`)
    return res.data
  },

  async createSimak(payload: SimakPayload) {
    const res = await apiClient.post<ApiResponse<Simak>>('/simak', payload)
    return res.data
  },

  async updateSimak(id: string, payload: Partial<SimakPayload>) {
    const res = await apiClient.put<ApiResponse<Simak>>(`/simak/${id}`, payload)
    return res.data
  },

  async deleteSimak(id: string) {
    const res = await apiClient.delete<ApiResponse<null>>(`/simak/${id}`)
    return res.data
  },

  async submitSimak(id: string, notes?: string) {
    const res = await apiClient.post<ApiResponse<Simak>>(`/simak/${id}/submit`, { notes })
    return res.data
  },

  async cancelSimak(id: string, reason?: string) {
    const res = await apiClient.post<ApiResponse<Simak>>(`/simak/${id}/cancel`, { reason })
    return res.data
  },

  async addAttachment(simakId: string, attachment: Partial<SimakAttachment>) {
    const res = await apiClient.post<ApiResponse<SimakAttachment>>(`/simak/${simakId}/attachments`, attachment)
    return res.data
  },

  async deleteAttachment(attachmentId: string) {
    const res = await apiClient.delete<ApiResponse<null>>(`/simak/attachments/${attachmentId}`)
    return res.data
  },
}
