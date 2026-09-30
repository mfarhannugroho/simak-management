import { apiClient } from './api'
import type { ApiResponse } from '@/types'
import type { Simak } from './simak'
import type { UserProfile } from '@/types'

export interface SimakDocument {
  id: string
  simak_id: string
  document_number: string
  file_name: string
  file_url: string
  doc_type: string
  qr_code_data: string
  issued_by?: string
  issued_at: string
  created_at: string
  simak?: Simak
  issuer?: UserProfile
}

export interface VerificationItemInfo {
  item_name: string
  brand: string
  model: string
  serial_number: string
  quantity: number
  unit: string
  description: string
}

export interface PublicVerificationResponse {
  valid: boolean
  status: 'VALID' | 'EXPIRED' | 'CANCELLED' | 'NOT_YET_VALID' | 'INVALID' | 'NOT_FOUND'
  reason?: string
  simak_number?: string
  applicant_name?: string
  worker_id?: string
  company?: string
  department?: string
  position?: string
  purpose?: string
  direction?: string
  destination_area?: string
  start_date?: string
  end_date?: string
  issued_at?: string
  total_items?: number
  items?: VerificationItemInfo[]
}

export interface IssuanceQueueItem {
  id: string
  simak_number: string
  purpose: string
  direction: 'MASUK' | 'KELUAR' | 'MASUK_KELUAR'
  destination_area: string
  start_date: string
  end_date: string
  status: string
  qr_token?: string
  issued_at?: string
  created_at: string
  worker_name: string
  worker_id: string
  company: string
  department: string
  total_items: number
}

export interface PaginatedIssuanceQueue {
  items: IssuanceQueueItem[]
  pagination: {
    page: number
    limit: number
    total_items: number
    total_pages: number
  }
}

export interface IssueSimakPayload {
  custom_number?: string
  notes?: string
}

export const documentService = {
  // Officer Issuance Queue
  async getIssuanceQueue(params?: { page?: number; limit?: number; status?: string; search?: string }) {
    const res = await apiClient.get<ApiResponse<PaginatedIssuanceQueue>>('/officer/issuance', { params })
    return res.data
  },

  // Issue Official SIMAK
  async issueSimak(simakId: string, payload?: IssueSimakPayload) {
    const res = await apiClient.post<ApiResponse<SimakDocument>>(`/simak/${simakId}/issue`, payload || {})
    return res.data
  },

  // Get Official Document Record
  async getDocument(simakId: string) {
    const res = await apiClient.get<ApiResponse<SimakDocument>>(`/simak/${simakId}/document`)
    return res.data
  },

  // Public QR Token Verification (No auth required)
  async verifyQRToken(qrToken: string) {
    const res = await apiClient.get<ApiResponse<PublicVerificationResponse>>(`/verification/${qrToken}`)
    return res.data
  },
}
