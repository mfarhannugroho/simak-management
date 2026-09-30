<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  discrepancyService,
  type Discrepancy,
  type DiscrepancyStatus,
  type DiscrepancyType,
  type ResolutionAction,
} from '@/services/discrepancy'
import { useAuthStore } from '@/stores/auth'
import Button from '@/components/ui/Button.vue'
import Badge from '@/components/ui/Badge.vue'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const loading = ref(true)
const submitting = ref(false)
const errorMsg = ref('')

const discrepancy = ref<Discrepancy | null>(null)

// Resolution form state
const selectedAction = ref<ResolutionAction>('ALLOW_ENTRY')
const resolutionNotes = ref('')
const finalStatus = ref<'RESOLVED' | 'CLOSED'>('RESOLVED')

const availableActions: { value: ResolutionAction; label: string; desc: string }[] = [
  {
    value: 'ALLOW_ENTRY',
    label: 'Diberikan Izin Masuk (Dispensasi Khusus)',
    desc: 'Alat diverifikasi aman dan diizinkan masuk setelah konfirmasi manajer area.',
  },
  {
    value: 'CONFISCATE_HOLD_AT_GATE',
    label: 'Barang Ditahan Sementara di Pos Jaga',
    desc: 'Barang tidak diizinkan masuk dan diamankan di pos jaga sampai izin diperbarui.',
  },
  {
    value: 'RETURN_TO_VENDOR',
    label: 'Barang Dikembalikan ke Vendor',
    desc: 'Pekerja diminta membawa pulang barang berlebih ke pihak rekanan.',
  },
  {
    value: 'DENY_ENTRY',
    label: 'Akses Pekerja Ditolak',
    desc: 'Seluruh akses dibatalkan karena pelanggaran serius prosedur keamanan pabrik.',
  },
  {
    value: 'REVISE_PERMIT',
    label: 'Instruksi Revisi Dokumen SIMAK',
    desc: 'Pekerja harus mengajukan penambahan alat melalui portal SIMAK terlebih dahulu.',
  },
  {
    value: 'WARNING_ISSUED',
    label: 'Diberikan Surat Teguran Keamanan',
    desc: 'Diberikan teguran K3LH dan pembinaan kepatuhan pengamanan aset PT INKA.',
  },
]

const canResolve = computed(() => {
  return (
    authStore.hasRole('Supervisor') ||
    authStore.hasRole('Admin') ||
    authStore.hasRole('SIMAK Officer')
  )
})

async function loadDetail() {
  loading.value = true
  errorMsg.value = ''
  const id = route.params.id as string

  try {
    const res = await discrepancyService.getDiscrepancy(id)
    if (res.data) {
      discrepancy.value = res.data
      if (res.data.resolution_action) {
        selectedAction.value = res.data.resolution_action
      }
      if (res.data.resolution_notes) {
        resolutionNotes.value = res.data.resolution_notes
      }
    }
  } catch (err: any) {
    console.error('Failed to load discrepancy detail:', err)
    errorMsg.value = err.response?.data?.message || 'Data temuan ketidaksesuaian tidak ditemukan.'
  } finally {
    loading.value = false
  }
}

async function handleStartReview() {
  if (!discrepancy.value) return

  submitting.value = true
  try {
    const res = await discrepancyService.startReview(discrepancy.value.id)
    if (res.data) {
      discrepancy.value = res.data
    }
  } catch (err: any) {
    console.error('Failed to mark IN_REVIEW:', err)
    alert(err.response?.data?.message || 'Gagal mengubah status menjadi Dalam Investigasi.')
  } finally {
    submitting.value = false
  }
}

async function handleResolve() {
  if (!discrepancy.value) return

  if (!resolutionNotes.value.trim()) {
    alert('Catatan investigasi dan keputusan resolusi wajib diisi.')
    return
  }

  submitting.value = true
  try {
    const res = await discrepancyService.resolveDiscrepancy(discrepancy.value.id, {
      resolution_action: selectedAction.value,
      resolution_notes: resolutionNotes.value.trim(),
      status: finalStatus.value,
    })

    if (res.data) {
      discrepancy.value = res.data
      alert('Kasus ketidaksesuaian barang berhasil diselesaikan!')
    }
  } catch (err: any) {
    console.error('Failed to resolve discrepancy:', err)
    alert(err.response?.data?.message || 'Gagal menyimpan keputusan resolusi.')
  } finally {
    submitting.value = false
  }
}

