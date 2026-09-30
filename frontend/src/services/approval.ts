import { apiClient } from './api'
import type { ApiResponse } from '@/types'
import type { Simak } from './simak'
import type { Worker, Department, Company } from './masterData'

export interface PendingApprovalItem {
  approval_id: string
  simak_id: string
  simak_number: string
  purpose: string
  direction: 'MASUK' | 'KELUAR' | 'MASUK_KELUAR'
  destination_area: string
  start_date: string
  end_date: string
  simak_status: string
  current_step: number
  total_steps: number
  step_name: string
  required_role: string
  step_status: string
  submitted_at: string
  total_items: number
  worker?: Worker
}

export interface SimakApprovalStep {
  id: string
  simak_id: string
  step_order: number
  step_name: string
  required_role: string
  assigned_user_id?: string
  status: 'PENDING' | 'WAITING' | 'APPROVED' | 'REJECTED' | 'REVISION_REQUESTED'
  action_by_user_id?: string
  action_by_name?: string
  action_at?: string
  comments?: string
  created_at: string
  updated_at: string
}

export interface ApprovalDetailResponse {
  simak: Simak
  current_approval?: SimakApprovalStep
  workflow_steps: SimakApprovalStep[]
  can_take_action: boolean
  restriction_note?: string
}

export interface ApprovalMatrixStep {
  id?: string
  matrix_id?: string
  step_order: number
  step_name: string
  required_role: string
  specific_user_id?: string
}

export interface ApprovalMatrix {
  id: string
  name: string
  description?: string
  department_id?: string
  company_id?: string
  is_active: boolean
  created_at: string
  updated_at: string
  department?: Department
  company?: Company
  steps: ApprovalMatrixStep[]
}

export interface CreateMatrixPayload {
  name: string
  description?: string
  department_id?: string | null
  company_id?: string | null
  is_active: boolean
  steps: {
    step_order: number
    step_name: string
    required_role: string
    specific_user_id?: string | null
  }[]
}

export interface ApprovalDecisionPayload {
  decision: 'APPROVE' | 'REJECT' | 'REQUEST_REVISION'
  comments: string
}

export interface PendingListParams {
  page?: number
  limit?: number
  search?: string
  status?: string
}

export interface PaginatedPendingApprovals {
  items: PendingApprovalItem[]
  pagination: {
    page: number
    limit: number
    total_items: number
    total_pages: number
  }
}

export const approvalService = {
  // Pending Queue
  async getPendingApprovals(params?: PendingListParams) {
    const res = await apiClient.get<ApiResponse<PaginatedPendingApprovals>>('/approvals/pending', { params })
    return res.data
  },

  // Review Detail
  async getApprovalDetail(simakId: string) {
    const res = await apiClient.get<ApiResponse<ApprovalDetailResponse>>(`/approvals/detail/${simakId}`)
    return res.data
  },

  // Generic decision
  async submitDecision(simakId: string, payload: ApprovalDecisionPayload) {
    const res = await apiClient.post<ApiResponse<Simak>>(`/approvals/${simakId}/decision`, payload)
    return res.data
  },

  // Specific shortcuts
  async approve(simakId: string, comments?: string) {
    const res = await apiClient.post<ApiResponse<Simak>>(`/simak/${simakId}/approve`, { comments })
    return res.data
  },

  async reject(simakId: string, comments: string) {
    const res = await apiClient.post<ApiResponse<Simak>>(`/simak/${simakId}/reject`, { comments })
    return res.data
  },

  async requestRevision(simakId: string, comments: string) {
    const res = await apiClient.post<ApiResponse<Simak>>(`/simak/${simakId}/revision`, { comments })
    return res.data
  },

  // Admin Approval Matrices
  async getMatrices() {
    const res = await apiClient.get<ApiResponse<ApprovalMatrix[]>>('/admin/approval-matrices')
    return res.data
  },

  async getMatrix(id: string) {
    const res = await apiClient.get<ApiResponse<ApprovalMatrix>>(`/admin/approval-matrices/${id}`)
    return res.data
  },

  async createMatrix(payload: CreateMatrixPayload) {
    const res = await apiClient.post<ApiResponse<ApprovalMatrix>>('/admin/approval-matrices', payload)
    return res.data
  },

  async updateMatrix(id: string, payload: Partial<CreateMatrixPayload>) {
    const res = await apiClient.put<ApiResponse<ApprovalMatrix>>(`/admin/approval-matrices/${id}`, payload)
    return res.data
  },

  async deleteMatrix(id: string) {
    const res = await apiClient.delete<ApiResponse<null>>(`/admin/approval-matrices/${id}`)
    return res.data
  },
}
