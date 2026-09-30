import { apiClient } from './api'

export interface AuditLogItem {
  id: string
  user_id?: string
  user_email: string
  user_name: string
  action: string
  entity_type: string
  entity_id: string
  details: string
  ip_address: string
  user_agent: string
  created_at: string
}

export interface AuditPaginationMeta {
  page: number
  page_size: number
  total: number
  total_pages: number
}

export interface AuditFilter {
  action?: string
  entity_type?: string
  user_id?: string
  search?: string
  date_from?: string
  date_to?: string
  page?: number
  page_size?: number
}

export interface PaginatedAuditLogs {
  data: AuditLogItem[]
  meta: AuditPaginationMeta
}

export const auditService = {
  async getAuditLogs(params: AuditFilter = {}): Promise<PaginatedAuditLogs> {
    const res = await apiClient.get<{ success: boolean; data: AuditLogItem[]; meta: AuditPaginationMeta }>('/audit-logs', { params })
    return { data: res.data.data, meta: res.data.meta }
  },

  async getAuditDetail(id: string): Promise<AuditLogItem> {
    const res = await apiClient.get<{ success: boolean; data: AuditLogItem }>(`/audit-logs/${id}`)
    return res.data.data
  },

  async exportCSV(params: AuditFilter = {}): Promise<void> {
    const response = await apiClient.get('/audit-logs/export', {
      params,
      responseType: 'blob',
    })

    const blob = new Blob([response.data], { type: 'text/csv;charset=utf-8;' })
    const link = document.createElement('a')
    const url = URL.createObjectURL(blob)
    const filename = `audit_trail_simak_${new Date().toISOString().slice(0, 10)}.csv`

    link.setAttribute('href', url)
    link.setAttribute('download', filename)
    link.style.visibility = 'hidden'
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(url)
  }
}
