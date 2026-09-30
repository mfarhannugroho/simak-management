import { apiClient } from './api'
import type { ApiResponse } from '@/types'

export interface LoginPayload {
  email: string
  password: string
}

export interface UserProfileDTO {
  id: string
  email: string
  full_name: string
  worker_id: string
  position: string
  phone: string
  company: string
  department: string
  roles: string[]
  permissions: string[]
}

export interface TokenPair {
  access_token: string
  refresh_token: string
  expires_at: string
  token_type: string
}

export interface LoginResponseData {
  tokens: TokenPair
  user: UserProfileDTO
}

export const authService = {
  async login(payload: LoginPayload): Promise<ApiResponse<LoginResponseData>> {
    const res = await apiClient.post<ApiResponse<LoginResponseData>>('/auth/login', payload)
    return res.data
  },

  async refreshToken(refreshToken: string): Promise<ApiResponse<TokenPair>> {
    const res = await apiClient.post<ApiResponse<TokenPair>>('/auth/refresh', {
      refresh_token: refreshToken,
    })
    return res.data
  },

  async logout(): Promise<ApiResponse<null>> {
    const res = await apiClient.post<ApiResponse<null>>('/auth/logout')
    return res.data
  },

  async getMe(): Promise<ApiResponse<UserProfileDTO>> {
    const res = await apiClient.get<ApiResponse<UserProfileDTO>>('/auth/me')
    return res.data
  },

  async changePassword(currentPassword: string, newPassword: string): Promise<ApiResponse<null>> {
    const res = await apiClient.put<ApiResponse<null>>('/auth/change-password', {
      current_password: currentPassword,
      new_password: newPassword,
    })
    return res.data
  },
}
