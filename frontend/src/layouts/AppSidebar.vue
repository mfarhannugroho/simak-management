<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import Badge from '@/components/ui/Badge.vue'
import type { NavGroup } from '@/types'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

interface ExtendedNavItem {
  name: string
  path: string
  icon: string
  badge?: string | number
  badgeColor?: string
  requiredRoles?: string[]
  requiredPermission?: string
}

interface ExtendedNavGroup {
  title: string
  items: ExtendedNavItem[]
}

const navGroups: ExtendedNavGroup[] = [
  {
    title: 'NAVIGASI UTAMA',
    items: [
      { name: 'Dashboard Overview', path: '/', icon: 'dashboard' },
      {
        name: 'Barang Rutin Saya',
        path: '/profile/my-items',
        icon: 'tool',
        badge: 'Rutin',
        badgeColor: 'bg-emerald-100 text-emerald-800',
        requiredRoles: ['Applicant', 'Admin'],
      },
      {
        name: 'Permohonan SIMAK',
        path: '/simak',
        icon: 'file-text',
        badge: 'Draft',
        requiredRoles: ['Applicant', 'Admin'],
        requiredPermission: 'simak.view',
      },
      {
        name: 'Persetujuan (Approvals)',
        path: '/approvals',
        icon: 'check-square',
        badge: 3,
        badgeColor: 'bg-amber-100 text-amber-800',
        requiredRoles: ['Approver', 'Supervisor', 'Admin'],
        requiredPermission: 'simak.approve',
      },
      {
        name: 'Penerbitan SIMAK',
        path: '/officer/issuance',
        icon: 'award',
        requiredRoles: ['SIMAK Officer', 'Admin'],
        requiredPermission: 'simak.issue',
      },
    ],
  },
  {
    title: 'OPERASIONAL POS GERBANG',
    items: [
      {
        name: 'Pemeriksaan Gate',
        path: '/gate',
        icon: 'scan',
        requiredRoles: ['Gate Officer', 'Admin'],
        requiredPermission: 'inspection.perform',
      },
      {
        name: 'Riwayat Verifikasi',
        path: '/gate/history',
        icon: 'history',
        requiredRoles: ['Gate Officer', 'Supervisor', 'Admin'],
        requiredPermission: 'inspection.view',
      },
    ],
  },
  {
    title: 'LAPORAN & PENGAWASAN',
    items: [
      {
        name: 'Discrepancy Log',
        path: '/discrepancies',
        icon: 'alert-triangle',
        badge: 'Alert',
        badgeColor: 'bg-rose-100 text-rose-800',
        requiredRoles: ['Supervisor', 'Admin'],
        requiredPermission: 'inspection.resolve_discrepancy',
      },
      {
        name: 'Audit Trail',
        path: '/audit',
        icon: 'shield',
        requiredRoles: ['Admin', 'Supervisor'],
        requiredPermission: 'audit.view',
      },
      {
        name: 'Rekap & Analitik',
        path: '/reports',
        icon: 'bar-chart',
        requiredRoles: ['Admin', 'Supervisor', 'SIMAK Officer', 'Approver', 'Gate Officer'],
      },
    ],
  },
  {
    title: 'PENGATURAN SISTEM',
    items: [
      {
        name: 'Master Organisasi',
        path: '/admin/master-data',
        icon: 'database',
        requiredRoles: ['Admin'],
      },
      {
        name: 'Data Pekerja & Profil',
        path: '/admin/workers',
        icon: 'users',
        requiredRoles: ['Admin'],
      },
      {
        name: 'Kategori Alat Kerja',
        path: '/admin/item-categories',
        icon: 'category',
        requiredRoles: ['Admin'],
      },
      {
        name: 'Manajemen Pengguna',
        path: '/admin/users',
        icon: 'user-check',
        requiredRoles: ['Admin'],
      },
      {
        name: 'Approval Matrix',
        path: '/admin/approval-matrix',
        icon: 'sliders',
        requiredRoles: ['Admin'],
      },
    ],
  },
]

