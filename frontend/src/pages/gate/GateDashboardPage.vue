<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { inspectionService, type GateDashboardMetrics, type Inspection } from '@/services/inspection'
import { documentService } from '@/services/document'
import Button from '@/components/ui/Button.vue'
import Badge from '@/components/ui/Badge.vue'
import StatCard from '@/components/ui/StatCard.vue'
import Modal from '@/components/ui/Modal.vue'

const router = useRouter()

const loading = ref(false)
const metrics = ref<GateDashboardMetrics>({
  inspections_today: 0,
  active_simaks: 0,
  mismatch_count: 0,
  recent_movements: [],
})

// Quick Lookup state
const quickQuery = ref('')
const searching = ref(false)
const searchError = ref('')

// QR Scanner Modal state
const qrModalOpen = ref(false)
const manualTokenInput = ref('')
const qrVideoRef = ref<HTMLVideoElement | null>(null)
const scannerActive = ref(false)
let mediaStream: MediaStream | null = null

// Inspection Detail Modal state
const detailModalOpen = ref(false)
const selectedInspection = ref<Inspection | null>(null)
const detailLoading = ref(false)

async function fetchMetrics() {
  loading.value = true
  try {
    const res = await inspectionService.getMetrics()
    if (res.data) {
      metrics.value = res.data
    }
  } catch (err) {
    console.error('Failed to fetch gate metrics:', err)
  } finally {
    loading.value = false
  }
}

async function handleQuickSearch() {
  if (!quickQuery.value.trim()) return

  searching.value = true
  searchError.value = ''

  const query = quickQuery.value.trim()

  try {
    // If it looks like a QR token (uuid or alphanumeric), try verifying as QR token
    if (query.length >= 10 && !query.includes('/')) {
      const res = await documentService.verifyQRToken(query)
      if (res.data && res.data.valid) {
        router.push({ path: '/gate/inspect', query: { token: query } })
        return
      }
    }

    // Try finding by SIMAK number or ID
    router.push({ path: '/gate/inspect', query: { q: query } })
  } catch (err: any) {
    console.error('Search error:', err)
    searchError.value = err.response?.data?.message || 'Dokumen SIMAK tidak ditemukan atau belum berstatus terbit.'
  } finally {
    searching.value = false
  }
}

function openQrScanner() {
  qrModalOpen.value = true
  manualTokenInput.value = ''
  startCamera()
}

function closeQrScanner() {
  stopCamera()
  qrModalOpen.value = false
}

async function startCamera() {
  if (navigator.mediaDevices && navigator.mediaDevices.getUserMedia) {
    try {
      mediaStream = await navigator.mediaDevices.getUserMedia({
        video: { facingMode: 'environment' },
      })
      if (qrVideoRef.value) {
        qrVideoRef.value.srcObject = mediaStream
        scannerActive.value = true
      }
    } catch (err) {
      console.warn('Camera access denied or unavailable:', err)
      scannerActive.value = false
    }
  }
}

function stopCamera() {
  if (mediaStream) {
    mediaStream.getTracks().forEach((track) => track.stop())
    mediaStream = null
  }
  scannerActive.value = false
}

function submitManualToken() {
  if (!manualTokenInput.value.trim()) return
  const token = manualTokenInput.value.trim()
  closeQrScanner()
  router.push({ path: '/gate/inspect', query: { token } })
}

async function viewInspectionDetail(id: string) {
  detailLoading.value = true
  detailModalOpen.value = true
  try {
    const res = await inspectionService.getInspection(id)
    if (res.data) {
      selectedInspection.value = res.data
    }
  } catch (err) {
    console.error('Failed to load inspection detail:', err)
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
  fetchMetrics()
})
</script>

