<script setup lang="ts">
import { ref, onMounted } from 'vue'
import Button from '@/components/ui/Button.vue'
import Badge from '@/components/ui/Badge.vue'
import Modal from '@/components/ui/Modal.vue'
import {
  auditService,
  type AuditLogItem,
  type AuditFilter,
  type AuditPaginationMeta
} from '@/services/audit'

const loading = ref(false)
const exporting = ref(false)
const auditLogs = ref<AuditLogItem[]>([])

const pagination = ref<AuditPaginationMeta>({
  page: 1,
  page_size: 15,
  total: 0,
  total_pages: 1,
})

// Filters
const search = ref('')
const actionFilter = ref('')
const entityFilter = ref('')
const dateFrom = ref('')
const dateTo = ref('')

// Detail Modal
const showDetailModal = ref(false)
const selectedLog = ref<AuditLogItem | null>(null)

async function fetchAuditLogs(page = 1) {
  loading.value = true
  pagination.value.page = page

  const params: AuditFilter = {
    search: search.value || undefined,
    action: actionFilter.value || undefined,
    entity_type: entityFilter.value || undefined,
    date_from: dateFrom.value || undefined,
    date_to: dateTo.value || undefined,
    page: pagination.value.page,
    page_size: pagination.value.page_size,
  }

  try {
    const res = await auditService.getAuditLogs(params)
    auditLogs.value = res.data
    pagination.value = res.meta
  } catch (err) {
    console.error('Failed to load audit logs', err)
  } finally {
    loading.value = false
  }
}

let debounceTimer: any = null
function onSearchChange() {
  clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    fetchAuditLogs(1)
  }, 400)
}

function resetFilters() {
  search.value = ''
  actionFilter.value = ''
  entityFilter.value = ''
  dateFrom.value = ''
  dateTo.value = ''
  fetchAuditLogs(1)
}

async function handleExport() {
  exporting.value = true
  try {
    const params: AuditFilter = {
      search: search.value || undefined,
      action: actionFilter.value || undefined,
      entity_type: entityFilter.value || undefined,
      date_from: dateFrom.value || undefined,
      date_to: dateTo.value || undefined,
    }
    await auditService.exportCSV(params)
  } catch (err) {
    console.error('Export failed', err)
    alert('Gagal mengekspor riwayat audit. Silakan coba kembali.')
  } finally {
    exporting.value = false
  }
}

function openDetail(item: AuditLogItem) {
  selectedLog.value = item
  showDetailModal.value = true
}

function formatJsonDetails(raw: string) {
  try {
    const parsed = JSON.parse(raw)
    return JSON.stringify(parsed, null, 2)
  } catch {
    return raw
  }
}

function getActionVariant(action: string): 'success' | 'danger' | 'warning' | 'neutral' | 'inka' {
  switch (action) {
    case 'LOGIN':
    case 'SIMAK_APPROVED':
    case 'SIMAK_ISSUED':
    case 'DISCREPANCY_RESOLVED':
      return 'success'
    case 'SIMAK_SUBMITTED':
    case 'DISCREPANCY_REVIEW_STARTED':
      return 'warning'
    case 'SIMAK_REJECTED':
    case 'DISCREPANCY_CREATED':
      return 'danger'
    case 'SIMAK_CREATED':
    case 'INSPECTION_RECORDED':
      return 'inka'
    default:
      return 'neutral'
  }
}

onMounted(() => {
  fetchAuditLogs(1)
})
</script>