// Filter nav groups and items based on active user's roles and permissions
const filteredNavGroups = computed(() => {
  return navGroups
    .map(group => {
      const items = group.items.filter(item => {
        if (!authStore.user) return false
        if (authStore.roles.includes('Admin')) return true

        if (item.requiredRoles && item.requiredRoles.length) {
          if (!authStore.hasAnyRole(item.requiredRoles)) return false
        }

        if (item.requiredPermission) {
          if (!authStore.hasPermission(item.requiredPermission)) return false
        }

        return true
      })

      return {
        title: group.title,
        items,
      }
    })
    .filter(group => group.items.length > 0)
})

function isActive(path: string) {
  if (route.path === path) return true
  if (path === '/') return route.path === '/'

  const allPaths = navGroups.flatMap(g => g.items.map(i => i.path))
  const matching = allPaths.filter(p => route.path === p || (p !== '/' && route.path.startsWith(p + '/')))
  matching.sort((a, b) => b.length - a.length)
  return matching[0] === path
}

function navigateTo(path: string) {
  router.push(path)
}

async function handleLogout() {
  if (confirm('Apakah Anda yakin ingin keluar dari sistem SIMAK?')) {
    await authStore.logout()
    router.push('/login')
  }
}
</script>

<template>
  <aside
    class="w-64 bg-white border-r border-slate-200 flex flex-col h-screen select-none shrink-0 transition-all duration-300 z-30"
  >
    <!-- Brand Header -->
    <div class="h-16 px-4 border-b border-slate-200 flex items-center gap-3 bg-white">
      <div class="h-10 px-2 rounded-lg bg-white flex items-center justify-center border border-slate-200 shadow-xs shrink-0">
        <img src="/logo.png" alt="PT INKA" class="h-6 w-auto object-contain max-w-[48px]" />
      </div>
      <div class="min-w-0">
        <h1 class="text-sm font-bold text-slate-900 truncate leading-tight">
          PT INKA (Persero)
        </h1>
        <p class="text-[11px] font-semibold text-inka-700 uppercase tracking-wider truncate">
          SIMAK Digital System
        </p>
      </div>
    </div>

    <!-- Navigation Menu Items (Role Filtered) -->
    <div class="flex-1 overflow-y-auto px-3 py-4 space-y-6">
      <div v-for="group in filteredNavGroups" :key="group.title" class="space-y-1">
        <p class="px-3 text-[11px] font-bold text-slate-400 uppercase tracking-wider mb-2">
          {{ group.title }}
        </p>
        <div class="space-y-0.5">
          <button
            v-for="item in group.items"
            :key="item.path"
            @click="navigateTo(item.path)"
            :class="[
              'w-full flex items-center justify-between px-3 py-2 text-xs font-medium rounded-lg transition-all text-left cursor-pointer group',
              isActive(item.path)
                ? 'bg-inka-50 text-inka-700 font-semibold border-r-2 border-inka-700'
                : 'text-slate-600 hover:text-slate-900 hover:bg-slate-50',
            ]"
          >
            <div class="flex items-center gap-2.5 min-w-0">
              <span
                :class="[
                  'w-4 h-4 shrink-0 transition-colors',
                  isActive(item.path) ? 'text-inka-700' : 'text-slate-400 group-hover:text-slate-600',
                ]"
              >
                <!-- Dashboard -->
                <svg v-if="item.icon === 'dashboard'" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z" /></svg>
                <!-- File Text -->
                <svg v-else-if="item.icon === 'file-text'" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" /></svg>
                <!-- Check Square -->
                <svg v-else-if="item.icon === 'check-square'" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4" /></svg>
                <!-- Award -->
                <svg v-else-if="item.icon === 'award'" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><circle cx="12" cy="8" r="7" stroke-width="2"/><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.21 13.89L7 23l5-3 5 3-1.21-9.12"/></svg>
                <!-- Scan -->
                <svg v-else-if="item.icon === 'scan'" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v1m6 11h2m-6 0h-2v4m0-11v3m0 0h.01M12 12h4.01M16 20h4M4 12h4m12 0h.01M5 8h2a1 1 0 001-1V5a1 1 0 00-1-1H5a1 1 0 00-1 1v2a1 1 0 001 1zm12 0h2a1 1 0 001-1V5a1 1 0 00-1-1h-2a1 1 0 00-1 1v2a1 1 0 001 1zM5 20h2a1 1 0 001-1v-2a1 1 0 00-1-1H5a1 1 0 00-1 1v2a1 1 0 001 1z" /></svg>
                <!-- History -->
                <svg v-else-if="item.icon === 'history'" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
                <!-- Alert Triangle -->
                <svg v-else-if="item.icon === 'alert-triangle'" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" /></svg>
                <!-- Shield -->
                <svg v-else-if="item.icon === 'shield'" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" /></svg>
                <!-- Bar Chart -->
                <svg v-else-if="item.icon === 'bar-chart'" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" /></svg>
                <!-- Tool / Routine Equipment -->
                <svg v-else-if="item.icon === 'tool'" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" /></svg>
                <!-- Database -->
                <svg v-else-if="item.icon === 'database'" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><ellipse cx="12" cy="5" rx="9" ry="3" stroke-width="2"/><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12c0 1.66-4 3-9 3s-9-1.34-9-3"/><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 5v14c0 1.66 4 3 9 3s9-1.34 9-3V5"/></svg>
                <!-- Users / Workers -->
                <svg v-else-if="item.icon === 'users'" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" /></svg>
                <!-- Category -->
                <svg v-else-if="item.icon === 'category'" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z" /></svg>
                <!-- User Check -->
                <svg v-else-if="item.icon === 'user-check'" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
                <!-- Default / Sliders -->
                <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6V4m0 2a2 2 0 100 4m0-4a2 2 0 110 4m-6 8a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4m6 6v10m6-2a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4" /></svg>
              </span>
              <span class="truncate">{{ item.name }}</span>
            </div>
            <span
              v-if="item.badge"
              :class="[
                'text-[10px] font-semibold px-1.5 py-0.5 rounded-full shrink-0',
                item.badgeColor || 'bg-slate-100 text-slate-600',
              ]"
            >
              {{ item.badge }}
            </span>
          </button>
        </div>
      </div>
    </div>

    <!-- User Profile Bar (Bottom of Sidebar) with Logout -->
    <div class="p-3 border-t border-slate-200 bg-slate-50/50 space-y-2">
      <div class="flex items-center justify-between p-2 rounded-lg bg-white border border-slate-200 shadow-sm">
        <div class="flex items-center gap-2.5 min-w-0">
          <div class="w-8 h-8 rounded-full bg-inka-700 text-white flex items-center justify-center font-bold text-xs shrink-0 shadow-sm">
            {{ (authStore.user?.full_name || 'U').charAt(0) }}
          </div>
          <div class="min-w-0 flex-1">
            <p class="text-xs font-bold text-slate-900 truncate">
              {{ authStore.user?.full_name || 'Pengguna SIMAK' }}
            </p>
            <div class="flex items-center gap-1.5 mt-0.5">
              <span class="inline-block text-[10px] font-semibold text-inka-700 uppercase tracking-wide truncate">
                {{ authStore.primaryRole }}
              </span>
              <span class="text-[10px] text-slate-400">·</span>
              <span class="text-[10px] text-slate-500 truncate">
                {{ authStore.user?.worker_id || 'ID' }}
              </span>
            </div>
          </div>
        </div>

        <!-- Logout Button -->
        <button
          @click="handleLogout"
          class="p-1.5 rounded-md text-slate-400 hover:text-rose-600 hover:bg-rose-50 transition-colors cursor-pointer shrink-0"
          title="Keluar (Logout)"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
          </svg>
        </button>
      </div>
    </div>
  </aside>
</template>
