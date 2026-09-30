import { apiClient } from './api'
import type { ApiResponse } from '@/types'

export interface AdminUser {
  id: string
  email: string
  is_active: boolean
  last_login_at: string | null
  created_at: string
  roles: string[]
  worker_id?: string
  full_name?: string
  position?: string
  phone?: string
  company_name?: string
  department_name?: string
}

export interface PermissionItem {
  id: string
  name: string
  description: string
}

export interface RoleItem {
  id: string
  name: string
  description: string
  permissions: PermissionItem[]
}

export const adminUserService = {
  async getUsers(params?: { search?: string; role?: string }): Promise<ApiResponse<AdminUser[]>> {
    const res = await apiClient.get<ApiResponse<AdminUser[]>>('/admin/users', { params })
    return res.data
  },

  async updateUserRole(userId: string, role: string): Promise<ApiResponse<null>> {
    const res = await apiClient.put<ApiResponse<null>>(`/admin/users/${userId}/role`, { role })
    return res.data
  },

  async toggleUserStatus(userId: string, isActive: boolean): Promise<ApiResponse<null>> {
    const res = await apiClient.put<ApiResponse<null>>(`/admin/users/${userId}/status`, { is_active: isActive })
    return res.data
  },

  async getRoles(): Promise<ApiResponse<RoleItem[]>> {
    const res = await apiClient.get<ApiResponse<RoleItem[]>>('/admin/roles')
    return res.data
  },
}
