<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { simakService, type Simak } from '@/services/simak'
import { documentService, type SimakDocument } from '@/services/document'
import Button from '@/components/ui/Button.vue'
import Badge from '@/components/ui/Badge.vue'
import Modal from '@/components/ui/Modal.vue'

const route = useRoute()
const router = useRouter()

const simak = ref<Simak | null>(null)
const loading = ref(false)

const showSubmitModal = ref(false)
const submitNotes = ref('')
const submitting = ref(false)

const showCancelModal = ref(false)
const cancelReason = ref('')
const cancelling = ref(false)

const showQrModal = ref(false)
const qrDoc = ref<SimakDocument | null>(null)
const qrLoading = ref(false)

async function openQrCodeModal() {
  if (!simak.value) return
  qrLoading.value = true
  showQrModal.value = true
  qrDoc.value = null
  try {
    const res = await documentService.getDocument(simak.value.id)
    if (res.data) {
      qrDoc.value = res.data
    }
  } catch (err: any) {
    alert(err?.response?.data?.message || 'Gagal memuat QR Code resmi.')
    showQrModal.value = false
  } finally {
    qrLoading.value = false
  }
}

function copyVerifyLink(qrToken?: string) {
  if (!qrToken) return
  const url = `${window.location.origin}/verify/${qrToken}`
  navigator.clipboard.writeText(url)
  alert('Tautan verifikasi keaslian berhasil disalin:\n' + url)
}

async function fetchDetail() {
  loading.value = true
  try {
    const res = await simakService.getSimak(route.params.id as string)
    if (res.data) {
      simak.value = res.data
    }
  } catch (err: any) {
    alert(err.response?.data?.message || 'Gagal memuat detail SIMAK')
    router.push('/simak')
  } finally {
    loading.value = false
  }
}

async function handleConfirmSubmit() {
  if (!simak.value) return
  submitting.value = true
  try {
    await simakService.submitSimak(simak.value.id, submitNotes.value)
    showSubmitModal.value = false
    alert('Permohonan SIMAK berhasil diajukan untuk ditinjau oleh Approver.')
    fetchDetail()
  } catch (err: any) {
    alert(err.response?.data?.message || 'Gagal mengajukan SIMAK')
  } finally {
    submitting.value = false
  }
}

async function handleConfirmCancel() {
  if (!simak.value) return
  cancelling.value = true
  try {
    await simakService.cancelSimak(simak.value.id, cancelReason.value)
    showCancelModal.value = false
    alert('Permohonan SIMAK telah dibatalkan.')
    fetchDetail()
  } catch (err: any) {
    alert(err.response?.data?.message || 'Gagal membatalkan SIMAK')
  } finally {
    cancelling.value = false
  }
}

async function handleDeleteDraft() {
  if (!simak.value) return
  if (confirm('Hapus permanen draft SIMAK ini?')) {
    try {
      await simakService.deleteSimak(simak.value.id)
      alert('Draft SIMAK berhasil dihapus.')
      router.push('/simak')
    } catch (err: any) {
      alert(err.response?.data?.message || 'Gagal menghapus draft')
    }
  }
}

function formatDate(iso: string) {
  if (!iso) return '-'
  return new Date(iso).toLocaleDateString('id-ID', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
  })
}

function getStatusBadge(status: string) {
  switch (status) {
    case 'DRAFT':
      return { variant: 'neutral' as const, label: 'Draft Pemohon', dot: false }
    case 'SUBMITTED':
      return { variant: 'warning' as const, label: 'Menunggu Review Approver', dot: true }
    case 'UNDER_REVIEW':
      return { variant: 'warning' as const, label: 'Sedang Ditinjau', dot: true }
    case 'REQUIRES_REVISION':
      return { variant: 'warning' as const, label: 'Perlu Revisi', dot: true }
    case 'REJECTED':
      return { variant: 'danger' as const, label: 'Permohonan Ditolak', dot: false }
    case 'APPROVED':
      return { variant: 'success' as const, label: 'Disetujui', dot: false }
    case 'ISSUED':
    case 'ACTIVE':
      return { variant: 'success' as const, label: 'SIMAK Aktif / Terbit', dot: true }
    case 'VERIFIED':
      return { variant: 'inka' as const, label: 'Terverifikasi Gate', dot: true }
    case 'COMPLETED':
      return { variant: 'neutral' as const, label: 'Selesai', dot: false }
    case 'CANCELLED':
      return { variant: 'neutral' as const, label: 'Dibatalkan', dot: false }
    case 'EXPIRED':
      return { variant: 'danger' as const, label: 'Kedaluwarsa', dot: false }
    default:
      return { variant: 'neutral' as const, label: status, dot: false }
  }
}

