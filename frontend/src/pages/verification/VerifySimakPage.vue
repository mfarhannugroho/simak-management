<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { documentService, type PublicVerificationResponse } from '@/services/document'

const route = useRoute()
const router = useRouter()
const token = computed(() => route.params.token as string)

const loading = ref(true)
const result = ref<PublicVerificationResponse | null>(null)
const errorMsg = ref('')

async function verifyToken() {
  loading.value = true
  errorMsg.value = ''
  try {
    const res = await documentService.verifyQRToken(token.value)
    if (res.data) {
      result.value = res.data
    }
  } catch (err: any) {
    errorMsg.value = err?.response?.data?.message || 'Gagal memverifikasi token QR SIMAK.'
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  if (token.value) {
    verifyToken()
  } else {
    loading.value = false
    errorMsg.value = 'Token verifikasi tidak ditemukan pada URL.'
  }
})
</script>

<template>
  <div class="min-h-screen bg-slate-100 flex flex-col justify-between py-6 px-4 sm:px-6">
    <!-- Top Branding -->
    <div class="max-w-md w-full mx-auto space-y-4">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-2">
          <div class="w-8 h-8 rounded bg-red-600 text-white font-extrabold text-sm flex items-center justify-center">
            INKA
          </div>
          <div>
            <h1 class="text-xs font-bold uppercase tracking-wider text-slate-800">PT INKA (PERSERO)</h1>
            <p class="text-[10px] text-slate-500">Verifikasi Dokumen Keaslian SIMAK</p>
          </div>
        </div>

        <button
          type="button"
          class="text-xs text-slate-500 hover:text-slate-800 transition-colors"
          @click="router.push('/login')"
        >
          Masuk Portal →
        </button>
      </div>

      <!-- Loading State -->
      <div v-if="loading" class="bg-white rounded-2xl p-8 border border-slate-200 text-center shadow-sm space-y-3">
        <div class="w-10 h-10 border-4 border-red-600 border-t-transparent rounded-full animate-spin mx-auto"></div>
        <p class="text-xs font-medium text-slate-600">Memeriksa keaslian kode QR SIMAK...</p>
      </div>

      <!-- Error State -->
      <div v-else-if="errorMsg" class="bg-white rounded-2xl p-6 border border-red-200 text-center shadow-sm space-y-3">
        <div class="w-12 h-12 rounded-full bg-red-100 text-red-600 flex items-center justify-center mx-auto">
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
        </div>
        <h3 class="text-sm font-bold text-slate-900">Gagal Memverifikasi</h3>
        <p class="text-xs text-slate-500 leading-relaxed">{{ errorMsg }}</p>
      </div>

      <!-- Result Card -->
      <div v-else-if="result" class="space-y-4">
        <!-- Status Banner -->
        <div
          v-if="result.valid"
          class="bg-emerald-600 text-white rounded-2xl p-6 shadow-md text-center space-y-2 relative overflow-hidden"
        >
          <div class="w-14 h-14 bg-white/20 backdrop-blur-xs rounded-full flex items-center justify-center mx-auto">
            <svg class="w-8 h-8 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 13l4 4L19 7" />
            </svg>
          </div>
          <h2 class="text-xl font-extrabold tracking-wide uppercase">SIMAK VALID & RESMI</h2>
          <p class="text-xs text-emerald-100 max-w-xs mx-auto leading-relaxed">
            Dokumen terverifikasi aktif pada sistem resmi PT Industri Kereta Api (Persero).
          </p>
        </div>

        <!-- Invalid / Expired / Cancelled Banner -->
        <div
          v-else
          class="bg-rose-600 text-white rounded-2xl p-6 shadow-md text-center space-y-2 relative overflow-hidden"
        >
          <div class="w-14 h-14 bg-white/20 backdrop-blur-xs rounded-full flex items-center justify-center mx-auto">
            <svg class="w-8 h-8 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </div>
          <h2 class="text-xl font-extrabold tracking-wide uppercase">
            {{ result.status === 'EXPIRED' ? 'SIMAK KEDALUWARSA' : result.status === 'CANCELLED' ? 'SIMAK DIBATALKAN' : 'SIMAK TIDAK VALID' }}
          </h2>
          <p class="text-xs text-rose-100 max-w-xs mx-auto leading-relaxed">
            {{ result.reason || 'Izin ini tidak dapat digunakan untuk verifikasi di pos pengamanan.' }}
          </p>
        </div>

        <!-- Details Sheet -->
        <div class="bg-white rounded-2xl border border-slate-200 shadow-sm p-5 space-y-4">
          <!-- Reference & Validity -->
          <div class="flex items-center justify-between pb-3 border-b border-slate-100">
            <div>
              <p class="text-[10px] text-slate-400 uppercase tracking-wider font-semibold">Nomor SIMAK</p>
              <p class="text-xs font-mono font-bold text-slate-900">{{ result.simak_number || '-' }}</p>
            </div>
            <div class="text-right">
              <p class="text-[10px] text-slate-400 uppercase tracking-wider font-semibold">Masa Berlaku</p>
              <p class="text-xs font-bold text-slate-900">{{ result.start_date }} - {{ result.end_date }}</p>
            </div>
          </div>

          <!-- Personil & Unit -->
          <div class="space-y-2.5 text-xs">
            <div>
              <p class="text-[11px] text-slate-400">Pemohon / Teknisi</p>
              <p class="font-bold text-slate-900">{{ result.applicant_name }}</p>
              <p class="text-[11px] text-slate-500 font-mono">{{ result.worker_id }}</p>
            </div>

            <div class="grid grid-cols-2 gap-2">
              <div>
                <p class="text-[11px] text-slate-400">Perusahaan / Vendor</p>
                <p class="font-semibold text-slate-800">{{ result.company }}</p>
              </div>
              <div v-if="result.department">
                <p class="text-[11px] text-slate-400">Departemen / Divisi</p>
                <p class="font-semibold text-slate-800">{{ result.department }}</p>
              </div>
            </div>

            <div class="grid grid-cols-2 gap-2 pt-1">
              <div>
                <p class="text-[11px] text-slate-400">Area Workshop Pabrik</p>
                <p class="font-semibold text-slate-800">{{ result.destination_area }}</p>
              </div>
              <div>
                <p class="text-[11px] text-slate-400">Arah Aliran</p>
                <span class="inline-flex items-center px-2 py-0.5 rounded text-[10px] font-bold bg-blue-50 text-blue-700 border border-blue-200 uppercase">
                  {{ result.direction }}
                </span>
              </div>
            </div>

            <div v-if="result.purpose" class="pt-1">
              <p class="text-[11px] text-slate-400">Keperluan Kerja</p>
              <p class="text-slate-700 bg-slate-50 p-2.5 rounded-lg border border-slate-100 text-[11px] leading-relaxed">
                {{ result.purpose }}
              </p>
            </div>
          </div>
        </div>

        <!-- Equipment List Card -->
        <div v-if="result.items && result.items.length > 0" class="bg-white rounded-2xl border border-slate-200 shadow-sm p-5 space-y-3">
          <div class="flex items-center justify-between pb-2 border-b border-slate-100">
            <h3 class="text-xs font-bold uppercase tracking-wider text-slate-900">
              Peralatan Kerja Terdaftar ({{ result.total_items }} Item)
            </h3>
            <span class="text-[10px] font-semibold text-emerald-700 bg-emerald-50 px-2 py-0.5 rounded-full border border-emerald-200">
              Checklist Gate
            </span>
          </div>

          <div class="divide-y divide-slate-100">
            <div
              v-for="(item, idx) in result.items"
              :key="idx"
              class="py-2.5 first:pt-0 last:pb-0 flex items-start justify-between gap-3 text-xs"
            >
              <div class="space-y-0.5 flex-1">
                <div class="flex items-center gap-1.5">
                  <span class="w-4 h-4 rounded-full bg-slate-100 text-slate-600 font-bold text-[10px] flex items-center justify-center flex-shrink-0">
                    {{ idx + 1 }}
                  </span>
                  <p class="font-bold text-slate-900">{{ item.item_name }}</p>
                </div>
                <p v-if="item.brand" class="text-[11px] text-slate-500 pl-5.5">
                  {{ item.brand }} {{ item.model || '' }}
                </p>
                <p v-if="item.serial_number" class="text-[10px] font-mono text-slate-600 pl-5.5">
                  S/N: {{ item.serial_number }}
                </p>
              </div>

              <div class="text-right flex-shrink-0">
                <span class="inline-flex items-center px-2 py-0.5 rounded bg-slate-100 text-slate-800 font-bold text-xs">
                  {{ item.quantity }} {{ item.unit }}
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Footer -->
    <div class="text-center text-[10px] text-slate-400 mt-6 font-mono">
      Sistem SIMAK PT INKA (Persero) • Pindai Terenkripsi Kriptografis
    </div>
  </div>
</template>
