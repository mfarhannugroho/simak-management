<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { inspectionService, type InspectionItemInput, type Inspection } from '@/services/inspection'
import { simakService, type Simak } from '@/services/simak'
import { documentService } from '@/services/document'
import Button from '@/components/ui/Button.vue'
import Badge from '@/components/ui/Badge.vue'
import Modal from '@/components/ui/Modal.vue'

interface InspectionRowItem extends InspectionItemInput {
  is_custom?: boolean
  category_name?: string
}

const route = useRoute()
const router = useRouter()

const loading = ref(true)
const submitting = ref(false)
const errorMsg = ref('')

const simakData = ref<Simak | null>(null)
const items = ref<InspectionRowItem[]>([])

// Form state
const selectedGate = ref('Pos Gerbang Utama PT INKA (Madiun)')
const availableGates = [
  'Pos Gerbang Utama PT INKA (Madiun)',
  'Pos 2 Barat (Pintu Kendaraan Berat)',
  'Pos 3 Timur (Akses Pekerja & Logistik)',
  'Pos Jalur Rel KA & Uji Coba Lintasan',
  'Pos Gerbang Workshop Fabrikasi & Finishing',
]

const inspectionType = ref<'MASUK' | 'KELUAR'>('MASUK')
const officerNotes = ref('')

// Completion modal state
const completeModalOpen = ref(false)
const inspectionResult = ref<Inspection | null>(null)

// Summary counts
const totalExpected = computed(() => {
  return items.value.reduce((sum, it) => sum + (it.expected_qty || 0), 0)
})

const totalActual = computed(() => {
  return items.value.reduce((sum, it) => sum + (it.actual_qty || 0), 0)
})

const hasMismatch = computed(() => {
  return items.value.some((it) => it.status === 'MISMATCH' || it.status === 'UNREGISTERED' || it.actual_qty !== it.expected_qty)
})

const overallStatus = computed(() => {
  return hasMismatch.value ? 'MISMATCH' : 'MATCH'
})

async function loadData() {
  loading.value = true
  errorMsg.value = ''

  const tokenParam = route.query.token as string | undefined
  const qParam = route.query.q as string | undefined
  const idParam = route.params.simakId as string | undefined

  try {
    let resolvedSimak: Simak | null = null

    // 1. If QR Token provided, verify and get SIMAK
    if (tokenParam) {
      const qrRes = await documentService.verifyQRToken(tokenParam)
      if (!qrRes.data || !qrRes.data.valid) {
        errorMsg.value = qrRes.data?.reason || 'Token QR Code tidak valid atau izin telah kedaluwarsa.'
        loading.value = false
        return
      }

      // Try searching via simak list with the simak_number or load directly
      if (qrRes.data.simak_number) {
        const listRes = await simakService.getSimaks({ search: qrRes.data.simak_number, limit: 1 })
        if (listRes.data && listRes.data.items.length > 0) {
          const matched = listRes.data.items[0]
          const fullRes = await simakService.getSimak(matched.id)
          resolvedSimak = fullRes.data || matched
        }
      }
    }

    // 2. If ID param or Q param provided
    const targetId = idParam || qParam
    if (!resolvedSimak && targetId) {
      try {
        const res = await simakService.getSimak(targetId)
        if (res.data) {
          resolvedSimak = res.data
        }
      } catch {
        // Try searching by number
        const listRes = await simakService.getSimaks({ search: targetId, limit: 1 })
        if (listRes.data && listRes.data.items.length > 0) {
          const matched = listRes.data.items[0]
          const fullRes = await simakService.getSimak(matched.id)
          resolvedSimak = fullRes.data || matched
        }
      }
    }

    if (!resolvedSimak) {
      errorMsg.value = 'Dokumen SIMAK tidak ditemukan. Pastikan nomor atau token QR yang dimasukkan benar.'
      loading.value = false
      return
    }

    simakData.value = resolvedSimak

    // Default inspection direction based on SIMAK direction
    if (resolvedSimak.direction === 'KELUAR') {
      inspectionType.value = 'KELUAR'
    } else {
      inspectionType.value = 'MASUK'
    }

    // Populate checklist items from registered SIMAK items
    if (resolvedSimak.items && resolvedSimak.items.length > 0) {
      items.value = resolvedSimak.items.map((it) => ({
        simak_item_id: it.id,
        item_name: it.item_name,
        brand: it.brand || '',
        serial_number: it.serial_number || '',
        expected_qty: it.quantity,
        actual_qty: it.quantity, // Default to expected qty for ease of guard
        unit: it.unit || 'Unit',
        status: 'MATCH',
        notes: '',
        is_custom: false,
        category_name: it.category?.name || 'Peralatan Kerja',
      }))
    } else {
      items.value = []
    }
  } catch (err: any) {
    console.error('Failed to load SIMAK for inspection:', err)
    errorMsg.value = err.response?.data?.message || 'Gagal memuat rincian dokumen SIMAK.'
  } finally {
    loading.value = false
  }
}