// Lifecycle step calculation
const currentStep = computed(() => {
  if (!simak.value) return 1
  switch (simak.value.status) {
    case 'DRAFT':
      return 1
    case 'SUBMITTED':
    case 'UNDER_REVIEW':
    case 'REQUIRES_REVISION':
      return 2
    case 'APPROVED':
      return 3
    case 'ISSUED':
    case 'ACTIVE':
      return 4
    case 'VERIFIED':
    case 'COMPLETED':
      return 5
    default:
      return 1
  }
})

onMounted(() => {
  fetchDetail()
})
</script>

<template>
  <div v-if="simak" class="space-y-6 max-w-5xl mx-auto pb-16">
    <!-- Breadcrumb -->
    <div class="flex items-center gap-2 text-xs text-slate-500">
      <router-link to="/" class="hover:text-slate-800">Dashboard</router-link>
      <span>/</span>
      <router-link to="/simak" class="hover:text-slate-800">Permohonan SIMAK</router-link>
      <span>/</span>
      <span class="text-slate-800 font-mono font-semibold">{{ simak.simak_number || 'Draft' }}</span>
    </div>

    <!-- Header Banner -->
    <div class="bg-white rounded-xl border border-slate-200 shadow-card p-5 space-y-4">
      <div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <div class="flex items-center gap-2.5 flex-wrap">
            <h1 class="text-xl font-bold font-mono text-slate-900">
              {{ simak.simak_number || 'DRAFT PERMOHONAN' }}
            </h1>
            <Badge
              :variant="getStatusBadge(simak.status).variant"
              :dot="getStatusBadge(simak.status).dot"
              size="sm"
            >
              {{ getStatusBadge(simak.status).label }}
            </Badge>
          </div>
          <p class="text-xs text-slate-500 mt-1">
            Dibuat pada {{ formatDate(simak.created_at) }} · Terakhir diperbarui {{ formatDate(simak.updated_at) }}
          </p>
        </div>

        <!-- Action buttons depending on state -->
        <div class="flex items-center gap-2 flex-wrap">
          <!-- Back button -->
          <router-link to="/simak">
            <Button variant="outline" size="sm">
              &larr; Kembali
            </Button>
          </router-link>

          <!-- Edit button (Draft or Requires Revision) -->
          <router-link
            v-if="simak.status === 'DRAFT' || simak.status === 'REQUIRES_REVISION'"
            :to="`/simak/${simak.id}/edit`"
          >
            <Button variant="outline" size="sm">
              Ubah Data
            </Button>
          </router-link>

          <!-- Delete Draft -->
          <Button
            v-if="simak.status === 'DRAFT'"
            variant="danger"
            size="sm"
            @click="handleDeleteDraft"
          >
            Hapus Draft
          </Button>

          <!-- Submit Draft -->
          <Button
            v-if="simak.status === 'DRAFT'"
            variant="primary"
            size="sm"
            @click="showSubmitModal = true"
          >
            Ajukan SIMAK
          </Button>

          <!-- Open Printable PDF Preview -->
          <router-link
            v-if="['ISSUED', 'ACTIVE', 'VERIFIED', 'COMPLETED'].includes(simak.status)"
            :to="`/simak/${simak.id}/preview`"
          >
            <Button variant="primary" size="sm" class="gap-1.5 shadow-sm">
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 17h2a2 2 0 002-2v-4a2 2 0 00-2-2H5a2 2 0 00-2 2v4a2 2 0 002 2h2m2 4h6a2 2 0 002-2v-4a2 2 0 00-2-2H9a2 2 0 00-2 2v4a2 2 0 002 2zm8-12V5a2 2 0 00-2-2H9a2 2 0 00-2 2v4h10z" />
              </svg>
              <span>Dokumen Resmi & PDF</span>
            </Button>
          </router-link>

          <!-- Show QR Code Modal -->
          <Button
            v-if="['ISSUED', 'ACTIVE', 'VERIFIED'].includes(simak.status)"
            variant="outline"
            size="sm"
            class="gap-1.5 text-slate-700"
            @click="openQrCodeModal"
          >
            <svg class="w-3.5 h-3.5 text-slate-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v1m6 11h2m-6 0h-2v4m0-11v3m0 0h.01M12 12h4.01M16 20h4M4 12h4m12 0h.01M5 8h2a1 1 0 001-1V5a1 1 0 00-1-1H5a1 1 0 00-1 1v2a1 1 0 001 1zm12 0h2a1 1 0 001-1V5a1 1 0 00-1-1h-2a1 1 0 00-1 1v2a1 1 0 001 1zM5 20h2a1 1 0 001-1v-2a1 1 0 00-1-1H5a1 1 0 00-1 1v2a1 1 0 001 1z" />
            </svg>
            <span>Tampilkan QR Pos</span>
          </Button>

          <!-- Cancel Permit (if pending) -->
          <Button
            v-if="['DRAFT', 'SUBMITTED', 'UNDER_REVIEW', 'REQUIRES_REVISION'].includes(simak.status)"
            variant="ghost"
            size="sm"
            class="text-rose-600 hover:bg-rose-50"
            @click="showCancelModal = true"
          >
            Batalkan
          </Button>
        </div>
      </div>

      <!-- Revision Alert Callout (if in REQUIRES_REVISION) -->
      <div
        v-if="simak.status === 'REQUIRES_REVISION' && simak.revision_notes"
        class="p-4 bg-amber-50 border-l-4 border-amber-500 rounded-r-lg text-amber-900 text-xs space-y-1"
      >
        <div class="flex items-center gap-2 font-bold">
          <svg class="w-4 h-4 text-amber-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
          </svg>
          Catatan Perbaikan dari Approver / Pemeriksa:
        </div>
        <p class="text-amber-800 leading-relaxed pl-6">{{ simak.revision_notes }}</p>
      </div>

      <!-- Rejection Alert Callout (if REJECTED) -->
      <div
        v-if="simak.status === 'REJECTED' && simak.rejection_reason"
        class="p-4 bg-rose-50 border-l-4 border-rose-500 rounded-r-lg text-rose-900 text-xs space-y-1"
      >
        <div class="flex items-center gap-2 font-bold">
          <svg class="w-4 h-4 text-rose-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
          Alasan Penolakan Permohonan:
        </div>
        <p class="text-rose-800 leading-relaxed pl-6">{{ simak.rejection_reason }}</p>
      </div>

      <!-- Lifecycle Progress Tracker -->
      <div class="pt-3 border-t border-slate-100">
        <div class="grid grid-cols-5 gap-2 text-center text-[11px] font-semibold">
          <div
            :class="[
              'p-2 rounded-lg border transition-all',
              currentStep >= 1 ? 'bg-inka-50 text-inka-700 border-inka-200' : 'bg-slate-50 text-slate-400 border-slate-100',
            ]"
          >
            1. Draft Pemohon
          </div>
          <div
            :class="[
              'p-2 rounded-lg border transition-all',
              currentStep >= 2 ? 'bg-inka-50 text-inka-700 border-inka-200' : 'bg-slate-50 text-slate-400 border-slate-100',
            ]"
          >
            2. Review Approver
          </div>
          <div
            :class="[
              'p-2 rounded-lg border transition-all',
              currentStep >= 3 ? 'bg-inka-50 text-inka-700 border-inka-200' : 'bg-slate-50 text-slate-400 border-slate-100',
            ]"
          >
            3. Disetujui
          </div>
          <div
            :class="[
              'p-2 rounded-lg border transition-all',
              currentStep >= 4 ? 'bg-inka-50 text-inka-700 border-inka-200' : 'bg-slate-50 text-slate-400 border-slate-100',
            ]"
          >
            4. Terbit & QR Code
          </div>
          <div
            :class="[
              'p-2 rounded-lg border transition-all',
              currentStep >= 5 ? 'bg-inka-50 text-inka-700 border-inka-200' : 'bg-slate-50 text-slate-400 border-slate-100',
            ]"
          >
            5. Selesai / Gate
          </div>
        </div>
      </div>
    </div>

    <!-- Main Grid Content: 2 Columns -->
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <!-- LEFT COLUMN: Permit Details & Items Table -->
      <div class="lg:col-span-2 space-y-6">
        <!-- Permit Information Card -->
        <div class="bg-white rounded-xl border border-slate-200 shadow-card p-5 space-y-4">
          <h2 class="text-sm font-bold text-slate-900 border-b border-slate-100 pb-2">
            Rincian Izin Membawa Alat Kerja
          </h2>

          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4 text-xs">
            <div>
              <span class="text-[10px] text-slate-400 font-semibold uppercase">Arah Aliran Barang</span>
              <div class="mt-1">
                <span
                  :class="[
                    'px-2.5 py-1 rounded text-xs font-bold uppercase',
                    simak.direction === 'MASUK'
                      ? 'bg-blue-50 text-blue-700 border border-blue-200'
                      : simak.direction === 'KELUAR'
                        ? 'bg-amber-50 text-amber-800 border border-amber-200'
                        : 'bg-purple-50 text-purple-700 border border-purple-200',
                  ]"
                >
                  {{ simak.direction === 'MASUK_KELUAR' ? 'Masuk & Keluar' : simak.direction }}
                </span>
              </div>
            </div>

            <div>
              <span class="text-[10px] text-slate-400 font-semibold uppercase">Masa Berlaku Izin</span>
              <p class="font-mono text-xs font-bold text-slate-900 mt-1">
                {{ formatDate(simak.start_date) }} s.d {{ formatDate(simak.end_date) }}
              </p>
            </div>

            <div class="sm:col-span-2">
              <span class="text-[10px] text-slate-400 font-semibold uppercase">Area Pabrik / Workshop Tujuan</span>
              <p class="text-xs font-bold text-slate-900 mt-1">
                {{ simak.destination_area }}
              </p>
            </div>

            <div class="sm:col-span-2">
              <span class="text-[10px] text-slate-400 font-semibold uppercase">Keperluan / Tujuan Kerja</span>
              <p class="text-xs text-slate-700 mt-1 leading-relaxed bg-slate-50 p-3 rounded-lg border border-slate-100">
                {{ simak.purpose }}
              </p>
            </div>

            <div v-if="simak.notes" class="sm:col-span-2">
              <span class="text-[10px] text-slate-400 font-semibold uppercase">Catatan Tambahan</span>
              <p class="text-xs text-slate-600 mt-0.5">{{ simak.notes }}</p>
            </div>
          </div>
        </div>

        <!-- Items Table Card -->
        <div class="bg-white rounded-xl border border-slate-200 shadow-card p-5 space-y-4">
          <div class="flex items-center justify-between border-b border-slate-100 pb-2">
            <h2 class="text-sm font-bold text-slate-900">
              Daftar Barang Bawaan Terdaftar ({{ simak.items ? simak.items.length : 0 }} Item)
            </h2>
            <span class="text-xs text-slate-400">Verifikasi fisik pos gerbang</span>
          </div>

          <div class="overflow-x-auto border border-slate-200 rounded-lg">
            <table class="w-full text-left text-xs border-collapse">
              <thead>
                <tr class="bg-slate-50 border-b border-slate-200 text-slate-600 font-bold uppercase tracking-wider text-[11px]">
                  <th class="py-2.5 px-3 w-10 text-center">No</th>
                  <th class="py-2.5 px-3">Nama Alat & Spesifikasi</th>
                  <th class="py-2.5 px-3">Kategori</th>
                  <th class="py-2.5 px-3 font-mono">No. Seri (S/N)</th>
                  <th class="py-2.5 px-3 text-center">Jumlah</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-100">
                <tr
                  v-for="(item, idx) in simak.items"
                  :key="item.id || idx"
                  class="hover:bg-slate-50/60 transition-colors"
                >
                  <td class="py-2.5 px-3 text-center text-slate-400 font-mono text-[11px]">
                    {{ idx + 1 }}
                  </td>
                  <td class="py-2.5 px-3">
                    <p class="font-bold text-slate-900">{{ item.item_name }}</p>
                    <p class="text-[11px] text-slate-500">
                      {{ item.brand || 'Tanpa Merk' }} {{ item.model ? `— ${item.model}` : '' }}
                      <span v-if="item.size_spec">({{ item.size_spec }})</span>
                    </p>
                    <p v-if="item.description" class="text-[11px] text-slate-400 mt-0.5">
                      {{ item.description }}
                    </p>
                  </td>
                  <td class="py-2.5 px-3">
                    <span class="px-2 py-0.5 rounded text-[10px] font-semibold bg-slate-100 text-slate-700">
                      {{ item.category?.name || 'Umum' }}
                    </span>
                  </td>
                  <td class="py-2.5 px-3 font-mono font-bold text-slate-800">
                    {{ item.serial_number || '-' }}
                  </td>
                  <td class="py-2.5 px-3 text-center">
                    <span class="font-bold text-slate-900 bg-slate-100 px-2 py-0.5 rounded border border-slate-200">
                      {{ item.quantity }} {{ item.unit }}
                    </span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <!-- RIGHT COLUMN: Worker Profile & Attachments -->
      <div class="space-y-6">
        <!-- Applicant Profile Card -->
        <div class="bg-white rounded-xl border border-slate-200 shadow-card p-5 space-y-4">
          <h2 class="text-sm font-bold text-slate-900 border-b border-slate-100 pb-2">
            Identitas Pemohon
          </h2>

          <div class="space-y-3 text-xs">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 rounded-full bg-inka-700 text-white flex items-center justify-center font-bold text-sm shrink-0">
                {{ (simak.worker?.full_name || 'P').charAt(0) }}
              </div>
              <div class="min-w-0">
                <p class="font-bold text-slate-900 truncate">{{ simak.worker?.full_name }}</p>
                <p class="text-[11px] text-slate-500 font-mono">NIP: {{ simak.worker?.worker_id }}</p>
              </div>
            </div>

            <div class="pt-2 border-t border-slate-100 space-y-2">
              <div>
                <span class="text-[10px] text-slate-400 uppercase font-semibold">Perusahaan / Vendor</span>
                <p class="font-bold text-slate-800 mt-0.5">{{ simak.worker?.company?.name || '-' }}</p>
              </div>

              <div>
                <span class="text-[10px] text-slate-400 uppercase font-semibold">Divisi / Departemen</span>
                <p class="font-bold text-slate-800 mt-0.5">{{ simak.worker?.department?.name || '-' }}</p>
              </div>

              <div>
                <span class="text-[10px] text-slate-400 uppercase font-semibold">Jabatan</span>
                <p class="text-slate-700 mt-0.5">{{ simak.worker?.position || '-' }}</p>
              </div>

              <div>
                <span class="text-[10px] text-slate-400 uppercase font-semibold">Kontak</span>
                <p class="text-slate-700 mt-0.5">{{ simak.worker?.phone || '-' }}</p>
              </div>
            </div>
          </div>
        </div>

        <!-- Attachments Card -->
        <div class="bg-white rounded-xl border border-slate-200 shadow-card p-5 space-y-4">
          <h2 class="text-sm font-bold text-slate-900 border-b border-slate-100 pb-2">
            Dokumen Lampiran Resmi
          </h2>

          <div v-if="!simak.attachments || simak.attachments.length === 0" class="text-xs text-slate-400 py-3 text-center">
            Tidak ada dokumen lampiran pendukung.
          </div>

          <div v-else class="space-y-2">
            <div
              v-for="att in simak.attachments"
              :key="att.id"
              class="flex items-center justify-between p-2.5 rounded-lg border border-slate-200 bg-slate-50 hover:bg-slate-100/60 transition-colors"
            >
              <div class="flex items-center gap-2.5 min-w-0">
                <svg class="w-5 h-5 text-red-600 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z" />
                </svg>
                <div class="min-w-0">
                  <p class="text-xs font-bold text-slate-800 truncate">{{ att.file_name }}</p>
                  <p class="text-[10px] text-slate-400 uppercase tracking-wider font-semibold">{{ att.attachment_type }}</p>
                </div>
              </div>

              <span class="text-[11px] font-semibold text-inka-700 hover:underline shrink-0">
                Unduh
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- MODAL: SUBMIT CONFIRMATION -->
    <Modal
      v-model="showSubmitModal"
      title="Ajukan Permohonan SIMAK"
      subtitle="Pastikan seluruh data peralatan kerja telah sesuai dan akurat"
    >
      <div class="space-y-4 text-xs">
        <p class="text-slate-700">
          Permohonan SIMAK ini akan dikirimkan ke antrean persetujuan (Approver) PT INKA. Setelah diajukan, data barang tidak dapat diubah kembali kecuali diminta revisi oleh Approver.
        </p>

        <div>
          <label class="block text-xs font-semibold text-slate-800 mb-1">Catatan untuk Approver (Opsional)</label>
          <input
            v-model="submitNotes"
            type="text"
            placeholder="misal: Mohon persetujuan segera karena pekerjaan dimulai besok pagi"
            class="w-full px-3 py-2 text-xs bg-slate-50 border border-slate-300 rounded-lg text-slate-900 focus:outline-none focus:ring-1 focus:ring-inka-600"
          />
        </div>
      </div>

      <template #footer>
        <Button variant="outline" size="sm" @click="showSubmitModal = false">
          Batal
        </Button>
        <Button
          variant="primary"
          size="sm"
          :disabled="submitting"
          @click="handleConfirmSubmit"
        >
          {{ submitting ? 'Mengajukan...' : 'Ya, Ajukan Permohonan' }}
        </Button>
      </template>
    </Modal>

    <!-- MODAL: CANCEL CONFIRMATION -->
    <Modal
      v-model="showCancelModal"
      title="Batalkan Permohonan SIMAK"
      subtitle="Tindakan ini akan menghentikan proses perizinan barang bawaan ini"
    >
      <div class="space-y-4 text-xs">
        <p class="text-slate-700">
          Apakah Anda yakin ingin membatalkan permohonan SIMAK ini?
        </p>

        <div>
          <label class="block text-xs font-semibold text-slate-800 mb-1">Alasan Pembatalan *</label>
          <textarea
            v-model="cancelReason"
            required
            rows="2"
            placeholder="Jelaskan alasan pembatalan permohonan..."
            class="w-full px-3 py-2 text-xs bg-slate-50 border border-slate-300 rounded-lg text-slate-900 focus:outline-none focus:ring-1 focus:ring-inka-600"
          ></textarea>
        </div>
      </div>

      <template #footer>
        <Button variant="outline" size="sm" @click="showCancelModal = false">
          Kembali
        </Button>
        <Button
          variant="danger"
          size="sm"
          :disabled="cancelling || !cancelReason.trim()"
          @click="handleConfirmCancel"
        >
          {{ cancelling ? 'Membatalkan...' : 'Batalkan SIMAK' }}
        </Button>
      </template>
    </Modal>

    <!-- MODAL: QR CODE VIEWER -->
    <Modal
      v-model="showQrModal"
      title="Kode QR Verifikasi Gerbang"
      subtitle="Tunjukkan kode QR ini kepada Petugas Keamanan (Gate Officer) saat di pos jaga"
    >
      <div v-if="qrLoading" class="py-8 text-center text-xs text-slate-400">
        Memuat data QR Code...
      </div>
      <div v-else-if="qrDoc" class="text-center space-y-4">
        <div class="p-3 bg-white border border-slate-200 rounded-xl inline-block shadow-xs">
          <img :src="qrDoc.qr_code_data" alt="QR Code SIMAK" class="w-48 h-48 mx-auto" />
        </div>
        <div>
          <p class="font-mono text-xs font-bold text-slate-900">{{ qrDoc.document_number }}</p>
          <p class="text-[11px] text-slate-500 mt-0.5">Diterbitkan: {{ formatDate(qrDoc.issued_at) }}</p>
        </div>
      </div>

      <template #footer>
        <Button variant="outline" size="sm" @click="copyVerifyLink(qrDoc?.simak?.qr_token)">
          Salin Tautan
        </Button>
        <Button variant="primary" size="sm" @click="showQrModal = false">
          Tutup
        </Button>
      </template>
    </Modal>
  </div>
</template>
