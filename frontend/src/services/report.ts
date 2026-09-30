import { apiClient } from './api'

export interface SimakReportItem {
  id: string
  simak_number: string
  worker_name: string
  worker_nip: string
  department_name: string
  company_name: string
  direction: string
  destination_area: string
  start_date: string
  end_date: string
  status: string
  total_items: number
  issued_at?: string
  created_at: string
}

export interface VerificationReportItem {
  id: string
  simak_number: string
  worker_name: string
  inspector_name: string
  gate_name: string
  inspection_type: string
  status: string
  total_registered_items: number
  total_actual_items: number
  notes: string
  inspected_at: string
}

export interface DiscrepancyReportItem {
  id: string
  discrepancy_number: string
  simak_number: string
  worker_name: string
  item_name: string
  discrepancy_type: string
  expected_qty: number
  actual_qty: number
  status: string
  resolution_notes: string
  reviewed_at?: string
  created_at: string
}

export interface ReportPaginationMeta {
  page: number
  page_size: number
  total: number
  total_pages: number
}

export interface ReportFilter {
  date_from?: string
  date_to?: string
  department_id?: string
  company_id?: string
  direction?: string
  status?: string
  gate_name?: string
  inspection_type?: string
  search?: string
  resolved?: string
  page?: number
  page_size?: number
}

export interface PaginatedReport<T> {
  data: T[]
  meta: ReportPaginationMeta
}

export const reportService = {
  async getSimakReport(params: ReportFilter = {}): Promise<PaginatedReport<SimakReportItem>> {
    const res = await apiClient.get<{ success: boolean; data: SimakReportItem[]; meta: ReportPaginationMeta }>('/reports/simak', { params })
    return { data: res.data.data, meta: res.data.meta }
  },

  async getVerificationReport(params: ReportFilter = {}): Promise<PaginatedReport<VerificationReportItem>> {
    const res = await apiClient.get<{ success: boolean; data: VerificationReportItem[]; meta: ReportPaginationMeta }>('/reports/verification', { params })
    return { data: res.data.data, meta: res.data.meta }
  },

  async getDiscrepancyReport(params: ReportFilter = {}): Promise<PaginatedReport<DiscrepancyReportItem>> {
    const res = await apiClient.get<{ success: boolean; data: DiscrepancyReportItem[]; meta: ReportPaginationMeta }>('/reports/discrepancies', { params })
    return { data: res.data.data, meta: res.data.meta }
  },

  async exportCSV(type: 'simak' | 'verification' | 'discrepancy', params: ReportFilter = {}): Promise<void> {
    const response = await apiClient.get('/reports/export', {
      params: { ...params, type },
      responseType: 'blob',
    })

    const blob = new Blob([response.data], { type: 'text/csv;charset=utf-8;' })
    const link = document.createElement('a')
    const url = URL.createObjectURL(blob)
    
    // Extract filename from header or fallback
    const disposition = response.headers['content-disposition']
    let filename = `laporan_${type}_${new Date().toISOString().slice(0, 10)}.csv`
    if (disposition && disposition.indexOf('filename=') !== -1) {
      const match = disposition.match(/filename="?([^"]+)"?/)
      if (match && match[1]) {
        filename = match[1]
      }
    }

    link.setAttribute('href', url)
    link.setAttribute('download', filename)
    link.style.visibility = 'hidden'
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(url)
  }
}