<template>
  <div class="space-y-6">
    <!-- Header Section -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div>
        <div class="flex items-center gap-2">
          <span class="px-2 py-0.5 text-xs font-semibold uppercase tracking-wider rounded bg-slate-100 text-slate-700">
            Operasional Pos
          </span>
          <span class="text-xs text-slate-500 font-medium">Pos Gerbang & Keamanan Fisik</span>
        </div>
        <h1 class="text-2xl font-bold text-slate-900 tracking-tight mt-1">
          Operasional Pos Gerbang & Pemeriksaan SIMAK
        </h1>
        <p class="text-sm text-slate-600 mt-0.5">
          Verifikasi identitas pekerja, daftar peralatan bawaan fisik, dan catat pergerakan masuk/keluar area pabrik PT INKA (Persero).
        </p>
      </div>

      <div class="flex items-center gap-2">
        <Button variant="secondary" @click="fetchMetrics">
          <svg class="w-4 h-4 mr-1.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
          Segarkan Data
        </Button>
        <router-link to="/gate/history">
          <Button variant="secondary">
            <svg class="w-4 h-4 mr-1.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            Log Riwayat
          </Button>
        </router-link>
      </div>
    </div>

    <!-- Quick Stat Cards -->
    <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
      <StatCard
        label="Pemeriksaan Hari Ini"
        :value="metrics.inspections_today"
        description="Total pergerakan masuk & keluar dicatat hari ini"
      />
      <StatCard
        label="Dokumen SIMAK Aktif"
        :value="metrics.active_simaks"
        description="Izin aktif siap diverifikasi di pos jaga"
      />
      <div class="bg-white border rounded-xl p-5 shadow-sm relative overflow-hidden" :class="metrics.mismatch_count > 0 ? 'border-rose-200 bg-rose-50/30' : 'border-slate-200'">
        <div class="flex items-center justify-between">
          <span class="text-sm font-medium" :class="metrics.mismatch_count > 0 ? 'text-rose-800' : 'text-slate-600'">
            Temuan Selisih (Mismatch)
          </span>
          <div class="w-10 h-10 rounded-lg flex items-center justify-center" :class="metrics.mismatch_count > 0 ? 'bg-rose-100 text-rose-600' : 'bg-slate-100 text-slate-500'">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
            </svg>
          </div>
        </div>
        <div class="mt-3">
          <div class="text-3xl font-bold" :class="metrics.mismatch_count > 0 ? 'text-rose-700' : 'text-slate-900'">
            {{ metrics.mismatch_count }}
          </div>
          <p class="text-xs mt-1" :class="metrics.mismatch_count > 0 ? 'text-rose-600 font-medium' : 'text-slate-500'">
            Insiden fisik tidak sesuai daftar terdaftar
          </p>
        </div>
      </div>
    </div>

    <!-- Quick Action Gate Inspection Panel (Touch-friendly Hero Banner) -->
    <div class="bg-gradient-to-br from-slate-900 via-slate-800 to-inka-950 text-white rounded-2xl p-6 sm:p-8 shadow-xl border border-slate-700/50">
      <div class="max-w-3xl">
        <div class="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-inka-600/30 border border-inka-500/40 text-inka-300 text-xs font-semibold mb-3">
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v1m6 11h2m-6 0h-2v4m0-11v3m0 0h.01M12 12h4.01M16 20h4M4 12h4m12 0h.01M5 8h2a1 1 0 001-1V5a1 1 0 00-1-1H5a1 1 0 00-1 1v2a1 1 0 001 1zm12 0h2a1 1 0 001-1V5a1 1 0 00-1-1h-2a1 1 0 00-1 1v2a1 1 0 001 1zM5 20h2a1 1 0 001-1v-2a1 1 0 00-1-1H5a1 1 0 00-1 1v2a1 1 0 001 1z" />
          </svg>
          Akses Pemeriksaan Cepat Petugas
        </div>
        <h2 class="text-2xl sm:text-3xl font-bold tracking-tight !text-white">
          Mulai Verifikasi Dokumen SIMAK
        </h2>
        <p class="text-slate-300 text-sm sm:text-base mt-2">
          Pindai kode QR pada lembar surat izin fisik / smartphone pekerja, atau ketikkan nomor SIMAK untuk membuka checklist barang bawaan.
        </p>

        <!-- Actions Container -->
        <div class="mt-6 flex flex-col sm:flex-row items-stretch sm:items-center gap-3">
          <!-- Big Touch Scan QR Button -->
          <button
            type="button"
            @click="openQrScanner"
            class="flex items-center justify-center gap-3 px-6 py-4 rounded-xl bg-inka-600 hover:bg-inka-700 active:bg-inka-800 text-white font-semibold text-base shadow-lg shadow-inka-900/40 transition-all cursor-pointer min-h-[52px]"
          >
            <svg class="w-6 h-6 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v1m6 11h2m-6 0h-2v4m0-11v3m0 0h.01M12 12h4.01M16 20h4M4 12h4m12 0h.01M5 8h2a1 1 0 001-1V5a1 1 0 00-1-1H5a1 1 0 00-1 1v2a1 1 0 001 1zm12 0h2a1 1 0 001-1V5a1 1 0 00-1-1h-2a1 1 0 00-1 1v2a1 1 0 001 1zM5 20h2a1 1 0 001-1v-2a1 1 0 00-1-1H5a1 1 0 00-1 1v2a1 1 0 001 1z" />
            </svg>
            <span>Scan QR Code Kamera</span>
          </button>

          <!-- Divider -->
          <div class="hidden sm:block text-slate-500 font-medium px-1">atau</div>

          <!-- Direct Input Box -->
          <div class="flex-1 flex gap-2">
            <div class="relative flex-1">
              <input
                v-model="quickQuery"
                type="text"
                placeholder="Ketik Nomor SIMAK atau Token..."
                @keyup.enter="handleQuickSearch"
                class="w-full h-full min-h-[52px] px-4 py-3 rounded-xl bg-slate-800/80 border border-slate-600 text-white placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-inka-500 text-sm sm:text-base font-medium"
              />
              <button
                v-if="quickQuery"
                @click="quickQuery = ''"
                class="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 hover:text-white p-1"
                aria-label="Hapus"
              >
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
            <button
              type="button"
              @click="handleQuickSearch"
              :disabled="searching || !quickQuery.trim()"
              class="px-5 py-3 rounded-xl bg-slate-700 hover:bg-slate-600 disabled:opacity-50 text-white font-medium text-sm sm:text-base transition-colors shrink-0 flex items-center gap-2 min-h-[52px]"
            >
              <span v-if="searching" class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"></span>
              <span>Cari</span>
            </button>
          </div>
        </div>

        <div v-if="searchError" class="mt-3 text-xs text-rose-300 font-medium bg-rose-950/40 border border-rose-800/60 rounded-lg p-2.5 flex items-center gap-2">
          <svg class="w-4 h-4 text-rose-400 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
          </svg>
          <span>{{ searchError }}</span>
        </div>
      </div>
    </div>

    <!-- Recent Gate Movements Section -->
    <div class="bg-white border border-slate-200 rounded-xl shadow-sm overflow-hidden">
      <div class="p-5 border-b border-slate-200 flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div>
          <h2 class="text-base font-bold text-slate-900">Pergerakan Terakhir di Pos Gerbang</h2>
          <p class="text-xs text-slate-500 mt-0.5">Catatan verifikasi keluar-masuk alat kerja terkini oleh petugas pos jaga</p>
        </div>
        <router-link to="/gate/history" class="text-xs font-semibold text-inka-600 hover:text-inka-700 flex items-center gap-1">
          Lihat Semua Riwayat
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
          </svg>
        </router-link>
      </div>

      <!-- Movement List / Table -->
      <div v-if="loading" class="p-8 text-center text-slate-500">
        <div class="inline-block w-6 h-6 border-2 border-inka-600 border-t-transparent rounded-full animate-spin"></div>
        <p class="text-xs mt-2">Memuat catatan pergerakan pos gerbang...</p>
      </div>

      <div v-else-if="metrics.recent_movements.length === 0" class="p-8 text-center text-slate-500">
        <svg class="w-10 h-10 mx-auto text-slate-400 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
        </svg>
        <p class="text-sm font-medium text-slate-700">Belum ada riwayat pergerakan hari ini</p>
        <p class="text-xs text-slate-400 mt-0.5">Lakukan pemeriksaan SIMAK pertama menggunakan tombol di atas</p>
      </div>

      <div v-else class="divide-y divide-slate-100">
        <div
          v-for="movement in metrics.recent_movements"
          :key="movement.id"
          class="p-4 sm:p-5 hover:bg-slate-50 transition-colors flex flex-col sm:flex-row sm:items-center justify-between gap-4"
        >
          <div class="flex items-start gap-3.5">
            <!-- Direction Icon Indicator -->
            <div
              class="w-10 h-10 rounded-xl flex items-center justify-center shrink-0 font-bold"
              :class="movement.inspection_type === 'MASUK' ? 'bg-emerald-100 text-emerald-700' : 'bg-blue-100 text-blue-700'"
            >
              <svg v-if="movement.inspection_type === 'MASUK'" class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 16l-4-4m0 0l4-4m-4 4h14m-5 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h7a3 3 0 013 3v1" />
              </svg>
              <svg v-else class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
              </svg>
            </div>

            <div>
              <div class="flex items-center gap-2 flex-wrap">
                <span class="font-bold text-slate-900 text-sm sm:text-base">
                  {{ movement.simak?.simak_number || 'SIMAK' }}
                </span>
                <span
                  class="px-2 py-0.5 rounded text-xs font-bold uppercase tracking-wider"
                  :class="movement.inspection_type === 'MASUK' ? 'bg-emerald-50 text-emerald-700 border border-emerald-200' : 'bg-blue-50 text-blue-700 border border-blue-200'"
                >
                  {{ movement.inspection_type }}
                </span>
                <Badge :variant="movement.status === 'MATCH' ? 'success' : 'danger'" :dot="true">
                  {{ movement.status === 'MATCH' ? 'COCOK (MATCH)' : 'SELISIH (MISMATCH)' }}
                </Badge>
              </div>

              <div class="text-xs text-slate-500 mt-1 flex items-center gap-2 flex-wrap">
                <span class="inline-flex items-center gap-1 font-medium text-slate-700">
                  <svg class="w-3.5 h-3.5 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                  </svg>
                  {{ movement.simak?.worker?.full_name || 'Pekerja' }}
                </span>
                <span class="text-slate-300">•</span>
                <span class="inline-flex items-center gap-1 text-slate-600">
                  <svg class="w-3.5 h-3.5 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
                  </svg>
                  {{ movement.simak?.worker?.company?.name || 'PT INKA (Persero)' }}
                </span>
                <span class="text-slate-300">•</span>
                <span class="inline-flex items-center gap-1 text-slate-600">
                  <svg class="w-3.5 h-3.5 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
                  </svg>
                  {{ movement.gate_name }}
                </span>
              </div>

              <p v-if="movement.notes" class="text-xs text-slate-600 italic mt-1 bg-slate-50 rounded px-2 py-1 border border-slate-200 inline-block">
                "{{ movement.notes }}"
              </p>
            </div>
          </div>

          <div class="flex items-center justify-between sm:justify-end gap-3 shrink-0 pt-2 sm:pt-0 border-t sm:border-t-0 border-slate-100">
            <div class="text-right">
              <div class="text-xs font-semibold text-slate-700">
                {{ movement.total_actual_items }} dari {{ movement.total_registered_items }} Alat
              </div>
              <div class="text-[11px] text-slate-400">
                {{ formatDate(movement.inspected_at) }}
              </div>
            </div>

            <Button variant="secondary" size="sm" @click="viewInspectionDetail(movement.id)">
              Rincian
            </Button>
          </div>
        </div>
      </div>
    </div>

    <!-- QR Code Scanner Modal -->
    <Modal
      v-model="qrModalOpen"
      title="Pindai QR Code Surat Izin (SIMAK)"
      @close="closeQrScanner"
    >
      <div class="space-y-4">
        <p class="text-xs text-slate-600">
          Arahkan kamera ke kode QR pada lembar SIMAK pekerja. Pastikan pencahayaan cukup terang.
        </p>

        <!-- Camera Viewport Container -->
        <div class="relative bg-slate-900 rounded-xl overflow-hidden aspect-[4/3] flex items-center justify-center border-2 border-slate-700">
          <video
            ref="qrVideoRef"
            autoplay
            playsinline
            muted
            class="w-full h-full object-cover"
          ></video>

          <!-- Scanner Reticle Overlay -->
          <div class="absolute inset-0 pointer-events-none flex items-center justify-center">
            <div class="w-48 h-48 border-2 border-inka-500 rounded-xl relative shadow-[0_0_0_9999px_rgba(0,0,0,0.4)]">
              <div class="absolute top-0 left-0 w-4 h-4 border-t-4 border-l-4 border-inka-400"></div>
              <div class="absolute top-0 right-0 w-4 h-4 border-t-4 border-r-4 border-inka-400"></div>
              <div class="absolute bottom-0 left-0 w-4 h-4 border-b-4 border-l-4 border-inka-400"></div>
              <div class="absolute bottom-0 right-0 w-4 h-4 border-b-4 border-r-4 border-inka-400"></div>
              <div class="absolute inset-x-0 top-1/2 h-0.5 bg-inka-400/80 animate-pulse"></div>
            </div>
          </div>

          <div v-if="!scannerActive" class="absolute inset-0 bg-slate-900/90 flex flex-col items-center justify-center p-4 text-center">
            <svg class="w-10 h-10 text-slate-500 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
            </svg>
            <p class="text-sm font-semibold text-white">Kamera Belum Aktif</p>
            <p class="text-xs text-slate-400 mt-1 max-w-xs">
              Izinkan akses kamera di peramban atau masukkan token QR verifikasi secara manual di bawah.
            </p>
          </div>
        </div>

        <!-- Manual Input Alternative -->
        <div class="pt-2 border-t border-slate-200">
          <label class="block text-xs font-semibold text-slate-700 mb-1">
            Atau Tempel Token QR SIMAK Manual:
          </label>
          <div class="flex gap-2">
            <input
              v-model="manualTokenInput"
              type="text"
              placeholder="Contoh: c9f6b98e-4a6f-..."
              @keyup.enter="submitManualToken"
              class="flex-1 px-3 py-2 text-sm border border-slate-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-inka-500"
            />
            <Button
              variant="primary"
              :disabled="!manualTokenInput.trim()"
              @click="submitManualToken"
            >
              Buka Checklist
            </Button>
          </div>
        </div>
      </div>
    </Modal>

    <!-- Inspection Detail Modal -->
    <Modal
      v-model="detailModalOpen"
      title="Rincian Hasil Pemeriksaan Fisik"
      @close="detailModalOpen = false"
    >
      <div v-if="detailLoading" class="py-12 text-center text-slate-500">
        <div class="inline-block w-6 h-6 border-2 border-inka-600 border-t-transparent rounded-full animate-spin"></div>
        <p class="text-xs mt-2">Memuat rincian pemeriksaan...</p>
      </div>

      <div v-else-if="selectedInspection" class="space-y-4">
        <!-- Header info -->
        <div class="bg-slate-50 p-3.5 rounded-xl border border-slate-200 flex items-center justify-between">
          <div>
            <div class="text-xs text-slate-500">Nomor SIMAK</div>
            <div class="font-bold text-slate-900 text-sm">{{ selectedInspection.simak?.simak_number }}</div>
          </div>
          <div class="text-right">
            <Badge :variant="selectedInspection.status === 'MATCH' ? 'success' : 'danger'" :dot="true">
              {{ selectedInspection.status === 'MATCH' ? 'COCOK' : 'SELISIH' }}
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
            <span class="text-slate-500">Waktu Periksa:</span>
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
                  v-for="item in selectedInspection.items"
                  :key="item.id"
                  :class="item.status === 'MATCH' ? '' : 'bg-rose-50/50'"
                >
                  <td class="p-2.5 font-medium text-slate-900">
                    <div>{{ item.item_name }}</div>
                    <div class="text-[11px] text-slate-500 font-normal">
                      {{ item.brand }} {{ item.serial_number ? `• SN: ${item.serial_number}` : '' }}
                    </div>
                  </td>
                  <td class="p-2.5 text-center font-semibold text-slate-700">
                    {{ item.expected_qty }} {{ item.unit }}
                  </td>
                  <td class="p-2.5 text-center font-bold" :class="item.status === 'MATCH' ? 'text-slate-900' : 'text-rose-600'">
                    {{ item.actual_qty }} {{ item.unit }}
                  </td>
                  <td class="p-2.5 text-center">
                    <Badge :variant="item.status === 'MATCH' ? 'success' : 'danger'" size="sm">
                      {{ item.status }}
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