// Stepper operations with auto-status recalculation
function incrementQty(item: InspectionRowItem) {
  item.actual_qty++
  recalcItemStatus(item)
}

function decrementQty(item: InspectionRowItem) {
  if (item.actual_qty > 0) {
    item.actual_qty--
    recalcItemStatus(item)
  }
}

function setManualQty(item: InspectionRowItem, event: Event) {
  const input = event.target as HTMLInputElement
  const val = parseInt(input.value)
  item.actual_qty = isNaN(val) || val < 0 ? 0 : val
  recalcItemStatus(item)
}

function recalcItemStatus(item: InspectionRowItem) {
  if (item.is_custom) {
    item.status = 'UNREGISTERED'
  } else if (item.actual_qty === item.expected_qty) {
    item.status = 'MATCH'
  } else {
    item.status = 'MISMATCH'
  }
}

// Quick action: Match all registered items
function matchAll() {
  items.value.forEach((it) => {
    if (!it.is_custom) {
      it.actual_qty = it.expected_qty
      it.status = 'MATCH'
    }
  })
}

// Add unregistered/custom item brought by worker
function addCustomItem() {
  items.value.push({
    item_name: '',
    brand: '',
    serial_number: '',
    expected_qty: 0,
    actual_qty: 1,
    unit: 'Unit',
    status: 'UNREGISTERED',
    notes: 'Temuan fisik di pos gerbang (tidak ada di izin SIMAK)',
    is_custom: true,
  })
}

function removeCustomItem(index: number) {
  items.value.splice(index, 1)
}

async function handleFinishInspection() {
  if (!simakData.value) return

  // Validate items
  for (const it of items.value) {
    if (!it.item_name.trim()) {
      alert('Nama seluruh peralatan harus terisi.')
      return
    }
  }

  submitting.value = true
  try {
    const payload = {
      simak_id: simakData.value.id,
      gate_name: selectedGate.value,
      inspection_type: inspectionType.value,
      notes: officerNotes.value.trim(),
      items: items.value.map((it) => ({
        simak_item_id: it.simak_item_id,
        item_name: it.item_name,
        brand: it.brand,
        serial_number: it.serial_number,
        expected_qty: it.expected_qty,
        actual_qty: it.actual_qty,
        unit: it.unit || 'Unit',
        status: it.status,
        notes: it.notes,
        photo_url: it.photo_url,
      })),
    }

    const res = await inspectionService.recordInspection(payload)
    if (res.data) {
      inspectionResult.value = res.data
      completeModalOpen.value = true
    }
  } catch (err: any) {
    console.error('Failed to submit inspection:', err)
    alert(err.response?.data?.message || 'Gagal menyimpan hasil verifikasi pos gerbang.')
  } finally {
    submitting.value = false
  }
}

function formatDate(dateStr?: string) {
  if (!dateStr) return '-'
  const d = new Date(dateStr)
  return d.toLocaleDateString('id-ID', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
  })
}

onMounted(() => {
  loadData()
})
</script>