function formatTypeLabel(type: DiscrepancyType) {
  switch (type) {
    case 'QUANTITY_MISMATCH':
      return 'Selisih Kuantitas'
    case 'UNREGISTERED_ITEM':
      return 'Barang Tidak Terdaftar'
    case 'MISSING_ITEM':
      return 'Barang Kurang / Hilang'
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

function formatActionLabel(action?: ResolutionAction) {
  if (!action) return '-'
  const match = availableActions.find((a) => a.value === action)
  return match ? match.label : action
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
  loadDetail()
})
</script>

<template>
  <div class="space-y-6 max-w-6xl mx-auto pb-16">
    <!-- Breadcrumb & Top Bar -->
    <div class="flex items-center justify-between gap-4">
      <div class="flex items-center gap-2">
        <router-link to="/discrepancies" class="text-sm font-semibold text-slate-500 hover:text-inka-600 flex items-center gap-1">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
          </svg>
          Kembali ke Log Discrepancy
        </router-link>
      </div>

      <div class="flex items-center gap-2">
        <span class="text-xs text-slate-500 font-medium">Sistem Resolusi Temuan Fisik</span>
      </div>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="bg-white border rounded-xl p-12 text-center shadow-sm">
      <div class="inline-block w-8 h-8 border-3 border-inka-600 border-t-transparent rounded-full animate-spin"></div>
      <p class="text-sm font-semibold text-slate-700 mt-3">Memuat Rincian Kasus Ketidaksesuaian...</p>
    </div>

    <!-- Error State -->
    <div v-else-if="errorMsg" class="bg-white border border-rose-200 rounded-xl p-8 text-center shadow-sm">
      <div class="w-14 h-14 bg-rose-100 text-rose-600 rounded-full flex items-center justify-center mx-auto mb-4">
        <svg class="w-7 h-7 text-rose-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
        </svg>
      </div>
      <h2 class="text-xl font-bold text-slate-900">Gagal Memuat Kasus</h2>
      <p class="text-sm text-rose-600 font-medium mt-1">{{ errorMsg }}</p>
      <div class="mt-6">
        <router-link to="/discrepancies">
          <Button variant="secondary">Kembali ke Daftar</Button>
        </router-link>
      </div>
    </div>

    <!-- Main Detail Content -->
    <div v-else-if="discrepancy" class="space-y-6">
      <!-- 1. Header Banner -->
      <div class="bg-white border border-slate-200 rounded-2xl shadow-sm overflow-hidden">
        <div class="bg-gradient-to-r from-slate-900 to-slate-800 text-white p-5 sm:p-6 flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div>
            <div class="flex items-center gap-2 flex-wrap">
              <span class="text-xs uppercase tracking-wider font-bold text-rose-400">
                TEMUAN KETIDAKSESUAIAN POS GERBANG
              </span>
              <span class="text-xs px-2 py-0.5 rounded bg-slate-700 text-slate-300 font-mono">
                {{ formatTypeLabel(discrepancy.discrepancy_type) }}
              </span>
              <Badge :variant="getStatusVariant(discrepancy.status)" :dot="true" size="sm">
                {{ discrepancy.status }}
              </Badge>
            </div>
            <h1 class="text-xl sm:text-2xl font-bold tracking-tight text-white mt-1 font-mono">
              {{ discrepancy.discrepancy_number }}
            </h1>
          </div>

          <div class="text-left sm:text-right">
            <div class="text-xs text-slate-400">Waktu Dilaporkan di Pos:</div>
            <div class="text-sm font-semibold text-slate-200">
              {{ formatDate(discrepancy.created_at) }}
            </div>
          </div>
        </div>
      </div>

      <!-- 2. Two-Column Grid: Investigation Info (Left) vs Resolution Action (Right) -->
      <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <!-- Left Column: Case Information & Comparison (span 2) -->
        <div class="lg:col-span-2 space-y-6">
          <!-- Card: Item Discrepancy Breakdown -->
          <div class="bg-white border border-slate-200 rounded-2xl p-5 sm:p-6 shadow-sm space-y-4">
            <h2 class="text-base font-bold text-slate-900 flex items-center gap-2">
              <svg class="w-4 h-4 text-inka-700" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
              </svg>
              <span>Rincian Selisih Fisik Peralatan</span>
            </h2>

            <div class="bg-slate-50 rounded-xl p-4 border border-slate-200">
              <div class="grid grid-cols-1 sm:grid-cols-2 gap-4 text-xs">
                <div>
                  <div class="text-slate-500 font-medium">Nama Peralatan Terkait:</div>
                  <div class="font-bold text-slate-900 text-sm mt-0.5">{{ discrepancy.item_name }}</div>
                  <div class="text-slate-500 mt-1">
                    Merk: <strong>{{ discrepancy.brand || '-' }}</strong> • No. Seri: <strong>{{ discrepancy.serial_number || '-' }}</strong>
                  </div>
                </div>

                <div>
                  <div class="text-slate-500 font-medium">Komparasi Jumlah Kuantitas:</div>
                  <div class="mt-1 flex items-center gap-4">
                    <div class="bg-white px-3 py-2 rounded-lg border border-slate-200 text-center">
                      <div class="text-[10px] uppercase font-semibold text-slate-500">Izin Terdaftar</div>
                      <div class="text-base font-bold text-slate-700">
                        {{ discrepancy.expected_qty }} <span class="text-xs font-normal">{{ discrepancy.unit }}</span>
                      </div>
                    </div>

                    <div class="text-slate-400 font-bold">vs</div>

                    <div class="bg-rose-50 px-3 py-2 rounded-lg border border-rose-200 text-center">
                      <div class="text-[10px] uppercase font-semibold text-rose-700">Fisik Aktual</div>
                      <div class="text-base font-bold text-rose-700">
                        {{ discrepancy.actual_qty }} <span class="text-xs font-normal">{{ discrepancy.unit }}</span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              <!-- Officer Description -->
              <div class="mt-4 pt-3 border-t border-slate-200 text-xs">
                <div class="font-semibold text-slate-700 mb-1">Keterangan Petugas Pos Jaga:</div>
                <p class="text-slate-600 bg-white p-3 rounded-lg border border-slate-200 italic">
                  "{{ discrepancy.description || 'Tidak ada catatan tambahan saat pemeriksaan fisik.' }}"
                </p>
              </div>
            </div>

            <!-- Reported By & Gate Info -->
            <div class="grid grid-cols-2 gap-3 text-xs text-slate-600 pt-1">
              <div>
                <span class="text-slate-400">Petugas Pelapor:</span>
                <span class="font-semibold text-slate-800 ml-1">
                  {{ discrepancy.reporter?.name || discrepancy.reporter?.email || 'Gate Officer' }}
                </span>
              </div>
              <div>
                <span class="text-slate-400">Lokasi Pos Gerbang:</span>
                <span class="font-semibold text-slate-800 ml-1">
                  {{ discrepancy.inspection?.gate_name || 'Pos Gerbang INKA' }}
                </span>
              </div>
            </div>
          </div>

          <!-- Card: Related SIMAK & Worker Profile -->
          <div class="bg-white border border-slate-200 rounded-2xl p-5 sm:p-6 shadow-sm space-y-4">
            <h2 class="text-base font-bold text-slate-900 flex items-center gap-2">
              <svg class="w-4 h-4 text-inka-700" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
              </svg>
              <span>Informasi Dokumen SIMAK Terkait</span>
            </h2>

            <div class="grid grid-cols-1 sm:grid-cols-2 gap-4 text-xs">
              <div>
                <div class="text-slate-500 font-medium">Nomor Surat Izin:</div>
                <div class="font-bold text-slate-900 text-sm mt-0.5 font-mono">
                  {{ discrepancy.simak?.simak_number }}
                </div>
                <div class="text-slate-500 mt-1">
                  Arah Izin: <strong>{{ discrepancy.simak?.direction }}</strong>
                </div>
              </div>

              <div>
                <div class="text-slate-500 font-medium">Pekerja / Pembawa Alat:</div>
                <div class="font-bold text-slate-900 text-sm mt-0.5">
                  {{ discrepancy.simak?.worker?.full_name }}
                </div>
                <div class="text-slate-500 mt-1">
                  Perusahaan: <strong>{{ discrepancy.simak?.worker?.company?.name || 'PT INKA (Persero)' }}</strong>
                </div>
              </div>

              <div>
                <div class="text-slate-500 font-medium">Area Pabrik Tujuan:</div>
                <div class="font-bold text-slate-800 mt-0.5 flex items-center gap-1.5">
                  <svg class="w-3.5 h-3.5 text-slate-400 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
                  </svg>
                  <span>{{ discrepancy.simak?.destination_area }}</span>
                </div>
              </div>

              <div>
                <div class="text-slate-500 font-medium">Keperluan Pekerjaan:</div>
                <div class="text-slate-700 mt-0.5">
                  "{{ discrepancy.simak?.purpose }}"
                </div>
              </div>
            </div>

            <div class="pt-3 border-t border-slate-100 flex items-center justify-between">
              <span class="text-xs text-slate-500">Lihat rincian lengkap dokumen izin:</span>
              <router-link :to="`/simak/${discrepancy.simak_id}`">
                <Button variant="secondary" size="sm">
                  Buka Detail SIMAK
                </Button>
              </router-link>
            </div>
          </div>

          <!-- Card: Audit Log / Timeline -->
          <div class="bg-white border border-slate-200 rounded-2xl p-5 sm:p-6 shadow-sm space-y-4">
            <h2 class="text-base font-bold text-slate-900 flex items-center gap-2">
              <span>⏱️</span> Kronologi Penanganan Kasus
            </h2>

            <div class="relative pl-6 space-y-6 border-l-2 border-slate-200 ml-2 text-xs">
              <!-- Step 1 -->
              <div class="relative">
                <div class="absolute -left-[31px] top-0.5 w-4 h-4 rounded-full bg-rose-500 border-2 border-white"></div>
                <div class="font-bold text-slate-900">Temuan Tercatat di Pos Gerbang</div>
                <div class="text-slate-500 mt-0.5">{{ formatDate(discrepancy.created_at) }}</div>
                <p class="text-slate-600 mt-1">
                  Ketidaksesuaian fisik dideteksi saat verifikasi gerbang oleh petugas keamanan.
                </p>
              </div>

              <!-- Step 2 -->
              <div v-if="discrepancy.status !== 'OPEN'" class="relative">
                <div class="absolute -left-[31px] top-0.5 w-4 h-4 rounded-full bg-amber-500 border-2 border-white"></div>
                <div class="font-bold text-slate-900">Pemeriksaan / Investigasi Pengawas (IN_REVIEW)</div>
                <div class="text-slate-500 mt-0.5">{{ formatDate(discrepancy.reviewed_at) }}</div>
                <p class="text-slate-600 mt-1">
                  Kasus sedang ditindaklanjuti dan dikonfirmasi dengan manajer area / unit terkait.
                </p>
              </div>

              <!-- Step 3 -->
              <div v-if="discrepancy.status === 'RESOLVED' || discrepancy.status === 'CLOSED'" class="relative">
                <div class="absolute -left-[31px] top-0.5 w-4 h-4 rounded-full bg-emerald-500 border-2 border-white"></div>
                <div class="font-bold text-slate-900">Kasus Telah Diselesaikan ({{ discrepancy.status }})</div>
                <div class="text-slate-500 mt-0.5">{{ formatDate(discrepancy.reviewed_at) }}</div>
                <p class="text-slate-600 mt-1">
                  Tindakan resolusi: <strong>{{ formatActionLabel(discrepancy.resolution_action) }}</strong>
                </p>
              </div>
            </div>
          </div>
        </div>

        <!-- Right Column: Resolution & Decision Panel (span 1) -->
        <div class="space-y-6">
          <!-- Card: Decision Panel -->
          <div class="bg-white border border-slate-200 rounded-2xl p-5 sm:p-6 shadow-sm space-y-4">
            <h2 class="text-base font-bold text-slate-900 flex items-center gap-2">
              <svg class="w-4 h-4 text-inka-700" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 6l3 1m0 0l-3 9a5.002 5.002 0 006.001 0M6 7l3 9M6 7l6-2m6 2l3-1m-3 1l-3 9a5.002 5.002 0 006.001 0M18 7l3 9m-3-9l-6-2m0-2v2m0 16V5m0 16H9m3 0h3" />
              </svg>
              <span>Putusan & Resolusi Kasus</span>
            </h2>

            <!-- If case already RESOLVED or CLOSED -->
            <div v-if="discrepancy.status === 'RESOLVED' || discrepancy.status === 'CLOSED'" class="space-y-3">
              <div class="p-4 rounded-xl bg-emerald-50 border border-emerald-200 text-xs space-y-2">
                <div class="flex items-center justify-between">
                  <span class="font-bold text-emerald-800">KASUS TELAH DISELESAIKAN</span>
                  <Badge variant="success" size="sm">{{ discrepancy.status }}</Badge>
                </div>

                <div>
                  <div class="text-slate-500">Tindakan yang Diambil:</div>
                  <div class="font-bold text-slate-900 mt-0.5">
                    {{ formatActionLabel(discrepancy.resolution_action) }}
                  </div>
                </div>

                <div class="pt-2 border-t border-emerald-200/60">
                  <div class="text-slate-500">Catatan & Keputusan Supervisor:</div>
                  <div class="text-slate-800 italic mt-0.5">
                    "{{ discrepancy.resolution_notes }}"
                  </div>
                </div>

                <div class="pt-2 border-t border-emerald-200/60 text-[11px] text-slate-500 flex justify-between">
                  <span>Diselesaikan pada:</span>
                  <span class="font-semibold">{{ formatDate(discrepancy.reviewed_at) }}</span>
                </div>
              </div>
            </div>

            <!-- If case is OPEN or IN_REVIEW and user can resolve -->
            <div v-else-if="canResolve" class="space-y-4">
              <!-- Start review button if OPEN -->
              <div v-if="discrepancy.status === 'OPEN'">
                <Button
                  variant="secondary"
                  class="w-full text-xs font-bold py-2.5 flex items-center justify-center gap-1.5"
                  :disabled="submitting"
                  @click="handleStartReview"
                >
                  <span v-if="submitting" class="w-3.5 h-3.5 border-2 border-slate-600 border-t-transparent rounded-full animate-spin mr-1.5"></span>
                  <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                  </svg>
                  <span>Tandai Sedang Diinvestigasi (IN_REVIEW)</span>
                </Button>
                <p class="text-[11px] text-slate-500 mt-1 text-center">
                  Memberi tahu tim bahwa kasus sedang diverifikasi ke lapangan.
                </p>
              </div>

              <!-- Form Input Resolusi -->
              <div class="pt-3 border-t border-slate-200 space-y-4 text-xs">
                <!-- Action Dropdown -->
                <div>
                  <label class="block font-bold text-slate-700 uppercase tracking-wider mb-1.5">
                    Tindakan Resolusi yang Diambil:
                  </label>
                  <select
                    v-model="selectedAction"
                    class="w-full px-3 py-2.5 border border-slate-300 rounded-xl bg-white text-slate-900 font-semibold focus:ring-2 focus:ring-inka-500 focus:outline-none"
                  >
                    <option v-for="act in availableActions" :key="act.value" :value="act.value">
                      {{ act.label }}
                    </option>
                  </select>
                  <p class="text-[11px] text-slate-500 mt-1 italic">
                    {{ availableActions.find(a => a.value === selectedAction)?.desc }}
                  </p>
                </div>

                <!-- Notes Textarea -->
                <div>
                  <label class="block font-bold text-slate-700 uppercase tracking-wider mb-1.5">
                    Catatan Investigasi & Alasan Putusan (Wajib):
                  </label>
                  <textarea
                    v-model="resolutionNotes"
                    rows="4"
                    placeholder="Tuliskan hasil investigasi, pihak yang telah dikonfirmasi, dan dasar keputusan..."
                    class="w-full px-3 py-2 border border-slate-300 rounded-xl text-xs focus:ring-2 focus:ring-inka-500 focus:outline-none"
                  ></textarea>
                </div>

                <!-- Final Status Picker -->
                <div>
                  <label class="block font-bold text-slate-700 uppercase tracking-wider mb-1.5">
                    Status Akhir Kasus:
                  </label>
                  <div class="grid grid-cols-2 gap-2">
                    <button
                      type="button"
                      @click="finalStatus = 'RESOLVED'"
                      class="py-2 px-3 rounded-lg border text-xs font-bold transition-colors cursor-pointer"
                      :class="finalStatus === 'RESOLVED' ? 'bg-emerald-600 text-white border-emerald-600' : 'bg-slate-50 text-slate-700 border-slate-200'"
                    >
                      RESOLVED
                    </button>
                    <button
                      type="button"
                      @click="finalStatus = 'CLOSED'"
                      class="py-2 px-3 rounded-lg border text-xs font-bold transition-colors cursor-pointer"
                      :class="finalStatus === 'CLOSED' ? 'bg-slate-800 text-white border-slate-800' : 'bg-slate-50 text-slate-700 border-slate-200'"
                    >
                      CLOSED
                    </button>
                  </div>
                </div>

                <!-- Submit Button -->
                <Button
                  variant="primary"
                  class="w-full text-sm font-bold py-3 mt-2"
                  :disabled="submitting || !resolutionNotes.trim()"
                  @click="handleResolve"
                >
                  <span v-if="submitting" class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin mr-2"></span>
                  <span>Simpan Putusan & Selesaikan</span>
                </Button>
              </div>
            </div>

            <!-- If user is not authorized to resolve -->
            <div v-else class="p-4 rounded-xl bg-slate-50 border border-slate-200 text-xs text-slate-600 text-center">
              <p class="font-semibold text-slate-700">Akses Terbatas</p>
              <p class="mt-1 text-slate-500">
                Penyelesaian kasus temuan ketidaksesuaian hanya dapat dilakukan oleh Supervisor atau Admin yang berwenang.
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
