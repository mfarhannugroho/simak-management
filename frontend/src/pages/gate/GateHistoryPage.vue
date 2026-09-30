<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { inspectionService, type InspectionListItem, type Inspection } from '@/services/inspection'
import DataTable, { type Column, type Pagination } from '@/components/ui/DataTable.vue'
import Button from '@/components/ui/Button.vue'
import Badge from '@/components/ui/Badge.vue'
import Modal from '@/components/ui/Modal.vue'

const items = ref<InspectionListItem[]>([])
const loading = ref(false)
const search = ref('')

const filterGate = ref('')
const filterType = ref('ALL')
const filterStatus = ref('ALL')
const filterDate = ref('')

const availableGates = [
  'Semua Pos Gerbang',
  'Pos Gerbang Utama PT INKA (Madiun)',
  'Pos 2 Barat (Pintu Kendaraan Berat)',
  'Pos 3 Timur (Akses Pekerja & Logistik)',
  'Pos Jalur Rel KA & Uji Coba Lintasan',
  'Pos Gerbang Workshop Fabrikasi & Finishing',
]

const pagination = ref<Pagination>({
  page: 1,
  limit: 10,
  total_items: 0,
  total_pages: 1,
})

// Inspection Detail Modal state
const detailModalOpen = ref(false)
const selectedInspection = ref<Inspection | null>(null)
const detailLoading = ref(false)

const columns: Column[] = [
  { key: 'inspected_at', label: 'Waktu Pemeriksaan', sortable: true },
  { key: 'simak_number', label: 'Nomor SIMAK' },
  { key: 'worker', label: 'Pekerja & Perusahaan' },
  { key: 'gate_name', label: 'Pos Gerbang' },
  { key: 'inspection_type', label: 'Arah', align: 'center' },
  { key: 'status', label: 'Status Fisik', align: 'center' },
  { key: 'total_items', label: 'Jumlah Barang', align: 'center' },
  { key: 'inspector_name', label: 'Petugas' },
  { key: 'actions', label: 'Aksi', align: 'right' },
]

async function fetchHistory(page = 1) {
  loading.value = true
  try {
    const res = await inspectionService.listInspections({
      page,
      limit: pagination.value.limit,
      search: search.value,
      gate_name: filterGate.value === 'Semua Pos Gerbang' ? '' : filterGate.value,
      inspection_type: filterType.value === 'ALL' ? '' : filterType.value,
      status: filterStatus.value === 'ALL' ? '' : filterStatus.value,
      date: filterDate.value,
    })

    if (res.data) {
      items.value = res.data.items || []
      if (res.data.pagination) {
        pagination.value = res.data.pagination
      }
    }
  } catch (err) {
    console.error('Failed to load inspection history:', err)
  } finally {
    loading.value = false
  }
}

function handleSearch(query: string) {
  search.value = query
  pagination.value.page = 1
  fetchHistory(1)
}

function handlePageChange(page: number) {
  pagination.value.page = page
  fetchHistory(page)
}

function handleFilterChange() {
  pagination.value.page = 1
  fetchHistory(1)
}

function resetFilters() {
  search.value = ''
  filterGate.value = ''
  filterType.value = 'ALL'
  filterStatus.value = 'ALL'
  filterDate.value = ''
  pagination.value.page = 1
  fetchHistory(1)
}

async function viewDetail(id: string) {
  detailLoading.value = true
  detailModalOpen.value = true
  try {
    const res = await inspectionService.getInspection(id)
    if (res.data) {
      selectedInspection.value = res.data
    }
  } catch (err) {
    console.error('Failed to load inspection details:', err)
  } finally {
    detailLoading.value = false
  }
}

