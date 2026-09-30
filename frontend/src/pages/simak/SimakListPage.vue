<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { simakService, type Simak } from '@/services/simak'
import DataTable, { type Column, type Pagination } from '@/components/ui/DataTable.vue'
import Button from '@/components/ui/Button.vue'
import Badge from '@/components/ui/Badge.vue'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const simaks = ref<Simak[]>([])
const loading = ref(false)
const search = ref('')
const activeTab = ref<'all' | 'draft' | 'pending' | 'active' | 'history'>('all')
const directionFilter = ref('ALL')

const pagination = ref<Pagination>({
  page: 1,
  limit: 10,
  total_items: 0,
  total_pages: 1,
})

const columns: Column[] = [
  { key: 'simak_number', label: 'No. SIMAK / Pengajuan', sortable: true },
  { key: 'purpose', label: 'Keperluan & Area Workshop' },
  { key: 'direction', label: 'Arah Aliran', align: 'center' },
  { key: 'validity', label: 'Masa Berlaku', align: 'center' },
  { key: 'items_count', label: 'Barang', align: 'center' },
  { key: 'status', label: 'Status Pengajuan', align: 'center' },
  { key: 'actions', label: 'Aksi', align: 'right' },
]

// Tab mappings to backend status query
const tabStatusMap: Record<string, string> = {
  all: 'ALL',
  draft: 'DRAFT',
  pending: 'SUBMITTED', // will filter in UI or query
  active: 'APPROVED',
  history: 'COMPLETED',
}

async function fetchSimaks(page = 1) {
  loading.value = true
  try {
    let statusQuery = tabStatusMap[activeTab.value]
    if (activeTab.value === 'all') statusQuery = ''

    const res = await simakService.getSimaks({
      page,
      limit: pagination.value.limit,
      search: search.value,
      status: statusQuery,
      direction: directionFilter.value !== 'ALL' ? directionFilter.value : undefined,
    })

    if (res.data) {
      simaks.value = res.data.items
      pagination.value = res.data.pagination
    }
  } catch (err: any) {
    console.error('Failed to load SIMAK list:', err)
  } finally {
    loading.value = false
  }
}

function handleTabChange(tab: 'all' | 'draft' | 'pending' | 'active' | 'history') {
  activeTab.value = tab
  fetchSimaks(1)
}

function navigateToCreate() {
  router.push('/simak/new')
}

function navigateToDetail(id: string) {
  router.push(`/simak/${id}`)
}

function navigateToEdit(id: string) {
  router.push(`/simak/${id}/edit`)
}

async function handleDeleteDraft(id: string) {
  if (confirm('Apakah Anda yakin ingin menghapus draft permohonan SIMAK ini?')) {
    try {
      await simakService.deleteSimak(id)
      fetchSimaks(pagination.value.page)
    } catch (err: any) {
      alert(err.response?.data?.message || 'Gagal menghapus draft')
    }
  }
}

async function handleSubmitDraft(id: string) {
  if (confirm('Ajukan permohonan SIMAK ini sekarang untuk ditinjau oleh Approver?')) {
    try {
      await simakService.submitSimak(id)
      fetchSimaks(pagination.value.page)
    } catch (err: any) {
      alert(err.response?.data?.message || 'Gagal mengajukan SIMAK')
    }
  }
}

function formatDate(iso: string) {
  if (!iso) return '-'
  return new Date(iso).toLocaleDateString('id-ID', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
  })
}

function getStatusBadge(status: string) {
  switch (status) {
    case 'DRAFT':
      return { variant: 'neutral' as const, label: 'Draft Pemohon', dot: false }
    case 'SUBMITTED':
      return { variant: 'warning' as const, label: 'Menunggu Review', dot: true }
    case 'UNDER_REVIEW':
      return { variant: 'warning' as const, label: 'Sedang Ditinjau', dot: true }
    case 'REQUIRES_REVISION':
      return { variant: 'warning' as const, label: 'Perlu Revisi', dot: true }
    case 'REJECTED':
      return { variant: 'danger' as const, label: 'Ditolak', dot: false }
    case 'APPROVED':
      return { variant: 'success' as const, label: 'Disetujui', dot: false }
    case 'ISSUED':
    case 'ACTIVE':
      return { variant: 'success' as const, label: 'SIMAK Aktif', dot: true }
    case 'VERIFIED':
      return { variant: 'inka' as const, label: 'Terverifikasi Gate', dot: true }
    case 'COMPLETED':
      return { variant: 'neutral' as const, label: 'Selesai', dot: false }
    case 'CANCELLED':
      return { variant: 'neutral' as const, label: 'Dibatalkan', dot: false }
    case 'EXPIRED':
      return { variant: 'danger' as const, label: 'Kedaluwarsa', dot: false }
    default:
      return { variant: 'neutral' as const, label: status, dot: false }
  }
}

