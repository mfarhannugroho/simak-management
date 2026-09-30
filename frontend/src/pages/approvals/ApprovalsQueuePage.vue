<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { approvalService, type PendingApprovalItem } from '@/services/approval'
import DataTable, { type Column, type Pagination } from '@/components/ui/DataTable.vue'
import Button from '@/components/ui/Button.vue'
import Badge from '@/components/ui/Badge.vue'

const router = useRouter()

const items = ref<PendingApprovalItem[]>([])
const loading = ref(false)
const search = ref('')
const activeTab = ref<'pending' | 'history'>('pending')

const pagination = ref<Pagination>({
  page: 1,
  limit: 10,
  total_items: 0,
  total_pages: 1,
})

const columns: Column[] = [
  { key: 'simak_number', label: 'No. SIMAK / Pengajuan', sortable: true },
  { key: 'worker', label: 'Pemohon & Unit / Vendor' },
  { key: 'purpose', label: 'Keperluan & Area Pabrik' },
  { key: 'direction', label: 'Arah Aliran', align: 'center' },
  { key: 'step_progress', label: 'Tahap Verifikasi', align: 'center' },
  { key: 'total_items', label: 'Barang', align: 'center' },
  { key: 'actions', label: 'Aksi', align: 'right' },
]

async function fetchApprovals(page = 1) {
  loading.value = true
  try {
    const res = await approvalService.getPendingApprovals({
      page,
      limit: pagination.value.limit,
      search: search.value,
      status: activeTab.value === 'history' ? 'COMPLETED' : 'PENDING',
    })

    if (res.data) {
      items.value = res.data.items || []
      pagination.value = res.data.pagination
    }
  } catch (err) {
    console.error('Failed to fetch approval queue:', err)
  } finally {
    loading.value = false
  }
}

function handleTabChange(tab: 'pending' | 'history') {
  activeTab.value = tab
  pagination.value.page = 1
  fetchApprovals(1)
}

function handleSearch() {
  pagination.value.page = 1
  fetchApprovals(1)
}

function formatDate(dateStr: string) {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  })
}

function getDirectionBadge(direction: string) {
  switch (direction) {
    case 'MASUK':
      return { variant: 'inka' as const, label: 'Masuk Plant' }
    case 'KELUAR':
      return { variant: 'warning' as const, label: 'Keluar Plant' }
    case 'MASUK_KELUAR':
      return { variant: 'neutral' as const, label: 'Masuk-Keluar' }
    default:
      return { variant: 'neutral' as const, label: direction }
  }
}

