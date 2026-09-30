<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import {
  discrepancyService,
  type DiscrepancyListItem,
  type DiscrepancyMetrics,
  type DiscrepancyStatus,
  type DiscrepancyType,
} from '@/services/discrepancy'
import DataTable, { type Column, type Pagination } from '@/components/ui/DataTable.vue'
import Button from '@/components/ui/Button.vue'
import Badge from '@/components/ui/Badge.vue'
import StatCard from '@/components/ui/StatCard.vue'

const router = useRouter()

const items = ref<DiscrepancyListItem[]>([])
const loading = ref(false)
const search = ref('')

const activeStatusTab = ref<string>('ALL')
const selectedType = ref<string>('ALL')
const selectedDate = ref<string>('')

const metrics = ref<DiscrepancyMetrics>({
  total: 0,
  open: 0,
  in_review: 0,
  resolved: 0,
  closed: 0,
})

const pagination = ref<Pagination>({
  page: 1,
  limit: 10,
  total_items: 0,
  total_pages: 1,
})

const statusTabs = [
  { key: 'ALL', label: 'Semua Kasus' },
  { key: 'OPEN', label: 'Butuh Tindakan (Open)', countKey: 'open' },
  { key: 'IN_REVIEW', label: 'Dalam Investigasi', countKey: 'in_review' },
  { key: 'RESOLVED', label: 'Selesai (Resolved)', countKey: 'resolved' },
  { key: 'CLOSED', label: 'Ditutup (Closed)', countKey: 'closed' },
]

const discrepancyTypes: { key: string; label: string }[] = [
  { key: 'ALL', label: 'Semua Tipe Temuan' },
  { key: 'QUANTITY_MISMATCH', label: 'Selisih Kuantitas (Kelebihan)' },
  { key: 'UNREGISTERED_ITEM', label: 'Barang Tidak Terdaftar' },
  { key: 'MISSING_ITEM', label: 'Barang Kurang / Hilang' },
  { key: 'ITEM_DESCRIPTION_MISMATCH', label: 'Spesifikasi / Serial Beda' },
  { key: 'OTHER', label: 'Lainnya' },
]

const columns: Column[] = [
  { key: 'discrepancy_number', label: 'No. Kasus & SIMAK', sortable: true },
  { key: 'item', label: 'Peralatan Terkait' },
  { key: 'type', label: 'Tipe Temuan', align: 'center' },
  { key: 'qty', label: 'Komparasi Fisik', align: 'center' },
  { key: 'worker', label: 'Pekerja & Perusahaan' },
  { key: 'status', label: 'Status Penanganan', align: 'center' },
  { key: 'created_at', label: 'Waktu Temuan', sortable: true },
  { key: 'actions', label: 'Aksi', align: 'right' },
]

async function fetchMetrics() {
  try {
    const res = await discrepancyService.getMetrics()
    if (res.data) {
      metrics.value = res.data
    }
  } catch (err) {
    console.error('Failed to load discrepancy metrics:', err)
  }
}

async function fetchDiscrepancies(page = 1) {
  loading.value = true
  try {
    const res = await discrepancyService.getDiscrepancies({
      page,
      limit: pagination.value.limit,
      search: search.value,
      status: activeStatusTab.value === 'ALL' ? '' : activeStatusTab.value,
      discrepancy_type: selectedType.value === 'ALL' ? '' : selectedType.value,
      date: selectedDate.value,
    })

    if (res.data) {
      items.value = res.data.items || []
      if (res.data.pagination) {
        pagination.value = res.data.pagination
      }
    }
  } catch (err) {
    console.error('Failed to load discrepancies list:', err)
  } finally {
    loading.value = false
  }
}

function handleTabChange(tabKey: string) {
  activeStatusTab.value = tabKey
  pagination.value.page = 1
  fetchDiscrepancies(1)
}

function handleFilterChange() {
  pagination.value.page = 1
  fetchDiscrepancies(1)
}

function handleSearch(query: string) {
  search.value = query
  pagination.value.page = 1
  fetchDiscrepancies(1)
}

function handlePageChange(page: number) {
  pagination.value.page = page
  fetchDiscrepancies(page)
}

function resetFilters() {
  search.value = ''
  activeStatusTab.value = 'ALL'
  selectedType.value = 'ALL'
  selectedDate.value = ''
  pagination.value.page = 1
  fetchDiscrepancies(1)
}

function goToDetail(id: string) {
  router.push(`/discrepancies/${id}`)
}

function formatTypeLabel(type: DiscrepancyType) {
  switch (type) {
    case 'QUANTITY_MISMATCH':
      return 'Selisih Kuantitas'
    case 'UNREGISTERED_ITEM':
      return 'Tidak Terdaftar'
    case 'MISSING_ITEM':
      return 'Barang Kurang'
    case 'ITEM_DESCRIPTION_MISMATCH':
      return 'Spesifikasi Beda'
    default:
      return 'Temuan Lainnya'
  }
}

