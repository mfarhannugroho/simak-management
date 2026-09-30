import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authService, type UserProfileDTO, type LoginPayload } from '@/services/auth'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<UserProfileDTO | null>(null)
  const accessToken = ref<string | null>(localStorage.getItem('simak_access_token'))
  const refreshToken = ref<string | null>(localStorage.getItem('simak_refresh_token'))
  const loading = ref(false)
  const error = ref<string | null>(null)

  const isAuthenticated = computed(() => !!accessToken.value && !!user.value)

  const roles = computed(() => user.value?.roles || [])
  const permissions = computed(() => user.value?.permissions || [])

  const primaryRole = computed(() => {
    if (!user.value || !user.value.roles.length) return 'Tamu'
    return user.value.roles[0]
  })

  function hasPermission(permission: string): boolean {
    if (!user.value) return false
    // Admin has access to all capabilities
    if (roles.value.includes('Admin')) return true
    return permissions.value.includes(permission) || permissions.value.includes('*')
  }

  function hasRole(role: string): boolean {
    if (!user.value) return false
    if (roles.value.includes('Admin')) return true
    return roles.value.includes(role)
  }

  function hasAnyRole(requiredRoles: string[]): boolean {
    if (!user.value) return false
    if (roles.value.includes('Admin')) return true
    return requiredRoles.some(r => roles.value.includes(r))
  }

  async function login(credentials: LoginPayload) {
    loading.value = true
    error.value = null
    try {
      const res = await authService.login(credentials)
      if (res.data) {
        accessToken.value = res.data.tokens.access_token
        refreshToken.value = res.data.tokens.refresh_token
        user.value = res.data.user

        localStorage.setItem('simak_access_token', res.data.tokens.access_token)
        localStorage.setItem('simak_refresh_token', res.data.tokens.refresh_token)
        localStorage.setItem('simak_user_cache', JSON.stringify(res.data.user))
        return true
      }
      return false
    } catch (err: any) {
      const msg = err.response?.data?.message || err.message || 'Login gagal, periksa email dan kata sandi'
      error.value = msg
      throw new Error(msg)
    } finally {
      loading.value = false
    }
  }

  async function logout() {
    try {
      if (accessToken.value) {
        await authService.logout().catch(() => {})
      }
    } finally {
      user.value = null
      accessToken.value = null
      refreshToken.value = null
      error.value = null
      localStorage.removeItem('simak_access_token')
      localStorage.removeItem('simak_refresh_token')
      localStorage.removeItem('simak_user_cache')
    }
  }

  async function fetchMe() {
    if (!accessToken.value) return null
    try {
      const res = await authService.getMe()
      if (res.data) {
        user.value = res.data
        localStorage.setItem('simak_user_cache', JSON.stringify(res.data))
        return res.data
      }
    } catch (err) {
      // If token expired/invalid, clear session
      await logout()
      return null
    }
  }

  // Restore session from localStorage cache on application bootstrap
  function initAuth() {
    const cachedUser = localStorage.getItem('simak_user_cache')
    if (cachedUser && accessToken.value) {
      try {
        user.value = JSON.parse(cachedUser)
      } catch {
        user.value = null
      }
    }
    if (accessToken.value) {
      fetchMe().catch(() => {})
    }
  }

  return {
    user,
    accessToken,
    refreshToken,
    loading,
    error,
    isAuthenticated,
    roles,
    permissions,
    primaryRole,
    hasPermission,
    hasRole,
    hasAnyRole,
    login,
    logout,
    fetchMe,
    initAuth,
  }
})
