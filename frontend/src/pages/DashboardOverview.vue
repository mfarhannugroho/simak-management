<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'
import StatCard from '@/components/ui/StatCard.vue'
import Button from '@/components/ui/Button.vue'
import Badge from '@/components/ui/Badge.vue'
import PermissionGate from '@/components/auth/PermissionGate.vue'
import { dashboardService, type DashboardSummary, type VolumeTrend, type ActivityItem } from '@/services/dashboard'

const router = useRouter()
const appStore = useAppStore()
const authStore = useAuthStore()

const loading = ref(true)
const summary = ref<DashboardSummary | null>(null)
const volumeTrend = ref<VolumeTrend | null>(null)
const activities = ref<ActivityItem[]>([])
const activePeriod = ref<'today' | '7d' | '30d' | 'all'>('7d')
const activityLoading = ref(false)

async function loadDashboardData() {
  loading.value = true
  try {
    const [sumData, trendData, actData] = await Promise.all([
      dashboardService.getSummary().catch(() => null),
      dashboardService.getVolumeTrend().catch(() => null),
      dashboardService.getActivities(activePeriod.value).catch(() => []),
    ])
    summary.value = sumData
    volumeTrend.value = trendData
    activities.value = actData
  } finally {
    loading.value = false
  }
}

async function loadActivities() {
  activityLoading.value = true
  try {
    activities.value = await dashboardService.getActivities(activePeriod.value)
  } catch (err) {
    console.error('Failed to load activities', err)
  } finally {
    activityLoading.value = false
  }
}

watch(activePeriod, () => {
  loadActivities()
})

onMounted(() => {
  appStore.fetchHealth()
  loadDashboardData()
})

const maxVolume = computed(() => {
  if (!volumeTrend.value || !volumeTrend.value.daily.length) return 10
  const max = Math.max(...volumeTrend.value.daily.map(d => d.total))
  return max > 0 ? max : 10
})

function formatTime(isoString: string) {
  try {
    const d = new Date(isoString)
    return d.toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' }) + ' WIB'
  } catch {
    return '-'
  }
}
</script>