onMounted(() => {
  fetchApprovals(1)
})
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
      <div>
        <div class="flex items-center gap-2">
          <span class="inline-block w-2.5 h-2.5 rounded-full bg-red-600 animate-pulse"></span>
          <p class="text-xs font-semibold uppercase tracking-wider text-red-600">Alur Persetujuan Bertingkat</p>
        </div>
        <h1 class="text-2xl font-bold text-slate-900 mt-1">Antrean Persetujuan SIMAK</h1>
        <p class="text-sm text-slate-500 mt-0.5">
          Tinjau dan proses permohonan surat ijin membawa alat kerja sesuai kewenangan Anda
        </p>
      </div>
    </div>

    <!-- Quick Stats Cards -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
      <div class="bg-white rounded-xl border border-red-100 p-4 shadow-sm flex items-center justify-between relative overflow-hidden">
        <div class="absolute left-0 top-0 bottom-0 w-1.5 bg-red-600"></div>
        <div>
          <p class="text-xs font-medium text-slate-500">Menunggu Tindakan Saya</p>
          <p class="text-2xl font-bold text-slate-900 mt-1">
            {{ activeTab === 'pending' ? pagination.total_items : '—' }}
          </p>
          <p class="text-xs text-red-600 font-medium mt-0.5">Memerlukan verifikasi & keputusan</p>
        </div>
        <div class="w-12 h-12 rounded-xl bg-red-50 text-red-600 flex items-center justify-center">
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
        </div>
      </div>

      <div class="bg-white rounded-xl border border-slate-200 p-4 shadow-sm flex items-center justify-between">
        <div>
          <p class="text-xs font-medium text-slate-500">Prinsip Keamanan</p>
          <p class="text-sm font-bold text-slate-900 mt-1">Segregation of Duties (SoD)</p>
          <p class="text-xs text-slate-500 mt-0.5">Pemohon dilarang menyetujui izin sendiri</p>
        </div>
        <div class="w-12 h-12 rounded-xl bg-slate-100 text-slate-700 flex items-center justify-center">
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
          </svg>
        </div>
      </div>

      <div class="bg-white rounded-xl border border-slate-200 p-4 shadow-sm flex items-center justify-between">
        <div>
          <p class="text-xs font-medium text-slate-500">Protokol Area Pabrik</p>
          <p class="text-sm font-bold text-slate-900 mt-1">Standar PT INKA (Persero)</p>
          <p class="text-xs text-slate-500 mt-0.5">Pemeriksaan fisik oleh Gate Officer di pos</p>
        </div>
        <div class="w-12 h-12 rounded-xl bg-amber-50 text-amber-600 flex items-center justify-center">
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
          </svg>
        </div>
      </div>
    </div>

    <!-- Main Card -->
    <div class="bg-white rounded-xl border border-slate-200 shadow-sm overflow-hidden">
      <!-- Tabs & Search Toolbar -->
      <div class="p-4 border-b border-slate-200 flex flex-col md:flex-row md:items-center md:justify-between gap-4">
        <!-- Tabs -->
        <div class="flex items-center gap-1 bg-slate-100 p-1 rounded-lg">
          <button
            type="button"
            class="px-4 py-1.5 text-xs font-medium rounded-md transition-colors"
            :class="activeTab === 'pending' ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-600 hover:text-slate-900'"
            @click="handleTabChange('pending')"
          >
            Menunggu Persetujuan
          </button>
          <button
            type="button"
            class="px-4 py-1.5 text-xs font-medium rounded-md transition-colors"
            :class="activeTab === 'history' ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-600 hover:text-slate-900'"
            @click="handleTabChange('history')"
          >
            Riwayat Keputusan Saya
          </button>
        </div>

        <!-- Search Bar -->
        <div class="relative w-full md:w-80">
          <input
            v-model="search"
            type="text"
            placeholder="Cari no. SIMAK, pemohon, keperluan..."
            class="w-full pl-9 pr-4 py-2 bg-slate-50 border border-slate-200 rounded-lg text-xs focus:outline-none focus:ring-2 focus:ring-red-500/20 focus:border-red-500 transition-all"
            @keyup.enter="handleSearch"
          />
          <svg class="w-4 h-4 text-slate-400 absolute left-3 top-2.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
        </div>
      </div>

      <!-- Table Component -->
      <DataTable
        :columns="columns"
        :items="items"
        :loading="loading"
        :pagination="pagination"
        hoverable
        @page-change="fetchApprovals"
      >
        <!-- Custom column: SIMAK Number -->
        <template #cell-simak_number="{ row }">
          <div class="space-y-1">
            <div class="flex items-center gap-1.5">
              <span class="font-mono text-xs font-semibold text-slate-900">{{ row.simak_number || 'DRAFT' }}</span>
              <span
                v-if="row.simak_status === 'UNDER_REVIEW'"
                class="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-semibold bg-amber-50 text-amber-700 border border-amber-200"
              >
                Review
              </span>
              <span
                v-else-if="row.simak_status === 'SUBMITTED'"
                class="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-semibold bg-blue-50 text-blue-700 border border-blue-200"
              >
                Baru
              </span>
            </div>
            <p class="text-[11px] text-slate-400">Diajukan: {{ formatDate(row.submitted_at) }}</p>
          </div>
        </template>

        <!-- Custom column: Worker -->
        <template #cell-worker="{ row }">
          <div class="space-y-0.5">
            <p class="text-xs font-medium text-slate-900">{{ row.worker?.full_name || 'Tidak diketahui' }}</p>
            <p class="text-[11px] text-slate-500">
              {{ row.worker?.company?.name || 'PT INKA (Persero)' }}
              <span v-if="row.worker?.department"> • {{ row.worker.department.code }}</span>
            </p>
          </div>
        </template>

        <!-- Custom column: Purpose -->
        <template #cell-purpose="{ row }">
          <div class="max-w-xs space-y-0.5">
            <p class="text-xs text-slate-800 line-clamp-1 font-medium">{{ row.purpose }}</p>
            <p class="text-[11px] text-slate-500 flex items-center gap-1">
              <svg class="w-3 h-3 text-slate-400 inline" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
              </svg>
              {{ row.destination_area }}
            </p>
          </div>
        </template>

        <!-- Custom column: Direction -->
        <template #cell-direction="{ row }">
          <Badge
            :variant="getDirectionBadge(row.direction).variant"
            size="sm"
          >
            {{ getDirectionBadge(row.direction).label }}
          </Badge>
        </template>

        <!-- Custom column: Step Progress -->
        <template #cell-step_progress="{ row }">
          <div class="flex flex-col items-center gap-1">
            <span class="inline-flex items-center px-2 py-0.5 rounded-full text-[11px] font-semibold bg-red-50 text-red-700 border border-red-200">
              Tahap {{ row.current_step }} dari {{ row.total_steps }}
            </span>
            <span class="text-[10px] text-slate-500 text-center max-w-[140px] truncate" :title="row.step_name">
              {{ row.step_name }}
            </span>
          </div>
        </template>

        <!-- Custom column: Items Count -->
        <template #cell-total_items="{ row }">
          <div class="text-center">
            <span class="inline-flex items-center px-2 py-0.5 rounded bg-slate-100 text-slate-700 font-semibold text-xs">
              {{ row.total_items }} Item
            </span>
          </div>
        </template>

        <!-- Custom column: Actions -->
        <template #cell-actions="{ row }">
          <div class="flex items-center justify-end">
            <Button
              variant="primary"
              size="sm"
              class="gap-1.5 shadow-sm"
              @click="router.push(`/approvals/${row.simak_id}`)"
            >
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4" />
              </svg>
              <span>Tinjau & Putuskan</span>
            </Button>
          </div>
        </template>

        <!-- Empty State -->
        <template #empty>
          <div class="py-12 text-center">
            <div class="w-16 h-16 mx-auto mb-3 rounded-full bg-slate-50 border border-slate-200 flex items-center justify-center text-slate-400">
              <svg class="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
            <p class="text-sm font-semibold text-slate-700">Tidak ada permohonan SIMAK yang menunggu</p>
            <p class="text-xs text-slate-400 mt-1 max-w-sm mx-auto">
              Semua pengajuan yang membutuhkan persetujuan Anda telah diproses. Permohonan baru akan muncul di sini secara otomatis.
            </p>
          </div>
        </template>
      </DataTable>
    </div>
  </div>
</template>
