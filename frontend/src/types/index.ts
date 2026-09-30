// Standard API response envelope
export interface ApiResponse<T = any> {
  success: boolean
  message: string
  data?: T
  errors?: Record<string, string[]> | string
}

// Health Check Response
export interface HealthServiceStatus {
  database: string
  redis: string
}

export interface HealthCheckData {
  status: 'healthy' | 'unhealthy'
  version: string
  timestamp: string
  services: HealthServiceStatus
}

// User & Role Types (for RBAC and Display)
export type RoleName = 
  | 'Applicant'
  | 'Approver'
  | 'SIMAK Officer'
  | 'Gate Officer'
  | 'Supervisor'
  | 'Admin'

export interface UserProfile {
  id: string
  nip: string
  name: string
  email: string
  role: RoleName
  department: string
  company: string
  avatarUrl?: string
}

// Navigation Item
export interface NavItem {
  name: string
  path: string
  icon: string
  badge?: string | number
  badgeColor?: string
  requiredRole?: RoleName[]
}

export interface NavGroup {
  title: string
  items: NavItem[]
}

// SIMAK Lifecycle Status
export type SimakStatus = 
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

// Gate Inspection Result
export type InspectionResult = 
  | 'MATCH'
  | 'MISMATCH'
  | 'INCOMPLETE'
  | 'ADDITIONAL_ITEM'
