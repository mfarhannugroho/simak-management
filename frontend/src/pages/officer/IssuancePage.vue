<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { documentService, type IssuanceQueueItem, type SimakDocument } from '@/services/document'
import DataTable, { type Column, type Pagination } from '@/components/ui/DataTable.vue'
import Button from '@/components/ui/Button.vue'
import Badge from '@/components/ui/Badge.vue'

const router = useRouter()

const items = ref<IssuanceQueueItem[]>([])
const loading = ref(false)
const search = ref('')
const activeTab = ref<'APPROVED' | 'ISSUED' | 'ALL'>('APPROVED')

const pagination = ref<Pagination>({
  page: 1,
  limit: 10,
  total_items: 0,
  total_pages: 1,
})

// Issuance modal state
const issueModalOpen = ref(false)
const issuing = ref(false)
const selectedItem = ref<IssuanceQueueItem | null>(null)
const customNumber = ref('')
const issueNotes = ref('')

// QR code viewer modal state
const qrModalOpen = ref(false)
const activeQrDoc = ref<SimakDocument | null>(null)
const qrLoading = ref(false)

const columns: Column[] = [
  { key: 'simak_number', label: 'No. SIMAK / Pengajuan', sortable: true },
  { key: 'worker', label: 'Pemohon & Unit / Vendor' },
  { key: 'purpose', label: 'Keperluan & Area Pabrik' },
  { key: 'validity', label: 'Periode Berlaku', align: 'center' },
  { key: 'total_items', label: 'Barang', align: 'center' },
  { key: 'status', label: 'Status Dokumen', align: 'center' },
  { key: 'actions', label: 'Aksi', align: 'right' },
]

async function fetchQueue(page = 1) {
  loading.value = true
  try {
    const res = await documentService.getIssuanceQueue({
      page,
      limit: pagination.value.limit,
      search: search.value,
      status: activeTab.value,
    })

    if (res.data) {
      items.value = res.data.items || []
      pagination.value = res.data.pagination
    }
  } catch (err) {
    console.error('Failed to load issuance queue:', err)
  } finally {
    loading.value = false
  }
}

function handleTabChange(tab: 'APPROVED' | 'ISSUED' | 'ALL') {
  activeTab.value = tab
  pagination.value.page = 1
  fetchQueue(1)
}

function handleSearch() {
  pagination.value.page = 1
  fetchQueue(1)
}

function openIssueModal(item: IssuanceQueueItem) {
  selectedItem.value = item
  customNumber.value = ''
  issueNotes.value = 'Dokumen resmi diterbitkan oleh Tim SIMAK Officer PT INKA (Persero).'
  issueModalOpen.value = true
}

async function handleConfirmIssue() {
  if (!selectedItem.value) return
  issuing.value = true
  try {
    await documentService.issueSimak(selectedItem.value.id, {
      custom_number: customNumber.value.trim() || undefined,
      notes: issueNotes.value.trim() || undefined,
    })
    issueModalOpen.value = false
    await fetchQueue(pagination.value.page)
  } catch (err: any) {
    alert(err?.response?.data?.message || 'Gagal menerbitkan dokumen SIMAK.')
  } finally {
    issuing.value = false
  }
}

async function openQrModal(simakId: string) {
  qrLoading.value = true
  qrModalOpen.value = true
  activeQrDoc.value = null
  try {
    const res = await documentService.getDocument(simakId)
    if (res.data) {
      activeQrDoc.value = res.data
    }
  } catch (err: any) {
    alert(err?.response?.data?.message || 'Gagal memuat dokumen QR.')
    qrModalOpen.value = false
  } finally {
    qrLoading.value = false
  }
}

function formatDate(dateStr?: string) {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  })
}

function copyVerifyLink(qrToken?: string) {
  if (!qrToken) return
  const url = `${window.location.origin}/verify/${qrToken}`
  navigator.clipboard.writeText(url)
  alert('Tautan verifikasi berhasil disalin ke clipboard:\n' + url)
}

