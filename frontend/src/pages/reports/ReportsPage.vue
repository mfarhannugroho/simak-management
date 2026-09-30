<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import Button from '@/components/ui/Button.vue'
import Badge from '@/components/ui/Badge.vue'
import {
  reportService,
  type SimakReportItem,
  type VerificationReportItem,
  type DiscrepancyReportItem,
  type ReportFilter,
  type ReportPaginationMeta
} from '@/services/report'

const activeTab = ref<'simak' | 'verification' | 'discrepancy'>('simak')
const loading = ref(false)
const exporting = ref(false)

// Data states
const simakData = ref<SimakReportItem[]>([])
const verifData = ref<VerificationReportItem[]>([])
const discData = ref<DiscrepancyReportItem[]>([])

const pagination = ref<ReportPaginationMeta>({
  page: 1,
  page_size: 15,
  total: 0,
  total_pages: 1,
})

// Filters
const search = ref('')
const dateFrom = ref('')
const dateTo = ref('')
const statusFilter = ref('')
const directionFilter = ref('')
const gateFilter = ref('')

async function fetchData(page = 1) {
  loading.value = true
  pagination.value.page = page

  const params: ReportFilter = {
    search: search.value || undefined,
    date_from: dateFrom.value || undefined,
    date_to: dateTo.value || undefined,
    page: pagination.value.page,
    page_size: pagination.value.page_size,
  }

  try {
    if (activeTab.value === 'simak') {
      if (statusFilter.value) params.status = statusFilter.value
      if (directionFilter.value) params.direction = directionFilter.value
      const res = await reportService.getSimakReport(params)
      simakData.value = res.data
      pagination.value = res.meta
    } else if (activeTab.value === 'verification') {
      if (statusFilter.value) params.status = statusFilter.value
      if (gateFilter.value) params.gate_name = gateFilter.value
      const res = await reportService.getVerificationReport(params)
      verifData.value = res.data
      pagination.value = res.meta
    } else {
      if (statusFilter.value) params.status = statusFilter.value
      const res = await reportService.getDiscrepancyReport(params)
      discData.value = res.data
      pagination.value = res.meta
    }
  } catch (err) {
    console.error('Failed to load report data', err)
  } finally {
    loading.value = false
  }
}

watch(activeTab, () => {
  // Reset page and specific filters when switching tab
  statusFilter.value = ''
  pagination.value.page = 1
  fetchData(1)
})

let debounceTimer: any = null
function onSearchChange() {
  clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    fetchData(1)
  }, 400)
}

function resetFilters() {
  search.value = ''
  dateFrom.value = ''
  dateTo.value = ''
  statusFilter.value = ''
  directionFilter.value = ''
  gateFilter.value = ''
  fetchData(1)
}

async function handleExport() {
  exporting.value = true
  try {
    const params: ReportFilter = {
      search: search.value || undefined,
      date_from: dateFrom.value || undefined,
      date_to: dateTo.value || undefined,
    }
    if (activeTab.value === 'simak') {
      if (statusFilter.value) params.status = statusFilter.value
      if (directionFilter.value) params.direction = directionFilter.value
    } else if (activeTab.value === 'verification') {
      if (statusFilter.value) params.status = statusFilter.value
      if (gateFilter.value) params.gate_name = gateFilter.value
    } else {
      if (statusFilter.value) params.status = statusFilter.value
    }

    await reportService.exportCSV(activeTab.value, params)
  } catch (err) {
    console.error('Export failed', err)
    alert('Gagal mengekspor laporan. Silakan coba kembali.')
  } finally {
    exporting.value = false
  }
}

function handlePrint() {
  window.print()
}

function getStatusVariant(status: string): 'success' | 'danger' | 'warning' | 'neutral' | 'inka' {
  switch (status) {
    case 'ACTIVE':
    case 'ISSUED':
    case 'APPROVED':
    case 'MATCH':
    case 'RESOLVED':
    case 'CLOSED':
      return 'success'
    case 'SUBMITTED':
    case 'UNDER_REVIEW':
    case 'IN_REVIEW':
      return 'warning'
    case 'REJECTED':
    case 'MISMATCH':
    case 'OPEN':
      return 'danger'
    default:
      return 'neutral'
  }
}