function formatDate(dateStr?: string) {
  if (!dateStr) return '-'
  const d = new Date(dateStr)
  return d.toLocaleDateString('id-ID', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

onMounted(() => {
  fetchHistory()
})
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div>
        <div class="flex items-center gap-2">
          <span class="px-2 py-0.5 text-xs font-semibold uppercase tracking-wider rounded bg-slate-100 text-slate-700">
            Log & Arsip
          </span>
          <span class="text-xs text-slate-500 font-medium">Rekonsiliasi Masuk & Keluar</span>
        </div>
        <h1 class="text-2xl font-bold text-slate-900 tracking-tight mt-1">
          Riwayat Verifikasi Pos Gerbang
        </h1>
        <p class="text-sm text-slate-600 mt-0.5">
          Log lengkap pencatatan fisik dan pergerakan alat kerja di seluruh pos jaga area pabrik PT INKA (Persero).
        </p>
      </div>

      <div class="flex items-center gap-2">
        <router-link to="/gate">
          <Button variant="primary">
            <svg class="w-4 h-4 mr-1.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v1m6 11h2m-6 0h-2v4m0-11v3m0 0h.01M12 12h4.01M16 20h4M4 12h4m12 0h.01M5 8h2a1 1 0 001-1V5a1 1 0 00-1-1H5a1 1 0 00-1 1v2a1 1 0 001 1zm12 0h2a1 1 0 001-1V5a1 1 0 00-1-1h-2a1 1 0 00-1 1v2a1 1 0 001 1zM5 20h2a1 1 0 001-1v-2a1 1 0 00-1-1H5a1 1 0 00-1 1v2a1 1 0 001 1z" />
            </svg>
            Buka Pemeriksaan Gate
          </Button>
        </router-link>
      </div>
    </div>

    <!-- Filters Bar -->
    <div class="bg-white border border-slate-200 rounded-xl p-4 shadow-sm space-y-3">
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3 text-xs">
        <!-- Pos Gerbang -->
        <div>
          <label class="block font-semibold text-slate-700 mb-1">Pos Gerbang</label>
          <select
            v-model="filterGate"
            @change="handleFilterChange"
            class="w-full px-3 py-2 border border-slate-300 rounded-lg bg-white text-slate-800 focus:outline-none focus:ring-1 focus:ring-inka-500"
          >
            <option value="">Semua Pos Gerbang</option>
            <option v-for="g in availableGates.slice(1)" :key="g" :value="g">
              {{ g }}
            </option>
          </select>
        </div>

        <!-- Arah Pergerakan -->
        <div>
          <label class="block font-semibold text-slate-700 mb-1">Arah Pergerakan</label>
          <select
            v-model="filterType"
            @change="handleFilterChange"
            class="w-full px-3 py-2 border border-slate-300 rounded-lg bg-white text-slate-800 focus:outline-none focus:ring-1 focus:ring-inka-500"
          >
            <option value="ALL">Semua Arah</option>
            <option value="MASUK">MASUK PABRIK</option>
            <option value="KELUAR">KELUAR PABRIK</option>
          </select>
        </div>

        <!-- Status Hasil Fisik -->
        <div>
          <label class="block font-semibold text-slate-700 mb-1">Status Fisik</label>
          <select
            v-model="filterStatus"
            @change="handleFilterChange"
            class="w-full px-3 py-2 border border-slate-300 rounded-lg bg-white text-slate-800 focus:outline-none focus:ring-1 focus:ring-inka-500"
          >
            <option value="ALL">Semua Status</option>
            <option value="MATCH">COCOK (MATCH)</option>
            <option value="MISMATCH">SELISIH (MISMATCH)</option>
          </select>
        </div>

        <!-- Tanggal -->
        <div>
          <label class="block font-semibold text-slate-700 mb-1">Tanggal</label>
          <div class="flex gap-2">
            <input
              v-model="filterDate"
              type="date"
              @change="handleFilterChange"
              class="w-full px-3 py-2 border border-slate-300 rounded-lg bg-white text-slate-800 focus:outline-none focus:ring-1 focus:ring-inka-500"
            />
            <button
              v-if="filterDate || filterGate || filterType !== 'ALL' || filterStatus !== 'ALL'"
              type="button"
              @click="resetFilters"
              class="px-2.5 py-1 text-slate-500 hover:text-rose-600 text-xs font-semibold shrink-0"
              title="Reset Filter"
            >
              Reset
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Data Table -->
    <DataTable
      :columns="columns"
      :items="items"
      :loading="loading"
      :pagination="pagination"
      search-placeholder="Cari No. SIMAK, Nama Pekerja, Perusahaan..."
      @search="handleSearch"
      @page-change="handlePageChange"
    >
      <!-- Inspected At Cell -->
      <template #cell(inspected_at)="{ item }">
        <div class="font-medium text-slate-900 text-xs">
          {{ formatDate(item.inspected_at) }}
        </div>
      </template>

      <!-- Simak Number Cell -->
      <template #cell(simak_number)="{ item }">
        <span class="font-bold text-slate-900 text-xs font-mono">
          {{ item.simak_number }}
        </span>
      </template>

      <!-- Worker & Company Cell -->
      <template #cell(worker)="{ item }">
        <div>
          <div class="font-semibold text-slate-900 text-xs">{{ item.worker_name }}</div>
          <div class="text-[11px] text-slate-500">{{ item.company }}</div>
        </div>
      </template>

      <!-- Gate Name Cell -->
      <template #cell(gate_name)="{ item }">
        <span class="text-xs text-slate-700">
          {{ item.gate_name }}
        </span>
      </template>

      <!-- Inspection Type Cell -->
      <template #cell(inspection_type)="{ item }">
        <span
          class="px-2 py-0.5 rounded text-[11px] font-bold tracking-wide"
          :class="item.inspection_type === 'MASUK' ? 'bg-emerald-50 text-emerald-700 border border-emerald-200' : 'bg-blue-50 text-blue-700 border border-blue-200'"
        >
          {{ item.inspection_type }}
        </span>
      </template>

      <!-- Status Cell -->
      <template #cell(status)="{ item }">
        <Badge :variant="item.status === 'MATCH' ? 'success' : 'danger'" :dot="true" size="sm">
          {{ item.status === 'MATCH' ? 'COCOK' : 'SELISIH' }}
        </Badge>
      </template>

      <!-- Total Items Cell -->
      <template #cell(total_items)="{ item }">
        <span class="text-xs font-semibold text-slate-800">
          {{ item.total_actual_items }} / {{ item.total_registered_items }} Unit
        </span>
      </template>

      <!-- Inspector Name Cell -->
      <template #cell(inspector_name)="{ item }">
        <span class="text-xs text-slate-600">
          {{ item.inspector_name || 'Petugas' }}
        </span>
      </template>

      <!-- Actions Cell -->
      <template #cell(actions)="{ item }">
        <Button variant="secondary" size="sm" @click="viewDetail(item.id)">
          Rincian
        </Button>
      </template>
    </DataTable>

    <!-- Inspection Detail Modal -->
    <Modal
      v-model="detailModalOpen"
      title="Rincian Rekonsiliasi Pos Gerbang"
      @close="detailModalOpen = false"
    >
      <div v-if="detailLoading" class="py-12 text-center text-slate-500">
        <div class="inline-block w-6 h-6 border-2 border-inka-600 border-t-transparent rounded-full animate-spin"></div>
        <p class="text-xs mt-2">Memuat data pemeriksaan...</p>
      </div>

      <div v-else-if="selectedInspection" class="space-y-4">
        <!-- Header summary -->
        <div class="bg-slate-50 p-4 rounded-xl border border-slate-200 flex items-center justify-between">
          <div>
            <div class="text-xs text-slate-500">Nomor SIMAK</div>
            <div class="font-bold text-slate-900 text-sm font-mono">{{ selectedInspection.simak?.simak_number }}</div>
          </div>
          <div class="text-right">
            <Badge :variant="selectedInspection.status === 'MATCH' ? 'success' : 'danger'" :dot="true">
              {{ selectedInspection.status === 'MATCH' ? 'COCOK (MATCH)' : 'SELISIH (MISMATCH)' }}
            </Badge>
            <div class="text-xs text-slate-500 mt-1 font-semibold">{{ selectedInspection.inspection_type }}</div>
          </div>
        </div>

        <div class="grid grid-cols-2 gap-3 text-xs">
          <div>
            <span class="text-slate-500">Pos Gerbang:</span>
            <span class="font-semibold text-slate-800 ml-1">{{ selectedInspection.gate_name }}</span>
          </div>
          <div>
            <span class="text-slate-500">Waktu Pemeriksaan:</span>
            <span class="font-semibold text-slate-800 ml-1">{{ formatDate(selectedInspection.inspected_at) }}</span>
          </div>
          <div>
            <span class="text-slate-500">Pekerja:</span>
            <span class="font-semibold text-slate-800 ml-1">{{ selectedInspection.simak?.worker?.full_name }}</span>
          </div>
          <div>
            <span class="text-slate-500">Perusahaan:</span>
            <span class="font-semibold text-slate-800 ml-1">{{ selectedInspection.simak?.worker?.company?.name }}</span>
          </div>
        </div>

        <!-- Line items table -->
        <div>
          <div class="text-xs font-bold text-slate-800 mb-2">Checklist Peralatan Fisik:</div>
          <div class="border border-slate-200 rounded-lg overflow-hidden max-h-64 overflow-y-auto">
            <table class="w-full text-left text-xs">
              <thead class="bg-slate-100 text-slate-700 font-semibold border-b border-slate-200">
                <tr>
                  <th class="p-2.5">Nama Peralatan</th>
                  <th class="p-2.5 text-center">Terdaftar</th>
                  <th class="p-2.5 text-center">Fisik</th>
                  <th class="p-2.5 text-center">Status</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-100">
                <tr
                  v-for="it in selectedInspection.items"
                  :key="it.id"
                  :class="it.status === 'MATCH' ? '' : 'bg-rose-50/60'"
                >
                  <td class="p-2.5 font-medium text-slate-900">
                    <div>{{ it.item_name }}</div>
                    <div class="text-[11px] text-slate-500 font-normal">
                      {{ it.brand }} {{ it.serial_number ? `• SN: ${it.serial_number}` : '' }}
                    </div>
                    <div v-if="it.notes" class="text-[10px] text-rose-600 italic mt-0.5">
                      Catatan: {{ it.notes }}
                    </div>
                  </td>
                  <td class="p-2.5 text-center font-semibold text-slate-700">
                    {{ it.expected_qty }} {{ it.unit }}
                  </td>
                  <td class="p-2.5 text-center font-bold" :class="it.status === 'MATCH' ? 'text-slate-900' : 'text-rose-600'">
                    {{ it.actual_qty }} {{ it.unit }}
                  </td>
                  <td class="p-2.5 text-center">
                    <Badge :variant="it.status === 'MATCH' ? 'success' : 'danger'" size="sm">
                      {{ it.status }}
                    </Badge>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <!-- Notes if any -->
        <div v-if="selectedInspection.notes" class="bg-amber-50 border border-amber-200 rounded-lg p-3 text-xs text-amber-900">
          <span class="font-bold">Catatan Petugas: </span>
          <span>{{ selectedInspection.notes }}</span>
        </div>
      </div>
    </Modal>
  </div>
</template>
