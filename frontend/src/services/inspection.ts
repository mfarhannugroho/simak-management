import { apiClient } from './api'
import type { ApiResponse } from '@/types'
import type { Simak } from './simak'
import type { UserProfile } from '@/types'

export interface InspectionItemInput {
  simak_item_id?: string
  item_name: string
  brand?: string
  serial_number?: string
  expected_qty: number
  actual_qty: number
  unit?: string
  status?: string // 'MATCH' | 'MISMATCH' | 'UNREGISTERED'
  notes?: string
  photo_url?: string
}

export interface CreateInspectionPayload {
  simak_id: string
  gate_name?: string
  inspection_type: 'MASUK' | 'KELUAR'
  notes?: string
  items: InspectionItemInput[]
}

export interface InspectionRecordItem {
  id: string
  inspection_id: string
  simak_item_id?: string
  item_name: string
  brand?: string
  serial_number?: string
  expected_qty: number
  actual_qty: number
  unit?: string
  status: string
  notes?: string
  photo_url?: string
  created_at: string
}

export interface Inspection {
  id: string
  simak_id: string
  inspector_id: string
  gate_name: string
  inspection_type: 'MASUK' | 'KELUAR'
  status: 'MATCH' | 'MISMATCH'
  total_registered_items: number
  total_actual_items: number
  notes?: string
  inspected_at: string
  created_at: string
  simak?: Simak
  inspector?: UserProfile
  items?: InspectionRecordItem[]
}

export interface InspectionListItem {
  id: string
  simak_id: string
  simak_number: string
  inspector_name: string
  worker_name: string
  company: string
  gate_name: string
  inspection_type: 'MASUK' | 'KELUAR'
  status: 'MATCH' | 'MISMATCH'
  total_registered_items: number
  total_actual_items: number
  notes: string
  inspected_at: string
}

export interface GateDashboardMetrics {
  inspections_today: number
  active_simaks: number
  mismatch_count: number
  recent_movements: Inspection[]
}

export interface InspectionFilterParams {
  page?: number
  limit?: number
  simak_id?: string
  gate_name?: string
  inspection_type?: string
  status?: string
  search?: string
  date?: string
}

export interface PaginatedInspections {
  items: InspectionListItem[]
  pagination: {
    page: number
    limit: number
    total_items: number
    total_pages: number
  }
}

export const inspectionService = {
  // Gate metrics
  async getMetrics() {
    const res = await apiClient.get<ApiResponse<GateDashboardMetrics>>('/gate/metrics')
    return res.data
  },

  // Record physical inspection
  async recordInspection(payload: CreateInspectionPayload) {
    const res = await apiClient.post<ApiResponse<Inspection>>('/verification/inspect', payload)
    return res.data
  },

  // List inspection history
  async listInspections(params?: InspectionFilterParams) {
    const res = await apiClient.get<ApiResponse<PaginatedInspections>>('/inspections', { params })
    return res.data
  },

  // Detail inspection
  async getInspection(id: string) {
    const res = await apiClient.get<ApiResponse<Inspection>>(`/inspections/${id}`)
    return res.data
  },
}