onMounted(() => {
  fetchData(1)
})
</script>

<template>
  <div class="space-y-6">
    <!-- Header with Print & Export actions -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 print:hidden">
      <div>
        <h1 class="text-2xl font-bold text-slate-900 tracking-tight flex items-center gap-2">
          <span>Laporan & Rekapitulasi Terpadu</span>
          <span class="text-xs px-2 py-0.5 rounded-full font-semibold bg-inka-100 text-inka-800 border border-inka-200">
            Audit Trail
          </span>
        </h1>
        <p class="text-xs text-slate-500 mt-1">
          Unduh, cetak, dan filter seluruh histori izin pergerakan, catatan pos gerbang, dan audit ketidaksesuaian peralatan.
        </p>
      </div>

      <div class="flex items-center gap-2.5">
        <Button variant="outline" size="sm" @click="handlePrint">
          <svg class="w-4 h-4 mr-1.5 text-slate-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 17h2a2 2 0 002-2v-4a2 2 0 00-2-2H5a2 2 0 00-2 2v4a2 2 0 002 2h2m2 4h6a2 2 0 002-2v-4a2 2 0 00-2-2H9a2 2 0 00-2 2v4a2 2 0 002 2zm8-12V5a2 2 0 00-2-2H9a2 2 0 00-2 2v4h10z" />
          </svg>
          Cetak Dokumen
        </Button>
        <Button variant="primary" size="sm" @click="handleExport" :loading="exporting">
          <svg class="w-4 h-4 mr-1.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
          </svg>
          Ekspor CSV (Excel)
        </Button>
      </div>
    </div>

    <!-- Print Only Header -->
    <div class="hidden print:block mb-6 border-b pb-4">
      <div class="flex items-center justify-between">
        <div>
          <h2 class="text-lg font-bold text-slate-900">PT INDUSTRI KERETA API (PERSERO)</h2>
          <p class="text-xs text-slate-500">Sistem Informasi Manajemen Barang Bawaan Pekerja (SIMAK)</p>
        </div>
        <div class="text-right text-xs text-slate-500">
          <p>Tanggal Cetak: {{ new Date().toLocaleDateString('id-ID', { dateStyle: 'full' }) }}</p>
          <p>Kategori: {{ activeTab.toUpperCase() }}</p>
        </div>
      </div>
    </div>

    <!-- Tabs Navigation -->
    <div class="border-b border-slate-200 print:hidden">
      <nav class="flex space-x-6">
        <button
          @click="activeTab = 'simak'"
          :class="activeTab === 'simak' ? 'border-inka-600 text-inka-600 font-bold' : 'border-transparent text-slate-500 hover:text-slate-700 font-medium'"
          class="py-3 px-1 border-b-2 text-sm flex items-center gap-2 cursor-pointer transition-colors"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
          </svg>
          <span>Rekapitulasi SIMAK</span>
        </button>

        <button
          @click="activeTab = 'verification'"
          :class="activeTab === 'verification' ? 'border-inka-600 text-inka-600 font-bold' : 'border-transparent text-slate-500 hover:text-slate-700 font-medium'"
          class="py-3 px-1 border-b-2 text-sm flex items-center gap-2 cursor-pointer transition-colors"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
          </svg>
          <span>Verifikasi Pos Gerbang</span>
        </button>

        <button
          @click="activeTab = 'discrepancy'"
          :class="activeTab === 'discrepancy' ? 'border-inka-600 text-inka-600 font-bold' : 'border-transparent text-slate-500 hover:text-slate-700 font-medium'"
          class="py-3 px-1 border-b-2 text-sm flex items-center gap-2 cursor-pointer transition-colors"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
          </svg>
          <span>Audit Ketidaksesuaian</span>
        </button>
      </nav>
    </div>

    <!-- Filter Control Bar -->
    <div class="bg-white p-4 rounded-xl border border-slate-200 shadow-sm space-y-3 print:hidden">
      <div class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-4 lg:grid-cols-5 gap-3">
        <!-- Search Field -->
        <div class="lg:col-span-2">
          <label class="block text-[11px] font-semibold text-slate-600 mb-1">Cari Data</label>
          <div class="relative">
            <input
              v-model="search"
              @input="onSearchChange"
              type="text"
              placeholder="No. SIMAK, personil, atau catatan..."
              class="w-full text-xs rounded-lg border border-slate-300 pl-8 pr-3 py-2 focus:ring-1 focus:ring-inka-600 focus:border-inka-600 outline-none"
            />
            <svg class="w-4 h-4 text-slate-400 absolute left-2.5 top-2.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
          </div>
        </div>

        <!-- Date From -->
        <div>
          <label class="block text-[11px] font-semibold text-slate-600 mb-1">Dari Tanggal</label>
          <input
            v-model="dateFrom"
            @change="fetchData(1)"
            type="date"
            class="w-full text-xs rounded-lg border border-slate-300 px-2.5 py-1.5 focus:ring-1 focus:ring-inka-600 focus:border-inka-600 outline-none"
          />
        </div>

        <!-- Date To -->
        <div>
          <label class="block text-[11px] font-semibold text-slate-600 mb-1">Sampai Tanggal</label>
          <input
            v-model="dateTo"
            @change="fetchData(1)"
            type="date"
            class="w-full text-xs rounded-lg border border-slate-300 px-2.5 py-1.5 focus:ring-1 focus:ring-inka-600 focus:border-inka-600 outline-none"
          />
        </div>

        <!-- Contextual Dropdown based on Tab -->
        <div>
          <label class="block text-[11px] font-semibold text-slate-600 mb-1">
            {{ activeTab === 'verification' ? 'Hasil Verifikasi' : 'Status' }}
          </label>
          <select
            v-model="statusFilter"
            @change="fetchData(1)"
            class="w-full text-xs rounded-lg border border-slate-300 px-2.5 py-2 focus:ring-1 focus:ring-inka-600 focus:border-inka-600 outline-none bg-white"
          >
            <option value="">Semua Status</option>
            <template v-if="activeTab === 'simak'">
              <option value="ACTIVE">ACTIVE</option>
              <option value="ISSUED">ISSUED</option>
              <option value="APPROVED">APPROVED</option>
              <option value="UNDER_REVIEW">UNDER_REVIEW</option>
              <option value="COMPLETED">COMPLETED</option>
              <option value="REJECTED">REJECTED</option>
            </template>
            <template v-else-if="activeTab === 'verification'">
              <option value="MATCH">100% MATCH</option>
              <option value="MISMATCH">MISMATCH (Ada Selisih)</option>
            </template>
            <template v-else>
              <option value="OPEN">TERBUKA (OPEN)</option>
              <option value="IN_REVIEW">DALAM TINJAUAN</option>
              <option value="RESOLVED">SELESAI (RESOLVED)</option>
              <option value="CLOSED">DITUTUP (CLOSED)</option>
            </template>
          </select>
        </div>
      </div>

      <div class="flex items-center justify-between pt-1 border-t border-slate-100 text-xs">
        <span class="text-slate-500">
          Ditemukan <strong class="text-slate-800">{{ pagination.total }}</strong> rekaman data
        </span>
        <button
          @click="resetFilters"
          class="text-inka-700 hover:text-inka-800 font-semibold cursor-pointer"
        >
          Reset Filter
        </button>
      </div>
    </div>

    <!-- Tab 1: SIMAK Report Table -->
    <div v-if="activeTab === 'simak'" class="bg-white rounded-xl border border-slate-200 shadow-sm overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-left text-xs">
          <thead class="bg-slate-50 border-b border-slate-200 text-slate-700 font-bold uppercase tracking-wider text-[11px]">
            <tr>
              <th class="py-3 px-4">No. SIMAK</th>
              <th class="py-3 px-4">Nama Pemohon & NIP</th>
              <th class="py-3 px-4">Departemen / Mitra</th>
              <th class="py-3 px-4">Arah & Area</th>
              <th class="py-3 px-4">Periode Berlaku</th>
              <th class="py-3 px-4 text-center">Jumlah Alat</th>
              <th class="py-3 px-4 text-center">Status</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100">
            <tr v-for="item in simakData" :key="item.id" class="hover:bg-slate-50/80 transition-colors">
              <td class="py-3 px-4 font-mono font-bold text-slate-900">
                {{ item.simak_number }}
              </td>
              <td class="py-3 px-4">
                <p class="font-semibold text-slate-900">{{ item.worker_name }}</p>
                <p class="text-[10px] text-slate-400 font-mono">{{ item.worker_nip }}</p>
              </td>
              <td class="py-3 px-4">
                <p class="text-slate-700 font-medium">{{ item.department_name }}</p>
                <p class="text-[10px] text-slate-400">{{ item.company_name }}</p>
              </td>
              <td class="py-3 px-4">
                <span class="inline-flex px-1.5 py-0.5 rounded text-[10px] font-semibold bg-slate-100 text-slate-700 mr-1.5">
                  {{ item.direction }}
                </span>
                <span class="text-slate-600">{{ item.destination_area }}</span>
              </td>
              <td class="py-3 px-4 text-slate-600">
                {{ item.start_date }} s.d {{ item.end_date }}
              </td>
              <td class="py-3 px-4 text-center font-bold text-slate-800">
                {{ item.total_items }} unit
              </td>
              <td class="py-3 px-4 text-center">
                <Badge :variant="getStatusVariant(item.status)" size="sm">
                  {{ item.status }}
                </Badge>
              </td>
            </tr>

            <tr v-if="!simakData.length && !loading">
              <td colspan="7" class="py-10 text-center text-slate-400 text-xs">
                Tidak ada data SIMAK yang cocok dengan filter yang dipilih.
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Tab 2: Verification Report Table -->
    <div v-else-if="activeTab === 'verification'" class="bg-white rounded-xl border border-slate-200 shadow-sm overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-left text-xs">
          <thead class="bg-slate-50 border-b border-slate-200 text-slate-700 font-bold uppercase tracking-wider text-[11px]">
            <tr>
              <th class="py-3 px-4">No. SIMAK</th>
              <th class="py-3 px-4">Nama Personil</th>
              <th class="py-3 px-4">Petugas & Pos Gerbang</th>
              <th class="py-3 px-4">Jenis Gerakan</th>
              <th class="py-3 px-4 text-center">Hasil Fisik</th>
              <th class="py-3 px-4 text-center">Qty (Fisik/Daftar)</th>
              <th class="py-3 px-4">Waktu Verifikasi</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100">
            <tr v-for="item in verifData" :key="item.id" class="hover:bg-slate-50/80 transition-colors">
              <td class="py-3 px-4 font-mono font-bold text-slate-900">
                {{ item.simak_number }}
              </td>
              <td class="py-3 px-4 font-semibold text-slate-800">
                {{ item.worker_name }}
              </td>
              <td class="py-3 px-4">
                <p class="font-medium text-slate-800">{{ item.gate_name }}</p>
                <p class="text-[10px] text-slate-400">Petugas: {{ item.inspector_name }}</p>
              </td>
              <td class="py-3 px-4">
                <span class="inline-flex px-2 py-0.5 rounded text-[10px] font-bold"
                  :class="item.inspection_type === 'MASUK' ? 'bg-blue-50 text-blue-700 border border-blue-200' : 'bg-amber-50 text-amber-700 border border-amber-200'">
                  {{ item.inspection_type }}
                </span>
              </td>
              <td class="py-3 px-4 text-center">
                <Badge :variant="item.status === 'MATCH' ? 'success' : 'danger'" size="sm">
                  {{ item.status }}
                </Badge>
              </td>
              <td class="py-3 px-4 text-center font-mono font-semibold"
                :class="item.total_actual_items === item.total_registered_items ? 'text-slate-800' : 'text-rose-600 font-bold'">
                {{ item.total_actual_items }} / {{ item.total_registered_items }}
              </td>
              <td class="py-3 px-4 text-slate-500 font-mono text-[11px]">
                {{ new Date(item.inspected_at).toLocaleString('id-ID') }}
              </td>
            </tr>

            <tr v-if="!verifData.length && !loading">
              <td colspan="7" class="py-10 text-center text-slate-400 text-xs">
                Tidak ada data verifikasi pos yang cocok dengan filter yang dipilih.
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Tab 3: Discrepancy Report Table -->
    <div v-else class="bg-white rounded-xl border border-slate-200 shadow-sm overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-left text-xs">
          <thead class="bg-slate-50 border-b border-slate-200 text-slate-700 font-bold uppercase tracking-wider text-[11px]">
            <tr>
              <th class="py-3 px-4">No. Kasus</th>
              <th class="py-3 px-4">No. SIMAK & Personil</th>
              <th class="py-3 px-4">Peralatan</th>
              <th class="py-3 px-4">Tipe Masalah</th>
              <th class="py-3 px-4 text-center">Qty (Riil/Daftar)</th>
              <th class="py-3 px-4 text-center">Status Solusi</th>
              <th class="py-3 px-4">Tgl Temuan</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100">
            <tr v-for="item in discData" :key="item.id" class="hover:bg-slate-50/80 transition-colors">
              <td class="py-3 px-4 font-mono font-bold text-rose-700">
                {{ item.discrepancy_number }}
              </td>
              <td class="py-3 px-4">
                <p class="font-semibold text-slate-900">{{ item.simak_number }}</p>
                <p class="text-[10px] text-slate-400">{{ item.worker_name }}</p>
              </td>
              <td class="py-3 px-4 font-medium text-slate-800">
                {{ item.item_name }}
              </td>
              <td class="py-3 px-4">
                <span class="inline-flex px-1.5 py-0.5 rounded text-[10px] font-semibold bg-rose-50 text-rose-700 border border-rose-200">
                  {{ item.discrepancy_type }}
                </span>
              </td>
              <td class="py-3 px-4 text-center font-mono font-bold text-slate-800">
                {{ item.actual_qty }} / {{ item.expected_qty }}
              </td>
              <td class="py-3 px-4 text-center">
                <Badge :variant="getStatusVariant(item.status)" size="sm">
                  {{ item.status }}
                </Badge>
              </td>
              <td class="py-3 px-4 text-slate-500 font-mono text-[11px]">
                {{ new Date(item.created_at).toLocaleDateString('id-ID') }}
              </td>
            </tr>

            <tr v-if="!discData.length && !loading">
              <td colspan="7" class="py-10 text-center text-slate-400 text-xs">
                Tidak ada data temuan ketidaksesuaian yang tercatat pada filter ini.
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Pagination Footer -->
    <div class="flex items-center justify-between pt-2 print:hidden text-xs text-slate-500">
      <div>
        Halaman {{ pagination.page }} dari {{ pagination.total_pages || 1 }} (Total {{ pagination.total }} data)
      </div>

      <div class="flex items-center gap-1.5">
        <Button
          variant="outline"
          size="sm"
          :disabled="pagination.page <= 1 || loading"
          @click="fetchData(pagination.page - 1)"
        >
          Sebelumnya
        </Button>
        <Button
          variant="outline"
          size="sm"
          :disabled="pagination.page >= pagination.total_pages || loading"
          @click="fetchData(pagination.page + 1)"
        >
          Selanjutnya
        </Button>
      </div>
    </div>
  </div>
</template>