onMounted(() => {
  fetchSimaks()
})
</script>

<template>
  <div class="space-y-6">
    <!-- Header Banner -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 bg-white p-5 rounded-xl border border-slate-200 shadow-card">
      <div>
        <div class="flex items-center gap-2">
          <h1 class="text-xl font-bold text-slate-900">Permohonan SIMAK</h1>
          <Badge variant="inka" size="sm">Applicant Flow</Badge>
        </div>
        <p class="text-xs text-slate-500 mt-1">
          Surat Ijin Membawa Alat Kerja — Pencatatan digital peralatan & perangkat kerja yang dibawa pekerja ke area PT INKA (Persero).
        </p>
      </div>

      <div class="flex items-center gap-2.5">
        <Button variant="primary" size="sm" @click="navigateToCreate">
          <svg class="w-4 h-4 mr-1.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
          </svg>
          Buat Permohonan Baru
        </Button>
      </div>
    </div>

    <!-- Tabs & Filter Bar -->
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-3 bg-white p-2 rounded-xl border border-slate-200 shadow-sm">
      <!-- Tabs -->
      <div class="flex items-center gap-1 overflow-x-auto pb-1 md:pb-0 text-xs font-semibold">
        <button
          @click="handleTabChange('all')"
          :class="[
            'px-3 py-1.5 rounded-lg transition-all cursor-pointer whitespace-nowrap',
            activeTab === 'all' ? 'bg-slate-900 text-white shadow-sm' : 'text-slate-600 hover:text-slate-900 hover:bg-slate-100',
          ]"
        >
          Semua Pengajuan
        </button>
        <button
          @click="handleTabChange('draft')"
          :class="[
            'px-3 py-1.5 rounded-lg transition-all cursor-pointer whitespace-nowrap',
            activeTab === 'draft' ? 'bg-slate-900 text-white shadow-sm' : 'text-slate-600 hover:text-slate-900 hover:bg-slate-100',
          ]"
        >
          Draft Saya
        </button>
        <button
          @click="handleTabChange('pending')"
          :class="[
            'px-3 py-1.5 rounded-lg transition-all cursor-pointer whitespace-nowrap',
            activeTab === 'pending' ? 'bg-slate-900 text-white shadow-sm' : 'text-slate-600 hover:text-slate-900 hover:bg-slate-100',
          ]"
        >
          Menunggu Review
        </button>
        <button
          @click="handleTabChange('active')"
          :class="[
            'px-3 py-1.5 rounded-lg transition-all cursor-pointer whitespace-nowrap',
            activeTab === 'active' ? 'bg-slate-900 text-white shadow-sm' : 'text-slate-600 hover:text-slate-900 hover:bg-slate-100',
          ]"
        >
          Aktif / Berlaku
        </button>
        <button
          @click="handleTabChange('history')"
          :class="[
            'px-3 py-1.5 rounded-lg transition-all cursor-pointer whitespace-nowrap',
            activeTab === 'history' ? 'bg-slate-900 text-white shadow-sm' : 'text-slate-600 hover:text-slate-900 hover:bg-slate-100',
          ]"
        >
          Riwayat / Selesai
        </button>
      </div>

      <!-- Direction Filter Dropdown -->
      <div class="flex items-center gap-2 shrink-0">
        <label class="text-[11px] font-semibold text-slate-500">Arah Aliran:</label>
        <select
          v-model="directionFilter"
          @change="fetchSimaks(1)"
          class="px-2.5 py-1 text-xs bg-slate-50 border border-slate-200 rounded-lg text-slate-800 focus:outline-none focus:ring-1 focus:ring-inka-600"
        >
          <option value="ALL">Semua Arah</option>
          <option value="MASUK">Membawa Masuk</option>
          <option value="KELUAR">Membawa Keluar</option>
          <option value="MASUK_KELUAR">Masuk & Keluar</option>
        </select>
      </div>
    </div>

    <!-- Data Table -->
    <DataTable
      :columns="columns"
      :items="simaks"
      :loading="loading"
      :pagination="pagination"
      search-placeholder="Cari nomor SIMAK, keperluan, nama pekerja, workshop..."
      @search="(q) => { search = q; fetchSimaks(1) }"
      @page-change="(p) => fetchSimaks(p)"
    >
      <template #simak_number="{ row, value }">
        <div class="space-y-0.5">
          <button
            @click="navigateToDetail(row.id)"
            class="font-mono text-xs font-bold text-slate-900 hover:text-inka-700 transition-colors text-left cursor-pointer"
          >
            {{ value || 'Draft Tanpa Nomor' }}
          </button>
          <p class="text-[11px] text-slate-400">
            Dibuat {{ formatDate(row.created_at) }}
          </p>
        </div>
      </template>

      <template #purpose="{ row, value }">
        <div class="max-w-xs space-y-0.5">
          <p class="text-xs font-semibold text-slate-800 line-clamp-1" :title="value">
            {{ value }}
          </p>
          <div class="flex items-center gap-1.5 text-[11px] text-slate-500">
            <svg class="w-3.5 h-3.5 text-slate-400 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
            </svg>
            <span class="truncate">{{ row.destination_area }}</span>
          </div>
        </div>
      </template>

      <template #direction="{ value }">
        <span
          :class="[
            'px-2 py-0.5 rounded text-[10px] font-bold uppercase tracking-wider',
            value === 'MASUK'
              ? 'bg-blue-50 text-blue-700 border border-blue-200'
              : value === 'KELUAR'
                ? 'bg-amber-50 text-amber-800 border border-amber-200'
                : 'bg-purple-50 text-purple-700 border border-purple-200',
          ]"
        >
          {{ value === 'MASUK_KELUAR' ? 'Masuk/Keluar' : value }}
        </span>
      </template>

      <template #validity="{ row }">
        <div class="text-[11px] font-mono text-slate-700 space-y-0.5 text-center">
          <div>{{ formatDate(row.start_date) }}</div>
          <div class="text-[10px] text-slate-400">s.d {{ formatDate(row.end_date) }}</div>
        </div>
      </template>

      <template #items_count="{ row }">
        <span class="font-bold text-xs text-slate-800 px-2 py-0.5 rounded bg-slate-100 border border-slate-200">
          {{ row.items ? row.items.length : 0 }} Item
        </span>
      </template>

      <template #status="{ value }">
        <Badge
          :variant="getStatusBadge(value).variant"
          :dot="getStatusBadge(value).dot"
          size="sm"
        >
          {{ getStatusBadge(value).label }}
        </Badge>
      </template>

      <template #actions="{ row }">
        <div class="flex items-center justify-end gap-1.5">
          <!-- View Detail Button -->
          <button
            @click="navigateToDetail(row.id)"
            class="px-2 py-1 text-[11px] font-semibold rounded text-slate-600 hover:bg-slate-100 hover:text-slate-900 cursor-pointer transition-colors"
          >
            Lihat
          </button>

          <!-- Edit Button (Draft or Requires Revision only) -->
          <button
            v-if="row.status === 'DRAFT' || row.status === 'REQUIRES_REVISION'"
            @click="navigateToEdit(row.id)"
            class="px-2 py-1 text-[11px] font-semibold rounded text-amber-700 hover:bg-amber-50 cursor-pointer transition-colors"
          >
            Ubah
          </button>

          <!-- Submit Button (Draft only) -->
          <button
            v-if="row.status === 'DRAFT'"
            @click="handleSubmitDraft(row.id)"
            class="px-2 py-1 text-[11px] font-semibold rounded text-emerald-700 hover:bg-emerald-50 cursor-pointer transition-colors"
          >
            Ajukan
          </button>

          <!-- Delete Button (Draft only) -->
          <button
            v-if="row.status === 'DRAFT'"
            @click="handleDeleteDraft(row.id)"
            class="px-2 py-1 text-[11px] font-semibold rounded text-rose-600 hover:bg-rose-50 cursor-pointer transition-colors"
          >
            Hapus
          </button>
        </div>
      </template>
    </DataTable>
  </div>
</template>
