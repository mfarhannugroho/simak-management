<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { documentService, type SimakDocument } from '@/services/document'
import Button from '@/components/ui/Button.vue'

const route = useRoute()
const router = useRouter()
const simakId = computed(() => route.params.id as string)

const loading = ref(true)
const doc = ref<SimakDocument | null>(null)
const errorMsg = ref('')

async function fetchDocument() {
  loading.value = true
  errorMsg.value = ''
  try {
    const res = await documentService.getDocument(simakId.value)
    if (res.data) {
      doc.value = res.data
    }
  } catch (err: any) {
    errorMsg.value = err?.response?.data?.message || 'Dokumen resmi belum diterbitkan untuk SIMAK ini.'
  } finally {
    loading.value = false
  }
}

function handlePrint() {
  window.print()
}

function formatDate(dateStr?: string) {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'long',
    year: 'numeric',
  })
}

function copyLink() {
  if (!doc.value?.simak?.qr_token) return
  const url = `${window.location.origin}/verify/${doc.value.simak.qr_token}`
  navigator.clipboard.writeText(url)
  alert('Tautan verifikasi keaslian berhasil disalin:\n' + url)
}

onMounted(() => {
  fetchDocument()
})
</script>

<template>
  <div class="space-y-6">
    <!-- Non-Printable Header & Actions Toolbar -->
    <div class="no-print flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 bg-white p-4 rounded-xl border border-slate-200 shadow-sm">
      <div class="flex items-center gap-3">
        <button
          type="button"
          class="p-2 rounded-lg border border-slate-200 bg-white text-slate-600 hover:bg-slate-50 transition-colors"
          @click="router.back()"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
          </svg>
        </button>
        <div>
          <h1 class="text-lg font-bold text-slate-900">Pratinjau Dokumen Resmi SIMAK</h1>
          <p class="text-xs text-slate-500">Standar Dokumen A4 PT INKA (Persero) dengan Pengaman QR Code</p>
        </div>
      </div>

      <div class="flex items-center gap-2">
        <Button variant="outline" size="sm" class="gap-1.5 text-xs text-slate-700" @click="copyLink">
          <svg class="w-4 h-4 text-slate-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
          </svg>
          <span>Salin Tautan</span>
        </Button>

        <Button variant="primary" size="sm" class="gap-1.5 shadow-sm" @click="handlePrint">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 17h2a2 2 0 002-2v-4a2 2 0 00-2-2H5a2 2 0 00-2 2v4a2 2 0 002 2h2m2 4h6a2 2 0 002-2v-4a2 2 0 00-2-2H9a2 2 0 00-2 2v4a2 2 0 002 2zm8-12V5a2 2 0 00-2-2H9a2 2 0 00-2 2v4h10z" />
          </svg>
          <span>Cetak Dokumen / Simpan PDF</span>
        </Button>
      </div>
    </div>

    <!-- Error State -->
    <div v-if="errorMsg" class="no-print p-6 rounded-xl bg-red-50 border border-red-200 text-center space-y-2">
      <p class="text-sm font-semibold text-red-800">{{ errorMsg }}</p>
      <p class="text-xs text-red-600">
        Pastikan SIMAK telah disetujui (APPROVED) dan diterbitkan secara resmi oleh SIMAK Officer.
      </p>
      <div class="pt-2">
        <Button variant="outline" size="sm" @click="router.push('/simak')">Kembali ke Daftar SIMAK</Button>
      </div>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="no-print py-16 text-center text-sm text-slate-400">
      Menyiapkan lembar dokumen resmi...
    </div>

    <!-- A4 Printable Document Container -->
    <div v-else-if="doc" class="flex justify-center pb-12">
      <div class="print-sheet bg-white text-slate-900 shadow-lg border border-slate-200 rounded-none w-full max-w-[210mm] min-h-[297mm] p-10 sm:p-14 relative flex flex-col justify-between">
        <!-- Top Letterhead / Kop Surat PT INKA -->
        <div class="space-y-3">
          <div class="flex items-center justify-between border-b-2 border-slate-900 pb-3">
            <div class="flex items-center gap-4">
              <!-- Official INKA Logo -->
              <img src="/logo.png" alt="PT INKA (Persero)" class="h-11 w-auto object-contain shrink-0" />
              <div>
                <h2 class="text-base font-extrabold uppercase tracking-wide text-slate-900">
                  PT INDUSTRI KERETA API (PERSERO)
                </h2>
                <p class="text-[10px] text-slate-600 font-medium">
                  KANTOR PUSAT & PABRIK: JL. YOS SUDARSO NO. 71 MADIUN 63122, JAWA TIMUR, INDONESIA
                </p>
                <p class="text-[9px] text-slate-500">
                  TELEPON: (0351) 451051 • EMAIL: sekretariat@inka.co.id • SITUS: www.inka.co.id
                </p>
              </div>
            </div>

            <!-- Security Badge -->
            <div class="text-right">
              <span class="inline-block px-2 py-0.5 border border-slate-800 text-[10px] font-bold uppercase tracking-wider">
                DOKUMEN RESMI
              </span>
            </div>
          </div>

          <!-- Document Title & Reference Number -->
          <div class="text-center pt-2 pb-2">
            <h3 class="text-base font-bold uppercase tracking-wider text-slate-950 underline decoration-2 underline-offset-4">
              SURAT IJIN MEMBAWA ALAT KERJA (SIMAK)
            </h3>
            <div class="flex items-center justify-center gap-4 text-xs font-mono font-bold mt-1 text-slate-800">
              <span>NOMOR: {{ doc.document_number }}</span>
              <span>•</span>
              <span class="font-sans font-normal text-slate-600">TANGGAL TERBIT: {{ formatDate(doc.issued_at) }}</span>
            </div>
          </div>

          <!-- Permission Statement -->
          <p class="text-xs text-slate-800 leading-relaxed text-justify">
            Berdasarkan permohonan yang telah diajukan dan diverifikasi sesuai Standar Operasional Prosedur Pengamanan Pabrik, dengan ini diberikan izin resmi kepada personil kerja berikut untuk membawa peralatan dan perlengkapan kerja ke dalam / ke luar lingkungan kerja PT Industri Kereta Api (Persero):
          </p>

          <!-- Applicant Information Grid -->
          <div class="border border-slate-300 rounded p-3 text-xs bg-slate-50/50 grid grid-cols-2 gap-y-1.5 gap-x-4">
            <div>
              <span class="text-slate-500 inline-block w-28">Nama Personil</span>:
              <strong class="text-slate-900 font-semibold">{{ doc.simak?.worker?.full_name }}</strong>
            </div>
            <div>
              <span class="text-slate-500 inline-block w-28">ID Pekerja / Badge</span>:
              <span class="font-mono">{{ doc.simak?.worker?.worker_id || '-' }}</span>
            </div>
            <div>
              <span class="text-slate-500 inline-block w-28">Perusahaan / Vendor</span>:
              <span>{{ doc.simak?.worker?.company?.name || 'PT INKA (Persero)' }}</span>
            </div>
            <div>
              <span class="text-slate-500 inline-block w-28">Departemen / Unit</span>:
              <span>{{ doc.simak?.worker?.department?.name || '-' }}</span>
            </div>
            <div>
              <span class="text-slate-500 inline-block w-28">Arah Pergerakan</span>:
              <strong class="uppercase text-slate-900">{{ doc.simak?.direction }}</strong>
            </div>
            <div>
              <span class="text-slate-500 inline-block w-28">Masa Berlaku</span>:
              <strong class="text-slate-900">{{ formatDate(doc.simak?.start_date) }} s/d {{ formatDate(doc.simak?.end_date) }}</strong>
            </div>
            <div class="col-span-2">
              <span class="text-slate-500 inline-block w-28">Area / Workshop</span>:
              <span>{{ doc.simak?.destination_area }}</span>
            </div>
            <div class="col-span-2">
              <span class="text-slate-500 inline-block w-28">Tujuan Pekerjaan</span>:
              <span>{{ doc.simak?.purpose }}</span>
            </div>
          </div>

          <!-- Equipment Item Table -->
          <div class="space-y-1 pt-1">
            <p class="text-xs font-bold uppercase tracking-wider text-slate-900">
              Daftar Rincian Peralatan / Barang Bawaan Kerja
            </p>
            <table class="w-full border-collapse border border-slate-300 text-xs">
              <thead>
                <tr class="bg-slate-100 text-slate-800 text-center font-bold">
                  <th class="border border-slate-300 py-1.5 px-2 w-8">No</th>
                  <th class="border border-slate-300 py-1.5 px-3 text-left">Nama Barang / Peralatan</th>
                  <th class="border border-slate-300 py-1.5 px-2 text-left">Merek / Tipe</th>
                  <th class="border border-slate-300 py-1.5 px-2 text-left">Nomor Seri (Serial No)</th>
                  <th class="border border-slate-300 py-1.5 px-2 w-16 text-center">Jumlah</th>
                  <th class="border border-slate-300 py-1.5 px-3 text-left">Keterangan / Kondisi</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(it, i) in doc.simak?.items" :key="it.id || i" class="text-slate-800">
                  <td class="border border-slate-300 py-1 px-2 text-center">{{ i + 1 }}</td>
                  <td class="border border-slate-300 py-1 px-3 font-semibold">{{ it.item_name }}</td>
                  <td class="border border-slate-300 py-1 px-2">{{ it.brand || '-' }} {{ it.model || '' }}</td>
                  <td class="border border-slate-300 py-1 px-2 font-mono text-[11px]">{{ it.serial_number || '-' }}</td>
                  <td class="border border-slate-300 py-1 px-2 text-center font-bold">{{ it.quantity }} {{ it.unit }}</td>
                  <td class="border border-slate-300 py-1 px-3 text-[11px] text-slate-600">{{ it.size_spec || it.description || '-' }}</td>
                </tr>
              </tbody>
            </table>
          </div>

          <!-- Terms and Conditions Notes -->
          <div class="text-[10px] text-slate-600 space-y-0.5 pt-1">
            <p class="font-bold uppercase text-slate-800 text-[10px]">Ketentuan Penggunaan Dokumen:</p>
            <ol class="list-decimal pl-4 space-y-0.5">
              <li>Surat ijin ini wajib dibawa dan ditunjukkan kepada Petugas Keamanan di Pos Gerbang setiap keluar / masuk pabrik.</li>
              <li>Peralatan yang dibawa harus cocok dan sesuai dengan jumlah fisik pada tabel daftar barang. Segala selisih akan dicatat sebagai temuan (discrepancy).</li>
              <li>Peralatan yang tercatat dalam dokumen ini BUKAN merupakan aset inventaris PT INKA (Persero).</li>
            </ol>
          </div>
        </div>

        <!-- Bottom Signatures & QR Section -->
        <div class="pt-6 border-t border-slate-300 mt-6">
          <div class="grid grid-cols-4 gap-2 text-center text-xs">
            <!-- Signature 1: Applicant -->
            <div class="flex flex-col justify-between h-28 border border-slate-200 rounded p-2">
              <p class="text-[10px] font-bold text-slate-600 uppercase">Pemohon / Teknisi</p>
              <div class="my-auto font-script text-slate-400 italic text-[11px]">Tercatat Digital</div>
              <div>
                <p class="font-bold text-[11px] text-slate-900 underline">{{ doc.simak?.worker?.full_name }}</p>
                <p class="text-[9px] text-slate-500">{{ doc.simak?.worker?.company?.name || 'PT INKA' }}</p>
              </div>
            </div>

            <!-- Signature 2: Approver -->
            <div class="flex flex-col justify-between h-28 border border-slate-200 rounded p-2">
              <p class="text-[10px] font-bold text-slate-600 uppercase">Atasan / Manajer Terkait</p>
              <div class="my-auto text-emerald-600 text-[10px] font-semibold flex items-center justify-center gap-1">
                <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 13l4 4L19 7" />
                </svg>
                <span>Disetujui</span>
              </div>
              <div>
                <p class="font-bold text-[11px] text-slate-900 underline">Pejabat Berwenang</p>
                <p class="text-[9px] text-slate-500">PT INKA (Persero)</p>
              </div>
            </div>

            <!-- Signature 3: Officer Issuer -->
            <div class="flex flex-col justify-between h-28 border border-slate-200 rounded p-2">
              <p class="text-[10px] font-bold text-slate-600 uppercase">Petugas Penerbit SIMAK</p>
              <div class="my-auto text-blue-600 text-[10px] font-semibold flex items-center justify-center gap-1">
                <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 13l4 4L19 7" />
                </svg>
                <span>Diterbitkan</span>
              </div>
              <div>
                <p class="font-bold text-[11px] text-slate-900 underline">SIMAK Officer</p>
                <p class="text-[9px] text-slate-500">Divisi Keamanan & Fasilitas</p>
              </div>
            </div>

            <!-- Signature 4: QR Security Stamp -->
            <div class="flex flex-col items-center justify-between h-28 border border-slate-300 rounded p-1.5 bg-slate-50">
              <p class="text-[9px] font-bold text-slate-700 uppercase">Pengaman Digital</p>
              <img :src="doc.qr_code_data" alt="QR Code" class="w-16 h-16 object-contain" />
              <p class="text-[8px] text-slate-500 text-center leading-tight">
                Pindai untuk verifikasi
              </p>
            </div>
          </div>

          <!-- Document Footer -->
          <div class="pt-3 flex items-center justify-between text-[9px] text-slate-400 font-mono">
            <span>Sistem SIMAK Digital PT INKA (Persero)</span>
            <span>TOKEN: {{ doc.simak?.qr_token }}</span>
            <span>Halaman 1 / 1</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
@media print {
  /* Hide all UI elements outside the printable paper sheet */
  :global(header),
  :global(aside),
  :global(nav),
  .no-print {
    display: none !important;
  }

  body, html {
    margin: 0 !important;
    padding: 0 !important;
    background: #ffffff !important;
  }

  .print-sheet {
    box-shadow: none !important;
    border: none !important;
    max-width: 100% !important;
    width: 100% !important;
    padding: 10mm 12mm !important;
    margin: 0 !important;
  }
}
</style>