<template>
  <div class="space-y-6 max-w-5xl mx-auto pb-24">
    <!-- Breadcrumb & Top Bar -->
    <div class="flex items-center justify-between gap-4">
      <div class="flex items-center gap-2">
        <router-link to="/gate" class="text-sm font-semibold text-slate-500 hover:text-inka-600 flex items-center gap-1">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
          </svg>
          Kembali ke Dashboard Gate
        </router-link>
      </div>

      <div class="flex items-center gap-2">
        <span class="px-2 py-0.5 text-xs font-semibold rounded bg-slate-100 text-slate-700">
          Moda Lapangan Pos Jaga
        </span>
      </div>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="bg-white border rounded-xl p-12 text-center shadow-sm">
      <div class="inline-block w-8 h-8 border-3 border-inka-600 border-t-transparent rounded-full animate-spin"></div>
      <p class="text-sm font-semibold text-slate-700 mt-3">Memverifikasi Dokumen & Memuat Checklist...</p>
      <p class="text-xs text-slate-400 mt-1">Mengambil data permohonan SIMAK resmi PT INKA</p>
    </div>

    <!-- Error State -->
    <div v-else-if="errorMsg" class="bg-white border border-rose-200 rounded-xl p-8 text-center shadow-sm">
      <div class="w-14 h-14 bg-rose-100 text-rose-600 rounded-full flex items-center justify-center mx-auto mb-4">
        <svg class="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
        </svg>
      </div>
      <h2 class="text-xl font-bold text-slate-900">Verifikasi SIMAK Tidak Berhasil</h2>
      <p class="text-sm text-rose-600 font-medium mt-1 max-w-md mx-auto">{{ errorMsg }}</p>
      <div class="mt-6 flex justify-center gap-3">
        <router-link to="/gate">
          <Button variant="secondary">Cari SIMAK Lain</Button>
        </router-link>
      </div>
    </div>

    <!-- Main Inspection Form -->
    <div v-else-if="simakData" class="space-y-6">
      <!-- 1. Header Validation Banner -->
      <div class="bg-white border border-slate-200 rounded-2xl shadow-sm overflow-hidden">
        <div class="bg-gradient-to-r from-slate-900 to-slate-800 text-white p-5 sm:p-6 flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div>
            <div class="flex items-center gap-2 flex-wrap">
              <span class="text-xs uppercase tracking-wider font-bold text-inka-400">SURAT IZIN MEMBAWA ALAT KERJA</span>
              <Badge variant="inka" size="sm">
                {{ simakData.status }}
              </Badge>
            </div>
            <h1 class="text-xl sm:text-2xl font-bold tracking-tight text-white mt-1">
              {{ simakData.simak_number || 'SIMAK/INKA/--/------' }}
            </h1>
          </div>

          <div class="text-left sm:text-right bg-slate-800/60 sm:bg-transparent p-3 sm:p-0 rounded-xl border sm:border-0 border-slate-700">
            <div class="text-xs text-slate-400">Masa Berlaku Izin:</div>
            <div class="text-sm font-bold text-emerald-400">
              {{ formatDate(simakData.start_date) }} s/d {{ formatDate(simakData.end_date) }}
            </div>
          </div>
        </div>

        <!-- Worker & Destination Meta -->
        <div class="p-5 sm:p-6 grid grid-cols-1 sm:grid-cols-3 gap-4 text-xs border-b border-slate-100 bg-slate-50/50">
          <div>
            <div class="text-slate-500 font-medium">Pemohon / Pembawa Alat:</div>
            <div class="font-bold text-slate-900 text-sm mt-0.5">
              {{ simakData.worker?.full_name || 'Pekerja' }}
            </div>
            <div class="text-slate-500 mt-0.5">
              NIK: {{ simakData.worker?.worker_id || '-' }} • {{ simakData.worker?.phone || '-' }}
            </div>
          </div>

          <div>
            <div class="text-slate-500 font-medium">Instansi / Vendor:</div>
            <div class="font-bold text-slate-900 text-sm mt-0.5">
              {{ simakData.worker?.company?.name || 'PT INKA (Persero)' }}
            </div>
            <div class="text-slate-500 mt-0.5">
              Departemen: {{ simakData.worker?.department?.name || 'Unit Operasional' }}
            </div>
          </div>

          <div>
            <div class="text-slate-500 font-medium">Area Pabrik & Keperluan:</div>
            <div class="font-bold text-slate-900 text-sm mt-0.5 flex items-center gap-1.5">
              <svg class="w-4 h-4 text-slate-500 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
              </svg>
              <span>{{ simakData.destination_area }}</span>
            </div>
            <div class="text-slate-600 mt-0.5">
              "{{ simakData.purpose }}"
            </div>
          </div>
        </div>
      </div>

      <!-- 2. Inspection Configuration (Pos & Direction) -->
      <div class="bg-white border border-slate-200 rounded-2xl p-5 sm:p-6 shadow-sm space-y-4">
        <h2 class="text-base font-bold text-slate-900 flex items-center gap-2">
          <svg class="w-4 h-4 text-inka-700" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
          </svg>
          <span>Parameter Pos Pemeriksaan</span>
        </h2>

        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <!-- Gate Selection -->
          <div>
            <label class="block text-xs font-bold text-slate-700 uppercase tracking-wider mb-1.5">
              Lokasi Pos Jaga Pemeriksa:
            </label>
            <select
              v-model="selectedGate"
              class="w-full px-3.5 py-3 rounded-xl border border-slate-300 bg-white text-slate-800 text-sm font-semibold focus:ring-2 focus:ring-inka-500 focus:outline-none"
            >
              <option v-for="gate in availableGates" :key="gate" :value="gate">
                {{ gate }}
              </option>
            </select>
          </div>

          <!-- Direction Selection (Big Touch Segmented Buttons) -->
          <div>
            <label class="block text-xs font-bold text-slate-700 uppercase tracking-wider mb-1.5">
              Arah Pergerakan Fisik:
            </label>
            <div class="grid grid-cols-2 gap-2">
              <button
                type="button"
                @click="inspectionType = 'MASUK'"
                class="py-3 px-4 rounded-xl border font-bold text-sm flex items-center justify-center gap-2 transition-all cursor-pointer min-h-[48px]"
                :class="
                  inspectionType === 'MASUK'
                    ? 'bg-emerald-600 text-white border-emerald-600 shadow-md shadow-emerald-700/20'
                    : 'bg-slate-50 text-slate-700 border-slate-200 hover:bg-slate-100'
                "
              >
                <svg class="w-5 h-5 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M11 16l-4-4m0 0l4-4m-4 4h14m-5 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h7a3 3 0 013 3v1" />
                </svg>
                <span>MASUK PABRIK</span>
              </button>

              <button
                type="button"
                @click="inspectionType = 'KELUAR'"
                class="py-3 px-4 rounded-xl border font-bold text-sm flex items-center justify-center gap-2 transition-all cursor-pointer min-h-[48px]"
                :class="
                  inspectionType === 'KELUAR'
                    ? 'bg-blue-600 text-white border-blue-600 shadow-md shadow-blue-700/20'
                    : 'bg-slate-50 text-slate-700 border-slate-200 hover:bg-slate-100'
                "
              >
                <svg class="w-5 h-5 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
                </svg>
                <span>KELUAR PABRIK</span>
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- 3. Physical Checklist Section -->
      <div class="bg-white border border-slate-200 rounded-2xl shadow-sm overflow-hidden">
        <div class="p-5 sm:p-6 border-b border-slate-200 flex flex-col sm:flex-row sm:items-center justify-between gap-3">
          <div>
            <h2 class="text-base font-bold text-slate-900">Checklist Pemeriksaan Fisik Barang</h2>
            <p class="text-xs text-slate-500 mt-0.5">
              Tekan tombol (+) / (-) atau tombol cepat untuk mencocokkan jumlah aktual fisik yang dibawa pekerja.
            </p>
          </div>

          <div class="flex items-center gap-2">
            <button
              type="button"
              @click="matchAll"
              class="px-3.5 py-2 rounded-xl bg-emerald-50 hover:bg-emerald-100 text-emerald-700 border border-emerald-300 font-bold text-xs flex items-center gap-1.5 transition-colors cursor-pointer"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 13l4 4L19 7" />
              </svg>
              <span>Cocokkan Semua (Quick Match)</span>
            </button>
          </div>
        </div>

        <!-- Checklist Item Rows -->
        <div class="divide-y divide-slate-100">
          <div
            v-for="(item, index) in items"
            :key="index"
            class="p-4 sm:p-5 transition-colors"
            :class="[
              item.status === 'MATCH'
                ? 'hover:bg-slate-50/70'
                : 'bg-rose-50/60 border-l-4 border-l-rose-500',
            ]"
          >
            <div class="flex flex-col lg:flex-row lg:items-center justify-between gap-4">
              <!-- Item Details -->
              <div class="flex-1 min-w-0">
                <div class="flex items-center gap-2 flex-wrap mb-1">
                  <span class="w-6 h-6 rounded-full bg-slate-100 text-slate-700 flex items-center justify-center text-xs font-bold shrink-0">
                    {{ index + 1 }}
                  </span>

                  <!-- Name -->
                  <div v-if="!item.is_custom" class="font-bold text-slate-900 text-sm sm:text-base">
                    {{ item.item_name }}
                  </div>
                  <input
                    v-else
                    v-model="item.item_name"
                    type="text"
                    placeholder="Nama barang temuan / tidak terdaftar..."
                    class="font-bold text-slate-900 text-sm border border-slate-300 rounded px-2 py-1 flex-1 min-w-[200px]"
                  />

                  <!-- Status Badge -->
                  <Badge
                    :variant="item.status === 'MATCH' ? 'success' : 'danger'"
                    :dot="true"
                  >
                    {{ item.status === 'MATCH' ? 'COCOK (MATCH)' : item.status === 'UNREGISTERED' ? 'TIDAK TERDAFTAR' : 'SELISIH (MISMATCH)' }}
                  </Badge>

                  <span v-if="item.is_custom" class="text-xs bg-amber-100 text-amber-800 font-semibold px-2 py-0.5 rounded">
                    Temuan Ekstra
                  </span>
                </div>

                <!-- Specs / Serial / Brand -->
                <div class="text-xs text-slate-600 flex items-center gap-3 flex-wrap pl-8">
                  <span v-if="item.brand"><strong>Merk:</strong> {{ item.brand }}</span>
                  <span v-if="item.serial_number"><strong>S/N:</strong> {{ item.serial_number }}</span>
                  <span><strong>Satuan:</strong> {{ item.unit }}</span>
                  <span v-if="item.category_name"><strong>Kategori:</strong> {{ item.category_name }}</span>
                </div>

                <!-- Custom Item Inputs if Extra -->
                <div v-if="item.is_custom" class="grid grid-cols-2 sm:grid-cols-3 gap-2 pl-8 mt-2">
                  <input
                    v-model="item.brand"
                    type="text"
                    placeholder="Merk / Brand"
                    class="text-xs px-2 py-1 border border-slate-200 rounded"
                  />
                  <input
                    v-model="item.serial_number"
                    type="text"
                    placeholder="No. Seri / Serial No"
                    class="text-xs px-2 py-1 border border-slate-200 rounded"
                  />
                  <input
                    v-model="item.unit"
                    type="text"
                    placeholder="Satuan (Pcs/Unit/Set)"
                    class="text-xs px-2 py-1 border border-slate-200 rounded"
                  />
                </div>

                <!-- Item Note Field if Mismatch -->
                <div v-if="item.status !== 'MATCH'" class="pl-8 mt-2">
                  <input
                    v-model="item.notes"
                    type="text"
                    placeholder="Alasan selisih / catatan kondisi barang..."
                    class="w-full text-xs px-3 py-1.5 border border-rose-300 rounded-lg bg-white placeholder-slate-400 focus:outline-none focus:ring-1 focus:ring-rose-500"
                  />
                </div>
              </div>

              <!-- Stepper & Comparison Controls (Touch targets min 44x44px) -->
              <div class="flex items-center justify-between sm:justify-end gap-5 pl-8 lg:pl-0 shrink-0">
                <!-- Expected Qty -->
                <div class="text-center">
                  <div class="text-[11px] font-semibold text-slate-500 uppercase">Terdaftar</div>
                  <div class="text-base font-bold text-slate-700 mt-0.5">
                    {{ item.expected_qty }} <span class="text-xs font-normal">{{ item.unit }}</span>
                  </div>
                </div>

                <div class="text-slate-300 font-light text-lg">vs</div>

                <!-- Touch Stepper for Actual Qty -->
                <div>
                  <div class="text-[11px] font-semibold text-slate-500 uppercase text-center mb-1">
                    Aktual Fisik
                  </div>
                  <div class="flex items-center gap-1.5">
                    <!-- Decrement Button -->
                    <button
                      type="button"
                      @click="decrementQty(item)"
                      class="w-11 h-11 rounded-xl bg-slate-100 hover:bg-slate-200 active:bg-slate-300 text-slate-800 font-bold text-xl flex items-center justify-center transition-colors cursor-pointer border border-slate-200 select-none shadow-sm"
                      title="Kurangi 1"
                    >
                      −
                    </button>

                    <!-- Actual Qty Input / Display -->
                    <input
                      type="number"
                      :value="item.actual_qty"
                      @change="setManualQty(item, $event)"
                      min="0"
                      class="w-16 h-11 text-center font-bold text-base rounded-xl border border-slate-300 bg-white text-slate-900 focus:outline-none focus:ring-2 focus:ring-inka-500 shadow-inner"
                    />

                    <!-- Increment Button -->
                    <button
                      type="button"
                      @click="incrementQty(item)"
                      class="w-11 h-11 rounded-xl bg-slate-100 hover:bg-slate-200 active:bg-slate-300 text-slate-800 font-bold text-xl flex items-center justify-center transition-colors cursor-pointer border border-slate-200 select-none shadow-sm"
                      title="Tambah 1"
                    >
                      +
                    </button>
                  </div>
                </div>

                <!-- Remove Custom Item Button -->
                <div v-if="item.is_custom" class="pt-5">
                  <button
                    type="button"
                    @click="removeCustomItem(index)"
                    class="p-2 text-rose-500 hover:bg-rose-50 rounded-lg cursor-pointer transition-colors"
                    title="Hapus barang temuan"
                  >
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                    </svg>
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Add Unregistered Item Button -->
        <div class="p-4 bg-slate-50 border-t border-slate-200 flex justify-center">
          <button
            type="button"
            @click="addCustomItem"
            class="px-4 py-2.5 rounded-xl border-2 border-dashed border-slate-300 hover:border-inka-400 text-slate-700 hover:text-inka-700 font-bold text-xs flex items-center gap-2 transition-colors cursor-pointer"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
            </svg>
            <span>+ Tambah Barang Temuan Fisik Tidak Terdaftar</span>
          </button>
        </div>
      </div>

      <!-- 4. Gate Officer Notes -->
      <div class="bg-white border border-slate-200 rounded-2xl p-5 sm:p-6 shadow-sm">
        <label class="block text-xs font-bold text-slate-700 uppercase tracking-wider mb-2">
          Catatan Verifikasi Petugas Pos Jaga (Opsional):
        </label>
        <textarea
          v-model="officerNotes"
          rows="3"
          placeholder="Catatan kondisi fisik, nomor segel, pengemasan, atau alasan khusus jika terdapat selisih jumlah barang..."
          class="w-full px-3.5 py-2.5 rounded-xl border border-slate-300 text-sm focus:outline-none focus:ring-2 focus:ring-inka-500 placeholder-slate-400"
        ></textarea>
      </div>
    </div>

    <!-- 5. Floating / Sticky Bottom Action Bar (Tablet & Mobile Optimized) -->
    <div
      v-if="simakData"
      class="fixed bottom-0 inset-x-0 bg-white/95 backdrop-blur-md border-t border-slate-300 shadow-2xl p-3 sm:p-4 z-40"
    >
      <div class="max-w-5xl mx-auto flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <!-- Status Indicator & Item Counter -->
        <div class="flex items-center gap-3">
          <div
            class="w-10 h-10 rounded-xl flex items-center justify-center shrink-0 font-bold"
            :class="hasMismatch ? 'bg-rose-100 text-rose-700' : 'bg-emerald-100 text-emerald-700'"
          >
            <svg v-if="hasMismatch" class="w-5 h-5 text-rose-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
            </svg>
            <svg v-else class="w-5 h-5 text-emerald-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
            </svg>
          </div>

          <div>
            <div class="text-xs font-bold" :class="hasMismatch ? 'text-rose-700' : 'text-emerald-700'">
              {{ hasMismatch ? 'STATUS: TERDETEKSI SELISIH BARANG (MISMATCH)' : 'STATUS: SEMUA BARANG COCOK (MATCH)' }}
            </div>
            <div class="text-xs text-slate-500 font-medium">
              Total Fisik: <strong>{{ totalActual }}</strong> dari <strong>{{ totalExpected }}</strong> alat terdaftar
            </div>
          </div>
        </div>

        <!-- Submit Button -->
        <div class="flex items-center gap-2">
          <Button
            variant="primary"
            size="lg"
            :disabled="submitting || items.length === 0"
            @click="handleFinishInspection"
            class="w-full sm:w-auto min-h-[48px] px-8 text-base font-bold shadow-lg shadow-inka-900/20"
          >
            <span v-if="submitting" class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin mr-2"></span>
            <span>Selesaikan & Catat Pemeriksaan</span>
          </Button>
        </div>
      </div>
    </div>

    <!-- Success Result Modal -->
    <Modal
      v-model="completeModalOpen"
      title="Pemeriksaan Pos Gerbang Berhasil Disimpan"
      @close="completeModalOpen = false"
    >
      <div v-if="inspectionResult" class="space-y-4 text-center">
        <!-- Big Success Icon -->
        <div
          class="w-16 h-16 rounded-full flex items-center justify-center mx-auto"
          :class="inspectionResult.status === 'MATCH' ? 'bg-emerald-100 text-emerald-600' : 'bg-amber-100 text-amber-600'"
        >
          <svg v-if="inspectionResult.status === 'MATCH'" class="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 13l4 4L19 7" />
          </svg>
          <svg v-else class="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
          </svg>
        </div>

        <div>
          <h3 class="text-lg font-bold text-slate-900">
            Pergerakan Fisik {{ inspectionResult.inspection_type }} Berhasil Dicatat
          </h3>
          <p class="text-xs text-slate-500 mt-1">
            Data pemeriksaan telah disimpan ke dalam log operasional pos jaga PT INKA.
          </p>
        </div>

        <!-- Receipt Card -->
        <div class="bg-slate-50 border border-slate-200 rounded-xl p-4 text-left space-y-2 text-xs">
          <div class="flex justify-between border-b border-slate-200/60 pb-1.5">
            <span class="text-slate-500">Nomor SIMAK:</span>
            <span class="font-bold text-slate-900">{{ simakData?.simak_number }}</span>
          </div>
          <div class="flex justify-between border-b border-slate-200/60 pb-1.5">
            <span class="text-slate-500">Pos Gerbang:</span>
            <span class="font-bold text-slate-900">{{ inspectionResult.gate_name }}</span>
          </div>
          <div class="flex justify-between border-b border-slate-200/60 pb-1.5">
            <span class="text-slate-500">Arah Pergerakan:</span>
            <span class="font-bold text-slate-900">{{ inspectionResult.inspection_type }}</span>
          </div>
          <div class="flex justify-between border-b border-slate-200/60 pb-1.5">
            <span class="text-slate-500">Hasil Rekonsiliasi:</span>
            <Badge :variant="inspectionResult.status === 'MATCH' ? 'success' : 'danger'" size="sm">
              {{ inspectionResult.status }}
            </Badge>
          </div>
          <div class="flex justify-between">
            <span class="text-slate-500">Total Alat Fisik:</span>
            <span class="font-bold text-slate-900">
              {{ inspectionResult.total_actual_items }} dari {{ inspectionResult.total_registered_items }} Unit
            </span>
          </div>
        </div>

        <!-- Action Buttons -->
        <div class="pt-2 flex flex-col sm:flex-row gap-2">
          <router-link to="/gate" class="flex-1">
            <Button variant="secondary" class="w-full">
              Kembali ke Dashboard Gate
            </Button>
          </router-link>
          <router-link to="/gate/history" class="flex-1">
            <Button variant="primary" class="w-full">
              Lihat Log Riwayat
            </Button>
          </router-link>
        </div>
      </div>
    </Modal>
  </div>
</template>