<template>
  <div class="space-y-6">
    <!-- Page Header -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div>
        <h1 class="text-2xl font-bold text-slate-900 tracking-tight flex items-center gap-2">
          <span>Audit Trail & Compliance Log</span>
          <span class="text-xs px-2.5 py-0.5 rounded-full font-semibold bg-emerald-100 text-emerald-800 border border-emerald-200">
            Immutable Log
          </span>
        </h1>
        <p class="text-xs text-slate-500 mt-1">
          Pencatatan tamper-proof seluruh aksi sensitif, alur persetujuan, pemeriksaan pos gerbang, dan identitas perangkat.
        </p>
      </div>

      <div class="flex items-center gap-2.5">
        <Button variant="primary" size="sm" @click="handleExport" :loading="exporting">
          <svg class="w-4 h-4 mr-1.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
          </svg>
          Ekspor Log CSV
        </Button>
      </div>
    </div>

    <!-- Filter Control Bar -->
    <div class="bg-white p-4 rounded-xl border border-slate-200 shadow-sm space-y-3">
      <div class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-5 gap-3">
        <!-- Search Field -->
        <div class="md:col-span-2">
          <label class="block text-[11px] font-semibold text-slate-600 mb-1">Cari Log</label>
          <div class="relative">
            <input
              v-model="search"
              @input="onSearchChange"
              type="text"
              placeholder="Aksi, nama pengguna, IP, no. SIMAK..."
              class="w-full text-xs rounded-lg border border-slate-300 pl-8 pr-3 py-2 focus:ring-1 focus:ring-inka-600 focus:border-inka-600 outline-none"
            />
            <svg class="w-4 h-4 text-slate-400 absolute left-2.5 top-2.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
          </div>
        </div>

        <!-- Action Filter -->
        <div>
          <label class="block text-[11px] font-semibold text-slate-600 mb-1">Jenis Aksi</label>
          <select
            v-model="actionFilter"
            @change="fetchAuditLogs(1)"
            class="w-full text-xs rounded-lg border border-slate-300 px-2.5 py-2 focus:ring-1 focus:ring-inka-600 focus:border-inka-600 outline-none bg-white"
          >
            <option value="">Semua Aksi</option>
            <option value="LOGIN">LOGIN</option>
            <option value="SIMAK_CREATED">SIMAK_CREATED</option>
            <option value="SIMAK_SUBMITTED">SIMAK_SUBMITTED</option>
            <option value="SIMAK_APPROVED">SIMAK_APPROVED</option>
            <option value="SIMAK_REJECTED">SIMAK_REJECTED</option>
            <option value="SIMAK_ISSUED">SIMAK_ISSUED</option>
            <option value="INSPECTION_RECORDED">INSPECTION_RECORDED</option>
            <option value="DISCREPANCY_CREATED">DISCREPANCY_CREATED</option>
            <option value="DISCREPANCY_RESOLVED">DISCREPANCY_RESOLVED</option>
          </select>
        </div>

        <!-- Entity Type Filter -->
        <div>
          <label class="block text-[11px] font-semibold text-slate-600 mb-1">Tipe Entitas</label>
          <select
            v-model="entityFilter"
            @change="fetchAuditLogs(1)"
            class="w-full text-xs rounded-lg border border-slate-300 px-2.5 py-2 focus:ring-1 focus:ring-inka-600 focus:border-inka-600 outline-none bg-white"
          >
            <option value="">Semua Entitas</option>
            <option value="AUTH">AUTH</option>
            <option value="SIMAK">SIMAK</option>
            <option value="INSPECTION">INSPECTION</option>
            <option value="DISCREPANCY">DISCREPANCY</option>
          </select>
        </div>

        <!-- Date Range -->
        <div>
          <label class="block text-[11px] font-semibold text-slate-600 mb-1">Tanggal</label>
          <input
            v-model="dateFrom"
            @change="fetchAuditLogs(1)"
            type="date"
            class="w-full text-xs rounded-lg border border-slate-300 px-2.5 py-1.5 focus:ring-1 focus:ring-inka-600 focus:border-inka-600 outline-none"
          />
        </div>
      </div>

      <div class="flex items-center justify-between pt-1 border-t border-slate-100 text-xs">
        <span class="text-slate-500">
          Total <strong class="text-slate-800">{{ pagination.total }}</strong> jejak audit tercatat
        </span>
        <button
          @click="resetFilters"
          class="text-inka-700 hover:text-inka-800 font-semibold cursor-pointer"
        >
          Reset Filter
        </button>
      </div>
    </div>

    <!-- Audit Data Table -->
    <div class="bg-white rounded-xl border border-slate-200 shadow-sm overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-left text-xs">
          <thead class="bg-slate-50 border-b border-slate-200 text-slate-700 font-bold uppercase tracking-wider text-[11px]">
            <tr>
              <th class="py-3 px-4">Waktu (WIB)</th>
              <th class="py-3 px-4">Aktor Pengguna</th>
              <th class="py-3 px-4 text-center">Aksi Sensitif</th>
              <th class="py-3 px-4">Entitas & ID</th>
              <th class="py-3 px-4">Alamat IP & Klien</th>
              <th class="py-3 px-4 text-center">Aksi</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100">
            <tr v-for="item in auditLogs" :key="item.id" class="hover:bg-slate-50/80 transition-colors">
              <td class="py-3 px-4 text-slate-600 font-mono text-[11px] whitespace-nowrap">
                {{ new Date(item.created_at).toLocaleString('id-ID') }}
              </td>
              <td class="py-3 px-4">
                <p class="font-semibold text-slate-900">{{ item.user_name || '-' }}</p>
                <p class="text-[10px] text-slate-400 font-mono">{{ item.user_email || '-' }}</p>
              </td>
              <td class="py-3 px-4 text-center">
                <Badge :variant="getActionVariant(item.action)" size="sm">
                  {{ item.action }}
                </Badge>
              </td>
              <td class="py-3 px-4">
                <span class="inline-flex px-1.5 py-0.5 rounded text-[10px] font-semibold bg-slate-100 text-slate-700 mr-1.5">
                  {{ item.entity_type }}
                </span>
                <span class="font-mono text-slate-800 font-semibold text-[11px]">
                  {{ item.entity_id || '-' }}
                </span>
              </td>
              <td class="py-3 px-4">
                <p class="font-mono text-[11px] text-slate-800">{{ item.ip_address || '-' }}</p>
                <p class="text-[10px] text-slate-400 truncate max-w-xs" :title="item.user_agent">
                  {{ item.user_agent || '-' }}
                </p>
              </td>
              <td class="py-3 px-4 text-center">
                <Button variant="outline" size="sm" class="text-[11px] py-1 px-2.5" @click="openDetail(item)">
                  Rincian
                </Button>
              </td>
            </tr>

            <tr v-if="!auditLogs.length && !loading">
              <td colspan="6" class="py-10 text-center text-slate-400 text-xs">
                Tidak ada log audit yang sesuai dengan filter yang dipilih.
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Pagination Footer -->
    <div class="flex items-center justify-between pt-2 text-xs text-slate-500">
      <div>
        Halaman {{ pagination.page }} dari {{ pagination.total_pages || 1 }} (Total {{ pagination.total }} log)
      </div>

      <div class="flex items-center gap-1.5">
        <Button
          variant="outline"
          size="sm"
          :disabled="pagination.page <= 1 || loading"
          @click="fetchAuditLogs(pagination.page - 1)"
        >
          Sebelumnya
        </Button>
        <Button
          variant="outline"
          size="sm"
          :disabled="pagination.page >= pagination.total_pages || loading"
          @click="fetchAuditLogs(pagination.page + 1)"
        >
          Selanjutnya
        </Button>
      </div>
    </div>

    <!-- Detail Mutasi Modal -->
    <Modal
      v-model="showDetailModal"
      title="Rincian Audit Log"
      :subtitle="`ID: ${selectedLog?.id || '-'}`"
      maxWidth="xl"
    >
      <div v-if="selectedLog" class="space-y-4 text-xs">
        <div class="grid grid-cols-2 gap-3 p-3 bg-slate-50 rounded-lg border border-slate-200">
          <div>
            <p class="text-[11px] text-slate-400">Aktor Pengguna</p>
            <p class="font-semibold text-slate-900">{{ selectedLog.user_name }} ({{ selectedLog.user_email }})</p>
          </div>
          <div>
            <p class="text-[11px] text-slate-400">Waktu Pencatatan</p>
            <p class="font-semibold text-slate-900">{{ new Date(selectedLog.created_at).toLocaleString('id-ID') }}</p>
          </div>
          <div>
            <p class="text-[11px] text-slate-400">Aksi & Entitas</p>
            <p class="font-semibold text-slate-900">{{ selectedLog.action }} ({{ selectedLog.entity_type }})</p>
          </div>
          <div>
            <p class="text-[11px] text-slate-400">ID Entitas Terkait</p>
            <p class="font-semibold font-mono text-slate-900">{{ selectedLog.entity_id }}</p>
          </div>
          <div class="col-span-2">
            <p class="text-[11px] text-slate-400">Alamat IP & User Agent</p>
            <p class="font-mono text-slate-800">{{ selectedLog.ip_address }} — {{ selectedLog.user_agent }}</p>
          </div>
        </div>

        <div>
          <label class="block text-[11px] font-semibold text-slate-700 mb-1">Payload Mutasi Data (JSON)</label>
          <pre class="bg-slate-900 text-emerald-400 p-3 rounded-lg overflow-x-auto font-mono text-[11px] leading-relaxed max-h-60 border border-slate-800">{{ formatJsonDetails(selectedLog.details) }}</pre>
        </div>
      </div>

      <template #footer>
        <Button variant="outline" size="sm" @click="showDetailModal = false">
          Tutup
        </Button>
      </template>
    </Modal>
  </div>
</template>
