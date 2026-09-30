import { apiClient } from './api'
import type { ApiResponse } from '@/types'
import type { Simak } from './simak'
import type { Inspection, InspectionRecordItem } from './inspection'
import type { UserProfile } from '@/types'

export type DiscrepancyType =
  | 'QUANTITY_MISMATCH'
  | 'UNREGISTERED_ITEM'
  | 'MISSING_ITEM'
  | 'ITEM_DESCRIPTION_MISMATCH'
  | 'OTHER'

export type DiscrepancyStatus = 'OPEN' | 'IN_REVIEW' | 'RESOLVED' | 'CLOSED'

export type ResolutionAction =
  | 'ALLOW_ENTRY'
  | 'DENY_ENTRY'
  | 'CONFISCATE_HOLD_AT_GATE'
  | 'RETURN_TO_VENDOR'
  | 'REVISE_PERMIT'
  | 'WARNING_ISSUED'

export interface Discrepancy {
  id: string
  discrepancy_number: string
  simak_id: string
  inspection_id: string
  inspection_item_id?: string
  item_name: string
  brand?: string
  serial_number?: string
  discrepancy_type: DiscrepancyType
  expected_qty: number
  actual_qty: number
  unit: string
  description?: string
  evidence_photo_url?: string
  reported_by: string
  reviewed_by?: string
  reviewed_at?: string
  resolution_action?: ResolutionAction
  resolution_notes?: string
  status: DiscrepancyStatus
  created_at: string
  updated_at: string

  // Relations
  simak?: Simak
  inspection?: Inspection
  inspection_item?: InspectionRecordItem
  reporter?: UserProfile
  reviewer?: UserProfile
}

export interface DiscrepancyListItem {
  id: string
  discrepancy_number: string
  simak_id: string
  simak_number: string
  worker_name: string
  company: string
  item_name: string
  brand?: string
  serial_number?: string
  discrepancy_type: DiscrepancyType
  expected_qty: number
  actual_qty: number
  unit: string
  status: DiscrepancyStatus
  reported_by_name: string
  created_at: string
}

export interface DiscrepancyMetrics {
  total: number
  open: number
  in_review: number
  resolved: number
  closed: number
}

export interface DiscrepancyFilterParams {
  page?: number
  limit?: number
  status?: string
  discrepancy_type?: string
  simak_id?: string
  search?: string
  date?: string
}

export interface PaginatedDiscrepancies {
  items: DiscrepancyListItem[]
  pagination: {
    page: number
    limit: number
    total_items: number
    total_pages: number
  }
}

export interface CreateDiscrepancyPayload {
  simak_id: string
  inspection_id: string
  inspection_item_id?: string
  item_name: string
  brand?: string
  serial_number?: string
  discrepancy_type: DiscrepancyType
  expected_qty: number
  actual_qty: number
  unit?: string
  description?: string
  evidence_photo_url?: string
}

export interface ResolveDiscrepancyPayload {
  resolution_action: ResolutionAction
  resolution_notes: string
  status?: 'RESOLVED' | 'CLOSED'
}

export const discrepancyService = {
  // List discrepancies
  async getDiscrepancies(params?: DiscrepancyFilterParams) {
    const res = await apiClient.get<ApiResponse<PaginatedDiscrepancies>>('/discrepancies', { params })
    return res.data
  },

  // Get discrepancy detail
  async getDiscrepancy(id: string) {
    const res = await apiClient.get<ApiResponse<Discrepancy>>(`/discrepancies/${id}`)
    return res.data
  },

  // Get metrics
  async getMetrics() {
    const res = await apiClient.get<ApiResponse<DiscrepancyMetrics>>('/discrepancies/metrics')
    return res.data
  },

  // Create manual discrepancy
  async createDiscrepancy(payload: CreateDiscrepancyPayload) {
    const res = await apiClient.post<ApiResponse<Discrepancy>>('/discrepancies', payload)
    return res.data
  },

  // Mark as IN_REVIEW
  async startReview(id: string) {
    const res = await apiClient.put<ApiResponse<Discrepancy>>(`/discrepancies/${id}/review`)
    return res.data
  },

  // Resolve discrepancy
  async resolveDiscrepancy(id: string, payload: ResolveDiscrepancyPayload) {
    const res = await apiClient.put<ApiResponse<Discrepancy>>(`/discrepancies/${id}/resolve`, payload)
    return res.data
  },
}