function getStatusVariant(status: DiscrepancyStatus): 'danger' | 'warning' | 'success' | 'neutral' {
  switch (status) {
    case 'OPEN':
      return 'danger'
    case 'IN_REVIEW':
      return 'warning'
    case 'RESOLVED':
      return 'success'
    case 'CLOSED':
    default:
      return 'neutral'
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
  fetchMetrics()
  fetchDiscrepancies()
})
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div>
        <div class="flex items-center gap-2">
          <span class="px-2 py-0.5 text-xs font-semibold uppercase tracking-wider rounded bg-rose-100 text-rose-800">
            Monitoring Selisih
          </span>
          <span class="text-xs text-slate-500 font-medium">Pengawasan Mutu & Keamanan Gerbang</span>
        </div>
        <h1 class="text-2xl font-bold text-slate-900 tracking-tight mt-1">
          Discrepancy Log (Temuan Ketidaksesuaian Barang)
        </h1>
        <p class="text-sm text-slate-600 mt-0.5">
          Daftar selisih kuantitas fisik, barang tidak terdaftar, dan insiden gerbang yang memerlukan investigasi serta resolusi Supervisor.
        </p>
      </div>

      <div class="flex items-center gap-2">
        <Button variant="secondary" @click="() => { fetchMetrics(); fetchDiscrepancies(1); }">
          <svg class="w-4 h-4 mr-1.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
          Segarkan
        </Button>
      </div>
    </div>

    <!-- KPI Summary Cards -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <StatCard
        label="Total Kasus Temuan"
        :value="metrics.total"
        description="Seluruh catatan selisih fisik di pos gerbang"
      />

      <div class="bg-white rounded-xl border border-rose-200 p-5 shadow-card relative transition-all bg-rose-50/20">
        <div class="flex items-center justify-between">
          <span class="text-xs font-semibold uppercase tracking-wider text-rose-700">
            Butuh Tindakan (Open)
          </span>
          <div class="w-8 h-8 rounded-lg bg-rose-100 flex items-center justify-center text-rose-600">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
            </svg>
          </div>
        </div>
        <div class="mt-3 flex items-baseline gap-2">
          <span class="text-3xl font-bold text-rose-700 tracking-tight">{{ metrics.open }}</span>
        </div>
        <p class="text-xs text-rose-600 mt-1 font-medium">Menunggu investigasi Supervisor</p>
      </div>

      <div class="bg-white rounded-xl border border-amber-200 p-5 shadow-card relative transition-all bg-amber-50/20">
        <div class="flex items-center justify-between">
          <span class="text-xs font-semibold uppercase tracking-wider text-amber-700">
            Dalam Investigasi (In Review)
          </span>
          <div class="w-8 h-8 rounded-lg bg-amber-100 flex items-center justify-center text-amber-600">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
          </div>
        </div>
        <div class="mt-3 flex items-baseline gap-2">
          <span class="text-3xl font-bold text-amber-700 tracking-tight">{{ metrics.in_review }}</span>
        </div>
        <p class="text-xs text-amber-600 mt-1 font-medium">Sedang ditindaklanjuti pengawas</p>
      </div>

      <div class="bg-white rounded-xl border border-emerald-200 p-5 shadow-card relative transition-all bg-emerald-50/20">
        <div class="flex items-center justify-between">
          <span class="text-xs font-semibold uppercase tracking-wider text-emerald-700">
            Kasus Selesai (Resolved)
          </span>
          <div class="w-8 h-8 rounded-lg bg-emerald-100 flex items-center justify-center text-emerald-600">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
            </svg>
          </div>
        </div>
        <div class="mt-3 flex items-baseline gap-2">
          <span class="text-3xl font-bold text-emerald-700 tracking-tight">{{ metrics.resolved + metrics.closed }}</span>
        </div>
        <p class="text-xs text-emerald-600 mt-1 font-medium">Resolusi disetujui & ditutup</p>
      </div>
    </div>

    <!-- Status Tabs -->
    <div class="flex items-center gap-2 overflow-x-auto pb-1 border-b border-slate-200">
      <button
        v-for="tab in statusTabs"
        :key="tab.key"
        type="button"
        @click="handleTabChange(tab.key)"
        class="px-4 py-2.5 rounded-lg text-xs font-semibold whitespace-nowrap transition-colors flex items-center gap-2 cursor-pointer"
        :class="
          activeStatusTab === tab.key
            ? 'bg-inka-600 text-white shadow-sm'
            : 'bg-white text-slate-600 hover:bg-slate-100 hover:text-slate-900 border border-slate-200'
        "
      >
        <span>{{ tab.label }}</span>
        <span
          v-if="tab.countKey && metrics[tab.countKey as keyof DiscrepancyMetrics] > 0"
          class="px-1.5 py-0.5 rounded-full text-[10px] font-bold"
          :class="activeStatusTab === tab.key ? 'bg-white/20 text-white' : 'bg-rose-100 text-rose-700'"
        >
          {{ metrics[tab.countKey as keyof DiscrepancyMetrics] }}
        </span>
      </button>
    </div>

    <!-- Filter Bar (Type & Date) -->
    <div class="bg-white border border-slate-200 rounded-xl p-3.5 shadow-sm flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-3 text-xs">
      <div class="flex items-center gap-3 flex-wrap">
        <div>
          <label class="block font-semibold text-slate-700 mb-1">Tipe Temuan Ketidaksesuaian:</label>
          <select
            v-model="selectedType"
            @change="handleFilterChange"
            class="px-3 py-1.5 border border-slate-300 rounded-lg bg-white text-slate-800 focus:outline-none focus:ring-1 focus:ring-inka-500 font-medium"
          >
            <option v-for="t in discrepancyTypes" :key="t.key" :value="t.key">
              {{ t.label }}
            </option>
          </select>
        </div>

        <div>
          <label class="block font-semibold text-slate-700 mb-1">Tanggal Temuan:</label>
          <input
            v-model="selectedDate"
            type="date"
            @change="handleFilterChange"
            class="px-3 py-1.5 border border-slate-300 rounded-lg bg-white text-slate-800 focus:outline-none focus:ring-1 focus:ring-inka-500 font-medium"
          />
        </div>

        <button
          v-if="selectedType !== 'ALL' || selectedDate || activeStatusTab !== 'ALL'"
          type="button"
          @click="resetFilters"
          class="mt-4 px-2.5 py-1.5 text-rose-600 hover:bg-rose-50 rounded-lg font-semibold cursor-pointer"
        >
          Reset Filter
        </button>
      </div>

      <div class="text-right text-slate-500">
        Menampilkan <strong>{{ items.length }}</strong> dari <strong>{{ pagination.total_items }}</strong> temuan
      </div>
    </div>

    <!-- Data Table -->
    <DataTable
      :columns="columns"
      :items="items"
      :loading="loading"
      :pagination="pagination"
      search-placeholder="Cari No. Kasus, No. SIMAK, Nama Barang, Pekerja..."
      @search="handleSearch"
      @page-change="handlePageChange"
    >
      <!-- Case & SIMAK Number Cell -->
      <template #cell(discrepancy_number)="{ item }">
        <div>
          <div class="font-bold text-slate-900 text-xs font-mono text-inka-700">
            {{ item.discrepancy_number }}
          </div>
          <div class="text-[11px] text-slate-500 font-mono mt-0.5">
            SIMAK: {{ item.simak_number }}
          </div>
        </div>
      </template>

      <!-- Item Details Cell -->
      <template #cell(item)="{ item }">
        <div>
          <div class="font-semibold text-slate-900 text-xs">{{ item.item_name }}</div>
          <div class="text-[11px] text-slate-500">
            {{ item.brand }} {{ item.serial_number ? `• SN: ${item.serial_number}` : '' }}
          </div>
        </div>
      </template>

      <!-- Type Badge Cell -->
      <template #cell(type)="{ item }">
        <span
          class="px-2 py-0.5 rounded text-[11px] font-bold"
          :class="
            item.discrepancy_type === 'UNREGISTERED_ITEM'
              ? 'bg-purple-50 text-purple-700 border border-purple-200'
              : item.discrepancy_type === 'QUANTITY_MISMATCH'
              ? 'bg-rose-50 text-rose-700 border border-rose-200'
              : 'bg-amber-50 text-amber-700 border border-amber-200'
          "
        >
          {{ formatTypeLabel(item.discrepancy_type) }}
        </span>
      </template>

      <!-- Quantity Comparison Cell -->
      <template #cell(qty)="{ item }">
        <div class="text-xs">
          <span class="font-bold text-rose-600">{{ item.actual_qty }}</span>
          <span class="text-slate-400"> / </span>
          <span class="font-semibold text-slate-700">{{ item.expected_qty }}</span>
          <span class="text-[11px] text-slate-500 ml-1">{{ item.unit }}</span>
        </div>
      </template>

      <!-- Worker & Company Cell -->
      <template #cell(worker)="{ item }">
        <div>
          <div class="font-semibold text-slate-900 text-xs">{{ item.worker_name }}</div>
          <div class="text-[11px] text-slate-500">{{ item.company }}</div>
        </div>
      </template>

      <!-- Status Cell -->
      <template #cell(status)="{ item }">
        <Badge :variant="getStatusVariant(item.status)" :dot="true" size="sm">
          {{ item.status }}
        </Badge>
      </template>

      <!-- Created At Cell -->
      <template #cell(created_at)="{ item }">
        <div class="text-xs text-slate-700 font-medium">
          {{ formatDate(item.created_at) }}
        </div>
      </template>

      <!-- Actions Cell -->
      <template #cell(actions)="{ item }">
        <Button variant="primary" size="sm" @click="goToDetail(item.id)">
          Tinjau Kasus
        </Button>
      </template>
    </DataTable>
  </div>
</template>
