import { apiClient } from './api'
import type { ApiResponse } from '@/types'

export interface Company {
  id: string
  organization_id?: string
  name: string
  type: string
  address: string
  contact_phone: string
  contact_email: string
  is_active: boolean
  created_at: string
}

export interface Department {
  id: string
  organization_id?: string
  name: string
  code: string
  is_active: boolean
  created_at: string
}

export interface Worker {
  id: string
  user_id: string
  company_id?: string
  department_id?: string
  worker_id: string
  full_name: string
  position: string
  phone: string
  created_at: string
  company?: Company
  department?: Department
}

export interface ItemCategory {
  id: string
  name: string
  description: string
  is_active: boolean
  created_at: string
}

export interface WorkerRegisteredItem {
  id: string
  worker_id: string
  category_id?: string
  item_name: string
  brand: string
  model: string
  serial_number: string
  quantity: number
  unit: string
  description: string
  photo_url?: string
  created_at: string
  category?: ItemCategory
}

export interface PaginatedResult<T> {
  items: T[]
  pagination: {
    page: number
    limit: number
    total_items: number
    total_pages: number
  }
}

export const masterDataService = {
  // Companies
  async getCompanies(params?: { search?: string; type?: string; page?: number; limit?: number }) {
    const res = await apiClient.get<ApiResponse<PaginatedResult<Company>>>('/companies', { params })
    return res.data
  },
  async createCompany(data: Partial<Company>) {
    const res = await apiClient.post<ApiResponse<Company>>('/companies', data)
    return res.data
  },
  async updateCompany(id: string, data: Partial<Company>) {
    const res = await apiClient.put<ApiResponse<Company>>(`/companies/${id}`, data)
    return res.data
  },
  async deleteCompany(id: string) {
    const res = await apiClient.delete<ApiResponse<null>>(`/companies/${id}`)
    return res.data
  },

  // Departments
  async getDepartments(params?: { search?: string; page?: number; limit?: number }) {
    const res = await apiClient.get<ApiResponse<PaginatedResult<Department>>>('/departments', { params })
    return res.data
  },
  async createDepartment(data: Partial<Department>) {
    const res = await apiClient.post<ApiResponse<Department>>('/departments', data)
    return res.data
  },
  async updateDepartment(id: string, data: Partial<Department>) {
    const res = await apiClient.put<ApiResponse<Department>>(`/departments/${id}`, data)
    return res.data
  },
  async deleteDepartment(id: string) {
    const res = await apiClient.delete<ApiResponse<null>>(`/departments/${id}`)
    return res.data
  },

  // Workers
  async getWorkers(params?: { search?: string; company_id?: string; department_id?: string; page?: number; limit?: number }) {
    const res = await apiClient.get<ApiResponse<PaginatedResult<Worker>>>('/workers', { params })
    return res.data
  },
  async createWorker(data: Partial<Worker>) {
    const res = await apiClient.post<ApiResponse<Worker>>('/workers', data)
    return res.data
  },
  async updateWorker(id: string, data: Partial<Worker>) {
    const res = await apiClient.put<ApiResponse<Worker>>(`/workers/${id}`, data)
    return res.data
  },
  async deleteWorker(id: string) {
    const res = await apiClient.delete<ApiResponse<null>>(`/workers/${id}`)
    return res.data
  },

  // Item Categories
  async getCategories(params?: { search?: string; active_only?: boolean }) {
    const res = await apiClient.get<ApiResponse<ItemCategory[]>>('/item-categories', { params })
    return res.data
  },
  async createCategory(data: Partial<ItemCategory>) {
    const res = await apiClient.post<ApiResponse<ItemCategory>>('/item-categories', data)
    return res.data
  },
  async updateCategory(id: string, data: Partial<ItemCategory>) {
    const res = await apiClient.put<ApiResponse<ItemCategory>>(`/item-categories/${id}`, data)
    return res.data
  },
  async deleteCategory(id: string) {
    const res = await apiClient.delete<ApiResponse<null>>(`/item-categories/${id}`)
    return res.data
  },

  // Worker Registered Items (Barang Bawaan Rutin Pekerja)
  async getMyRegisteredItems() {
    const res = await apiClient.get<ApiResponse<WorkerRegisteredItem[]>>('/workers/me/registered-items')
    return res.data
  },
  async getWorkerRegisteredItems(workerId: string) {
    const res = await apiClient.get<ApiResponse<WorkerRegisteredItem[]>>(`/workers/${workerId}/registered-items`)
    return res.data
  },
  async createRegisteredItem(data: Partial<WorkerRegisteredItem>) {
    const res = await apiClient.post<ApiResponse<WorkerRegisteredItem>>('/workers/me/registered-items', data)
    return res.data
  },
  async updateRegisteredItem(itemId: string, data: Partial<WorkerRegisteredItem>) {
    const res = await apiClient.put<ApiResponse<WorkerRegisteredItem>>(`/registered-items/${itemId}`, data)
    return res.data
  },
  async deleteRegisteredItem(itemId: string) {
    const res = await apiClient.delete<ApiResponse<null>>(`/registered-items/${itemId}`)
    return res.data
  },
}
