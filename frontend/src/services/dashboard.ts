import { apiClient } from './api'

export interface RoleMetric {
  label: string
  value: number
  description: string
  trend: string
  trend_type: 'positive' | 'negative' | 'neutral'
  highlight: boolean
}

export interface DashboardSummary {
  role: string
  user_name: string
  worker_id?: string
  position?: string
  department?: string
  company?: string
  metrics: RoleMetric[]
  total_simak: number
  active_simak: number
}

export interface DailyVolumeItem {
  day: string
  date: string
  masuk: number
  keluar: number
  total: number
  is_peak: boolean
}

export interface VolumeTrend {
  period: string
  total_volume: number
  peak_day: string
  daily: DailyVolumeItem[]
}

export interface ActivityItem {
  id: string
  type: string
  title: string
  description: string
  status: string
  variant: 'success' | 'danger' | 'warning' | 'neutral' | 'inka'
  timestamp: string
  action_url: string
}

export interface ActivitiesResponse {
  period: string
  activities: ActivityItem[]
}

export const dashboardService = {
  async getSummary(): Promise<DashboardSummary> {
    const res = await apiClient.get<{ success: boolean; data: DashboardSummary }>('/dashboard/summary')
    return res.data.data
  },

  async getVolumeTrend(): Promise<VolumeTrend> {
    const res = await apiClient.get<{ success: boolean; data: VolumeTrend }>('/dashboard/charts/volume-trend')
    return res.data.data
  },

  async getActivities(period: string = 'all'): Promise<ActivityItem[]> {
    const res = await apiClient.get<{ success: boolean; data: ActivitiesResponse }>(`/dashboard/activities`, {
      params: { period }
    })
    return res.data.data.activities || []
  }
}