<template>
  <div class="space-y-6">
    <!-- Dashboard Header: Welcome Banner & Quick Actions -->
    <div class="bg-white p-5 rounded-xl border border-slate-200 shadow-card flex flex-col md:flex-row md:items-center justify-between gap-4">
      <div class="flex items-center gap-4">
        <div class="w-14 h-14 rounded-2xl bg-inka-700 text-white flex items-center justify-center font-bold text-xl shadow-md shrink-0">
          {{ (summary?.user_name || authStore.user?.full_name || 'U').charAt(0) }}
        </div>
        <div class="min-w-0">
          <div class="flex flex-wrap items-center gap-2">
            <h1 class="text-xl font-bold text-slate-900 truncate">
              {{ summary?.user_name || authStore.user?.full_name || 'Pengguna SIMAK' }}
            </h1>
            <Badge variant="inka" size="sm">
              {{ summary?.role || authStore.primaryRole }}
            </Badge>
            <span v-if="summary?.worker_id" class="text-xs text-slate-400 font-mono">
              {{ summary.worker_id }}
            </span>
          </div>
          <p class="text-xs text-slate-500 mt-1">
            {{ summary?.position || authStore.user?.position || 'Staf Operasional' }} · {{ summary?.department || authStore.user?.department || 'PT Industri Kereta Api (Persero)' }}
          </p>
          <div class="flex items-center gap-2 mt-1.5">
            <span class="inline-flex items-center gap-1.5 text-xs text-slate-600 font-medium">
              <svg class="w-3.5 h-3.5 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
              </svg>
              {{ summary?.company || authStore.user?.company || 'PT Industri Kereta Api (Persero)' }}
            </span>
          </div>
        </div>
      </div>

      <!-- Quick Action Buttons Protected by Granular RBAC Permissions -->
      <div class="flex flex-wrap items-center gap-2 shrink-0">
        <!-- Applicant capability -->
        <PermissionGate permission="simak.create">
          <Button variant="primary" size="sm" @click="router.push('/simak/new')">
            <svg class="w-3.5 h-3.5 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
            </svg>
            Ajukan SIMAK
          </Button>
        </PermissionGate>

        <!-- Approver capability -->
        <PermissionGate permission="simak.approve">
          <Button variant="outline" size="sm" @click="router.push('/approvals')">
            <svg class="w-3.5 h-3.5 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            Review Antrean
          </Button>
        </PermissionGate>

        <!-- Gate Officer capability -->
        <PermissionGate permission="inspection.perform">
          <Button variant="primary" size="sm" @click="router.push('/gate')">
            <svg class="w-3.5 h-3.5 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v1m6 11h2m-6 0h-2v4m0-11v3m0 0h.01M12 12h4.01M16 20h4M4 12h4m12 0h.01M5 8h2a1 1 0 001-1V5a1 1 0 00-1-1H5a1 1 0 00-1 1v2a1 1 0 001 1zm12 0h2a1 1 0 001-1V5a1 1 0 00-1-1h-2a1 1 0 00-1 1v2a1 1 0 001 1zM5 20h2a1 1 0 001-1v-2a1 1 0 00-1-1H5a1 1 0 00-1 1v2a1 1 0 001 1z" />
            </svg>
            Scan QR Gate
          </Button>
        </PermissionGate>

        <!-- Reports Button -->
        <Button variant="outline" size="sm" @click="router.push('/reports')">
          <svg class="w-3.5 h-3.5 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 17v-2m3 2v-4m3 4v-6m2 10H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
          </svg>
          Laporan
        </Button>
      </div>
    </div>

    <!-- Pattern 1: 3-Card Role-Aware KPI Row (design.md §4A) -->
    <div v-if="summary && summary.metrics.length" class="grid grid-cols-1 md:grid-cols-3 gap-5">
      <StatCard
        v-for="(metric, idx) in summary.metrics"
        :key="idx"
        :label="metric.label.toUpperCase()"
        :value="metric.value"
        :trend="metric.trend"
        :trendType="metric.trend_type"
        :description="metric.description"
        :highlight="metric.highlight"
      >
        <template #icon>
          <div
            class="w-9 h-9 rounded-lg flex items-center justify-center"
            :class="metric.highlight ? 'bg-rose-50 text-rose-600' : 'bg-slate-100 text-slate-600'"
          >
            <svg v-if="idx === 0" class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <svg v-else-if="idx === 1" class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
            </svg>
            <svg v-else class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
            </svg>
          </div>
        </template>
      </StatCard>
    </div>

    <!-- Fallback skeleton if loading -->
    <div v-else-if="loading" class="grid grid-cols-1 md:grid-cols-3 gap-5">
      <div v-for="i in 3" :key="i" class="h-32 bg-slate-100 rounded-xl animate-pulse"></div>
    </div>

    <!-- Pattern 2 & 3: Middle Section (Volume Trend Bar Chart & Recent Activities Feed) -->
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <!-- Volume Pergerakan Barang Bawaan (2 Columns) -->
      <div class="lg:col-span-2 bg-white rounded-xl border border-slate-200 p-5 shadow-card flex flex-col justify-between">
        <div>
          <div class="flex items-center justify-between pb-4 border-b border-slate-100">
            <div>
              <h2 class="text-sm font-bold text-slate-900">
                Volume Pergerakan Alat Bawaan (7 Hari Terakhir)
              </h2>
              <p class="text-xs text-slate-500">
                Pemeriksaan fisik pos gerbang dengan sorotan hari puncak pergerakan
              </p>
            </div>
            <div v-if="volumeTrend" class="flex items-center gap-1.5 text-xs text-slate-600 bg-slate-50 px-2.5 py-1 rounded-lg border border-slate-200">
              <span class="w-2.5 h-2.5 rounded bg-inka-700"></span>
              <span>Puncak ({{ volumeTrend.peak_day }})</span>
            </div>
          </div>

          <!-- Visual Bar Chart with Merah INKA Peak Highlight (#D32F2F) -->
          <div class="pt-6 pb-2">
            <div v-if="volumeTrend && volumeTrend.daily.length" class="relative">
              <!-- Background reference grid lines for scale -->
              <div class="absolute inset-0 h-44 flex flex-col justify-between pointer-events-none pr-2">
                <div class="w-full border-b border-dashed border-slate-200"></div>
                <div class="w-full border-b border-dashed border-slate-100"></div>
                <div class="w-full border-b border-dashed border-slate-100"></div>
                <div class="w-full border-b border-slate-200"></div>
              </div>

              <!-- Bar Columns Row with explicit height -->
              <div class="h-44 flex items-end justify-between gap-3 sm:gap-4 px-2 sm:px-4 relative z-10">
                <div
                  v-for="item in volumeTrend.daily"
                  :key="item.date"
                  class="flex-1 h-full flex flex-col justify-end items-center group relative cursor-pointer"
                >
                  <!-- Tooltip on hover showing Masuk & Keluar -->
                  <div class="opacity-0 group-hover:opacity-100 transition-all duration-200 absolute -top-12 px-3 py-1.5 rounded-lg bg-slate-900 text-white text-[11px] font-medium pointer-events-none whitespace-nowrap z-30 shadow-xl text-center">
                    <p class="font-bold text-white">{{ item.total }} Gerakan</p>
                    <p class="text-slate-300 text-[10px]">Masuk: {{ item.masuk }} · Keluar: {{ item.keluar }}</p>
                    <div class="w-2 h-2 bg-slate-900 rotate-45 mx-auto -mb-1 mt-0.5"></div>
                  </div>

                  <!-- Count label above bar -->
                  <span
                    class="text-[11px] font-bold mb-1.5 transition-colors"
                    :class="item.is_peak ? 'text-inka-700 font-extrabold' : 'text-slate-500 group-hover:text-slate-800'"
                  >
                    {{ item.total }}
                  </span>

                  <!-- Bar Track & Bar Column with explicit pixel height -->
                  <div class="w-full max-w-[44px] flex items-end justify-center">
                    <div
                      class="w-full rounded-t-lg transition-all duration-300"
                      :style="{ height: `${Math.max(16, Math.round((item.total / maxVolume) * 115))}px` }"
                      :class="item.is_peak 
                        ? 'bg-gradient-to-t from-inka-700 via-rose-600 to-rose-500 shadow-md shadow-rose-600/30 ring-2 ring-rose-300' 
                        : 'bg-gradient-to-t from-slate-300 to-slate-200 group-hover:from-slate-400 group-hover:to-slate-300 shadow-xs'"
                    ></div>
                  </div>

                  <!-- Day label -->
                  <div class="mt-2 text-center">
                    <span
                      class="text-xs block font-bold transition-colors"
                      :class="item.is_peak ? 'text-inka-700' : 'text-slate-600 group-hover:text-slate-900'"
                    >
                      {{ item.day }}
                    </span>
                  </div>
                </div>
              </div>
            </div>

            <div v-else class="h-44 flex items-center justify-center text-xs text-slate-400">
              Belum ada data pergerakan tercatat dalam periode ini.
            </div>
          </div>
        </div>

        <div class="mt-4 pt-3 border-t border-slate-100 flex items-center justify-between text-xs text-slate-500">
          <span>Total Lalu Lintas: <strong class="text-slate-900">{{ volumeTrend?.total_volume || 0 }} pergerakan</strong></span>
          <span class="text-inka-700 font-semibold flex items-center gap-1 cursor-pointer hover:underline" @click="router.push('/reports')">
            Buka Analisis Laporan Lengkap →
          </span>
        </div>
      </div>

      <!-- Pattern 3: Recent Activity Stream (1 Column) -->
      <div class="bg-white rounded-xl border border-slate-200 p-5 shadow-card flex flex-col">
        <div class="flex items-center justify-between pb-3 border-b border-slate-100">
          <h2 class="text-sm font-bold text-slate-900">
            Aktivitas Terkini
          </h2>
          <!-- Filter Pills -->
          <div class="flex items-center gap-0.5 bg-slate-100 p-0.5 rounded-lg text-[10px] font-medium">
            <button
              v-for="p in [
                { id: 'today', label: 'Hari Ini' },
                { id: '7d', label: '7 Hari' },
                { id: 'all', label: 'Semua' }
              ]"
              :key="p.id"
              @click="activePeriod = p.id as any"
              :class="activePeriod === p.id ? 'bg-white text-slate-900 shadow-sm font-bold' : 'text-slate-600 hover:text-slate-900'"
              class="px-2 py-0.5 rounded transition-all cursor-pointer"
            >
              {{ p.label }}
            </button>
          </div>
        </div>

        <!-- Activity Items -->
        <div class="mt-3 space-y-2.5 flex-1 overflow-y-auto max-h-[280px] pr-1">
          <div
            v-for="act in activities"
            :key="act.id"
            @click="act.action_url ? router.push(act.action_url) : null"
            class="p-2.5 rounded-lg border border-slate-100 hover:border-slate-300 transition-colors bg-slate-50/50 cursor-pointer"
          >
            <div class="flex items-center justify-between gap-2">
              <span class="text-[11px] font-bold text-slate-900 truncate">
                {{ act.title }}
              </span>
              <Badge :variant="act.variant" size="sm">
                {{ act.status }}
              </Badge>
            </div>
            <p class="text-[11px] text-slate-600 mt-1 line-clamp-2">
              {{ act.description }}
            </p>
            <span class="text-[10px] text-slate-400 block mt-1">
              {{ formatTime(act.timestamp) }}
            </span>
          </div>

          <div v-if="!activities.length && !activityLoading" class="py-8 text-center text-xs text-slate-400">
            Tidak ada aktivitas pada periode ini.
          </div>

          <div v-if="activityLoading" class="space-y-2">
            <div v-for="i in 3" :key="i" class="h-14 bg-slate-100 rounded-lg animate-pulse"></div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