onMounted(() => {
  fetchQueue(1)
})
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
      <div>
        <div class="flex items-center gap-2">
          <span class="inline-block w-2.5 h-2.5 rounded-full bg-red-600 animate-pulse"></span>
          <p class="text-xs font-semibold uppercase tracking-wider text-red-600">Otoritas SIMAK Officer</p>
        </div>
        <h1 class="text-2xl font-bold text-slate-900 mt-1">Penerbitan Dokumen Resmi & QR Code</h1>
        <p class="text-sm text-slate-500 mt-0.5">
          Terbitkan dokumen resmi SIMAK berformat standar INKA lengkap dengan penomoran unik dan kode QR pengaman
        </p>
      </div>
    </div>

    <!-- Quick Stats Cards -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
      <div class="bg-white rounded-xl border border-slate-200 p-4 shadow-sm flex items-center justify-between">
        <div>
          <p class="text-xs font-medium text-slate-500">Siap Diterbitkan</p>
          <p class="text-2xl font-bold text-amber-600 mt-1">
            {{ activeTab === 'APPROVED' ? pagination.total_items : '—' }}
          </p>
          <p class="text-xs text-slate-400 mt-0.5">Telah disetujui seluruh tingkatan</p>
        </div>
        <div class="w-12 h-12 rounded-xl bg-amber-50 text-amber-600 flex items-center justify-center">
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
        </div>
      </div>

      <div class="bg-white rounded-xl border border-slate-200 p-4 shadow-sm flex items-center justify-between">
        <div>
          <p class="text-xs font-medium text-slate-500">Standar Penomoran Resmi</p>
          <p class="text-sm font-mono font-bold text-slate-900 mt-1">SIMAK/INKA/YYYY/XXXXXX</p>
          <p class="text-xs text-slate-400 mt-0.5">Penomoran otomatis berurutan</p>
        </div>
        <div class="w-12 h-12 rounded-xl bg-slate-100 text-slate-700 flex items-center justify-center">
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 20l4-16m2 16l4-16M6 9h14M4 15h14" />
          </svg>
        </div>
      </div>

      <div class="bg-white rounded-xl border border-slate-200 p-4 shadow-sm flex items-center justify-between">
        <div>
          <p class="text-xs font-medium text-slate-500">Pengamanan Digital</p>
          <p class="text-sm font-bold text-slate-900 mt-1">Secure Cryptographic QR</p>
          <p class="text-xs text-slate-400 mt-0.5">Pemeriksaan instan di Pos Gerbang</p>
        </div>
        <div class="w-12 h-12 rounded-xl bg-emerald-50 text-emerald-600 flex items-center justify-center">
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v1m6 11h2m-6 0h-2v4m0-11v3m0 0h.01M12 12h4.01M16 20h4M4 12h4m12 0h.01M5 8h2a1 1 0 001-1V5a1 1 0 00-1-1H5a1 1 0 00-1 1v2a1 1 0 001 1zm12 0h2a1 1 0 001-1V5a1 1 0 00-1-1h-2a1 1 0 00-1 1v2a1 1 0 001 1zM5 20h2a1 1 0 001-1v-2a1 1 0 00-1-1H5a1 1 0 00-1 1v2a1 1 0 001 1z" />
          </svg>
        </div>
      </div>
    </div>

    <!-- Main Table Card -->
    <div class="bg-white rounded-xl border border-slate-200 shadow-sm overflow-hidden">
      <!-- Toolbar -->
      <div class="p-4 border-b border-slate-200 flex flex-col md:flex-row md:items-center md:justify-between gap-4">
        <!-- Tabs -->
        <div class="flex items-center gap-1 bg-slate-100 p-1 rounded-lg">
          <button
            type="button"
            class="px-4 py-1.5 text-xs font-medium rounded-md transition-colors"
            :class="activeTab === 'APPROVED' ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-600 hover:text-slate-900'"
            @click="handleTabChange('APPROVED')"
          >
            Siap Diterbitkan
          </button>
          <button
            type="button"
            class="px-4 py-1.5 text-xs font-medium rounded-md transition-colors"
            :class="activeTab === 'ISSUED' ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-600 hover:text-slate-900'"
            @click="handleTabChange('ISSUED')"
          >
            Telah Diterbitkan (Issued)
          </button>
          <button
            type="button"
            class="px-4 py-1.5 text-xs font-medium rounded-md transition-colors"
            :class="activeTab === 'ALL' ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-600 hover:text-slate-900'"
            @click="handleTabChange('ALL')"
          >
            Semua
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
        @page-change="fetchQueue"
      >
        <!-- Cell: SIMAK Number -->
        <template #cell-simak_number="{ row }">
          <div class="space-y-0.5">
            <p class="font-mono text-xs font-bold text-slate-900">{{ row.simak_number }}</p>
            <p class="text-[11px] text-slate-400">Dibuat: {{ formatDate(row.created_at) }}</p>
          </div>
        </template>

        <!-- Cell: Worker -->
        <template #cell-worker="{ row }">
          <div class="space-y-0.5">
            <p class="text-xs font-semibold text-slate-900">{{ row.worker_name }}</p>
            <p class="text-[11px] text-slate-500">
              {{ row.company }}
              <span v-if="row.department"> • {{ row.department }}</span>
            </p>
          </div>
        </template>

        <!-- Cell: Purpose -->
        <template #cell-purpose="{ row }">
          <div class="max-w-xs space-y-0.5">
            <p class="text-xs text-slate-800 line-clamp-1 font-medium">{{ row.purpose }}</p>
            <p class="text-[11px] text-slate-500 flex items-center gap-1">
              <svg class="w-3 h-3 text-slate-400 inline" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
              </svg>
              {{ row.destination_area }} ({{ row.direction }})
            </p>
          </div>
        </template>

        <!-- Cell: Validity -->
        <template #cell-validity="{ row }">
          <div class="text-center text-[11px] text-slate-600">
            <p class="font-medium">{{ formatDate(row.start_date) }}</p>
            <p class="text-slate-400 text-[10px]">s/d {{ formatDate(row.end_date) }}</p>
          </div>
        </template>

        <!-- Cell: Total Items -->
        <template #cell-total_items="{ row }">
          <div class="text-center">
            <span class="inline-flex items-center px-2 py-0.5 rounded bg-slate-100 text-slate-700 font-semibold text-xs">
              {{ row.total_items }} Item
            </span>
          </div>
        </template>

        <!-- Cell: Status -->
        <template #cell-status="{ row }">
          <div class="text-center">
            <Badge v-if="row.status === 'APPROVED'" variant="warning" size="sm">
              Siap Terbit
            </Badge>
            <Badge v-else-if="row.status === 'ISSUED' || row.status === 'ACTIVE'" variant="success" size="sm">
              Diterbitkan
            </Badge>
            <Badge v-else variant="neutral" size="sm">
              {{ row.status }}
            </Badge>
          </div>
        </template>

        <!-- Cell: Actions -->
        <template #cell-actions="{ row }">
          <div class="flex items-center justify-end gap-1.5">
            <!-- If Approved: Issue button -->
            <Button
              v-if="row.status === 'APPROVED'"
              variant="primary"
              size="sm"
              class="gap-1 shadow-sm"
              @click="openIssueModal(row)"
            >
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4M7.835 4.697a3.42 3.42 0 001.946-.806 3.42 3.42 0 014.438 0 3.42 3.42 0 001.946.806 3.42 3.42 0 013.138 3.138 3.42 3.42 0 00.806 1.946 3.42 3.42 0 010 4.438 3.42 3.42 0 00-.806 1.946 3.42 3.42 0 01-3.138 3.138 3.42 3.42 0 00-1.946.806 3.42 3.42 0 01-4.438 0 3.42 3.42 0 00-1.946-.806 3.42 3.42 0 01-3.138-3.138 3.42 3.42 0 00-.806-1.946 3.42 3.42 0 010-4.438 3.42 3.42 0 00.806-1.946 3.42 3.42 0 013.138-3.138z" />
              </svg>
              <span>Terbitkan SIMAK</span>
            </Button>

            <!-- If Already Issued -->
            <template v-else>
              <Button
                variant="outline"
                size="sm"
                class="gap-1 text-slate-700"
                @click="router.push(`/simak/${row.id}/preview`)"
              >
                <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 17h2a2 2 0 002-2v-4a2 2 0 00-2-2H5a2 2 0 00-2 2v4a2 2 0 002 2h2m2 4h6a2 2 0 002-2v-4a2 2 0 00-2-2H9a2 2 0 00-2 2v4a2 2 0 002 2zm8-12V5a2 2 0 00-2-2H9a2 2 0 00-2 2v4h10z" />
                </svg>
                <span>Cetak / PDF</span>
              </Button>

              <Button
                variant="outline"
                size="sm"
                class="gap-1 text-slate-700"
                title="Tampilkan QR Code"
                @click="openQrModal(row.id)"
              >
                <svg class="w-3.5 h-3.5 text-slate-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v1m6 11h2m-6 0h-2v4m0-11v3m0 0h.01M12 12h4.01M16 20h4M4 12h4m12 0h.01M5 8h2a1 1 0 001-1V5a1 1 0 00-1-1H5a1 1 0 00-1 1v2a1 1 0 001 1zm12 0h2a1 1 0 001-1V5a1 1 0 00-1-1h-2a1 1 0 00-1 1v2a1 1 0 001 1zM5 20h2a1 1 0 001-1v-2a1 1 0 00-1-1H5a1 1 0 00-1 1v2a1 1 0 001 1z" />
                </svg>
                <span>QR</span>
              </Button>
            </template>
          </div>
        </template>
      </DataTable>
    </div>

    <!-- Modal Konfirmasi Penerbitan Dokumen -->
    <div
      v-if="issueModalOpen"
      class="fixed inset-0 z-50 bg-slate-900/60 backdrop-blur-xs flex items-center justify-center p-4"
    >
      <div class="bg-white rounded-2xl max-w-md w-full p-6 shadow-xl border border-slate-100 space-y-4">
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 rounded-full bg-red-100 text-red-600 flex items-center justify-center flex-shrink-0">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4M7.835 4.697a3.42 3.42 0 001.946-.806 3.42 3.42 0 014.438 0 3.42 3.42 0 001.946.806 3.42 3.42 0 013.138 3.138 3.42 3.42 0 00.806 1.946 3.42 3.42 0 010 4.438 3.42 3.42 0 00-.806 1.946 3.42 3.42 0 01-3.138 3.138 3.42 3.42 0 00-1.946.806 3.42 3.42 0 01-4.438 0 3.42 3.42 0 00-1.946-.806 3.42 3.42 0 01-3.138-3.138 3.42 3.42 0 00-.806-1.946 3.42 3.42 0 010-4.438 3.42 3.42 0 00.806-1.946 3.42 3.42 0 013.138-3.138z" />
            </svg>
          </div>
          <div>
            <h4 class="text-base font-bold text-slate-900">Terbitkan Dokumen Resmi SIMAK</h4>
            <p class="text-xs text-slate-500">Pemohon: {{ selectedItem?.worker_name }}</p>
          </div>
        </div>

        <div class="text-xs text-slate-600 space-y-3">
          <p class="leading-relaxed">
            Sistem akan menerbitkan nomor SIMAK resmi permanen, membuat tautan verifikasi kriptografis, dan mengunci daftar peralatan kerja.
          </p>

          <div class="space-y-1">
            <label class="font-semibold text-slate-700">Nomor SIMAK Khusus (Opsional)</label>
            <input
              v-model="customNumber"
              type="text"
              placeholder="Kosongkan untuk nomor otomatis SIMAK/INKA/2026/XXXXXX"
              class="w-full p-2.5 bg-slate-50 border border-slate-200 rounded-lg text-xs font-mono focus:outline-none focus:ring-2 focus:ring-red-500/20 focus:border-red-500"
            />
          </div>

          <div class="space-y-1">
            <label class="font-semibold text-slate-700">Catatan Penerbitan</label>
            <textarea
              v-model="issueNotes"
              rows="2"
              class="w-full p-2.5 bg-slate-50 border border-slate-200 rounded-lg text-xs focus:outline-none focus:ring-2 focus:ring-red-500/20 focus:border-red-500"
            ></textarea>
          </div>
        </div>

        <div class="flex items-center justify-end gap-2 pt-2 border-t border-slate-200">
          <Button variant="outline" size="sm" :disabled="issuing" @click="issueModalOpen = false">
            Batal
          </Button>
          <Button variant="primary" size="sm" :loading="issuing" @click="handleConfirmIssue">
            Terbitkan Sekarang
          </Button>
        </div>
      </div>
    </div>

    <!-- Modal QR Code Viewer -->
    <div
      v-if="qrModalOpen"
      class="fixed inset-0 z-50 bg-slate-900/60 backdrop-blur-xs flex items-center justify-center p-4"
      @click.self="qrModalOpen = false"
    >
      <div class="bg-white rounded-2xl max-w-sm w-full p-6 shadow-xl border border-slate-100 text-center space-y-4">
        <div class="flex items-center justify-between pb-2 border-b border-slate-200">
          <p class="text-xs font-bold uppercase tracking-wider text-slate-500">Kode QR Verifikasi Gerbang</p>
          <button class="text-slate-400 hover:text-slate-600 text-lg leading-none" @click="qrModalOpen = false">×</button>
        </div>

        <div v-if="qrLoading" class="py-12 text-xs text-slate-400">
          Memuat QR Code...
        </div>

        <div v-else-if="activeQrDoc" class="space-y-3">
          <div class="p-3 bg-white border border-slate-200 rounded-xl inline-block shadow-xs">
            <img :src="activeQrDoc.qr_code_data" alt="QR Code SIMAK" class="w-48 h-48 mx-auto" />
          </div>

          <div>
            <p class="font-mono text-xs font-bold text-slate-900">{{ activeQrDoc.document_number }}</p>
            <p class="text-[11px] text-slate-500 mt-0.5">Diterbitkan: {{ formatDate(activeQrDoc.issued_at) }}</p>
          </div>

          <div class="pt-2 flex flex-col gap-2">
            <Button
              variant="outline"
              size="sm"
              class="w-full justify-center gap-1.5 text-xs text-slate-700"
              @click="copyVerifyLink(activeQrDoc.simak?.qr_token)"
            >
              <svg class="w-3.5 h-3.5 text-slate-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
              </svg>
              <span>Salin Tautan Verifikasi</span>
            </Button>

            <Button
              variant="primary"
              size="sm"
              class="w-full justify-center gap-1.5 text-xs"
              @click="router.push(`/simak/${activeQrDoc.simak_id}/preview`)"
            >
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 17h2a2 2 0 002-2v-4a2 2 0 00-2-2H5a2 2 0 00-2 2v4a2 2 0 002 2h2m2 4h6a2 2 0 002-2v-4a2 2 0 00-2-2H9a2 2 0 00-2 2v4a2 2 0 002 2zm8-12V5a2 2 0 00-2-2H9a2 2 0 00-2 2v4h10z" />
              </svg>
              <span>Buka Dokumen PDF Resmi</span>
            </Button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
