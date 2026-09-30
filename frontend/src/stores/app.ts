import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { HealthCheckData, UserProfile } from '@/types'
import { healthService } from '@/services/api'

export const useAppStore = defineStore('app', () => {
  // Sidebar state
  const sidebarOpen = ref(true)
  const isMobile = ref(false)

  // Current active user (Sample active profile for phase 1 demo)
  const currentUser = ref<UserProfile>({
    id: 'usr-101',
    nip: 'INKA-2024-8891',
    name: 'Budi Santoso, S.T.',
    email: 'budi.santoso@inka.co.id',
    role: 'Supervisor',
    department: 'Divisi Rekayasa & Fabrikasi Kereta',
    company: 'PT INKA (Persero) - Madiun',
    avatarUrl: 'https://images.unsplash.com/photo-1534528741775-53994a69daeb?auto=format&fit=crop&q=80&w=256',
  })

  // System Health status
  const systemHealth = ref<HealthCheckData | null>(null)
  const healthLoading = ref(false)
  const healthError = ref<string | null>(null)
  const lastHealthCheck = ref<string | null>(null)

  function toggleSidebar() {
    sidebarOpen.value = !sidebarOpen.value
  }

  function setSidebar(open: boolean) {
    sidebarOpen.value = open
  }

  async function fetchHealth() {
    healthLoading.value = true
    healthError.value = null
    try {
      const res = await healthService.checkHealth()
      if (res.data) {
        systemHealth.value = res.data
        lastHealthCheck.value = new Date().toLocaleTimeString('id-ID')
      }
    } catch (err: any) {
      healthError.value = err.message || 'Koneksi ke backend gagal'
      systemHealth.value = null
    } finally {
      healthLoading.value = false
    }
  }

  return {
    sidebarOpen,
    isMobile,
    currentUser,
    systemHealth,
    healthLoading,
    healthError,
    lastHealthCheck,
    toggleSidebar,
    setSidebar,
    fetchHealth,
  }
})
