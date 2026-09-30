<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { approvalService, type ApprovalDetailResponse, type SimakApprovalStep } from '@/services/approval'
import type { SimakItem } from '@/services/simak'
import Button from '@/components/ui/Button.vue'
import Badge from '@/components/ui/Badge.vue'

const route = useRoute()
const router = useRouter()
const simakId = computed(() => route.params.id as string)

const loading = ref(true)
const submitting = ref(false)
const detail = ref<ApprovalDetailResponse | null>(null)
const errorMsg = ref('')
const successMsg = ref('')

// Decision state
const selectedDecision = ref<'APPROVE' | 'REQUEST_REVISION' | 'REJECT'>('APPROVE')
const comments = ref('')
const confirmModalOpen = ref(false)

// Item preview photo modal
const activePhotoUrl = ref<string | null>(null)

async function fetchDetail() {
  loading.value = true
  errorMsg.value = ''
  try {
    const res = await approvalService.getApprovalDetail(simakId.value)
    if (res.data) {
      detail.value = res.data
    }
  } catch (err: any) {
    errorMsg.value = err?.response?.data?.message || 'Gagal memuat detail permohonan SIMAK'
  } finally {
    loading.value = false
  }
}

function handleOpenConfirm(decision: 'APPROVE' | 'REQUEST_REVISION' | 'REJECT') {
  selectedDecision.value = decision
  if (decision === 'APPROVE' && !comments.value) {
    comments.value = 'Disetujui sesuai regulasi dan protokol kerja PT INKA (Persero).'
  }
  confirmModalOpen.value = true
}

async function executeDecision() {
  if ((selectedDecision.value === 'REJECT' || selectedDecision.value === 'REQUEST_REVISION') && !comments.value.trim()) {
    alert('Catatan atau alasan wajib diisi untuk penolakan atau permintaan revisi.')
    return
  }

  submitting.value = true
  try {
    await approvalService.submitDecision(simakId.value, {
      decision: selectedDecision.value,
      comments: comments.value,
    })

    confirmModalOpen.value = false
    successMsg.value = `Keputusan berhasil disimpan: Permohonan telah di-${
      selectedDecision.value === 'APPROVE' ? 'setujui' : selectedDecision.value === 'REJECT' ? 'tolak' : 'minta revisi'
    }.`

    // Refresh detail
    await fetchDetail()
  } catch (err: any) {
    alert(err?.response?.data?.message || 'Gagal menyimpan keputusan persetujuan.')
  } finally {
    submitting.value = false
  }
}

function formatDate(dateStr?: string) {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'long',
    year: 'numeric',
  })
}

function formatDateTime(dateStr?: string) {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function getStepBadge(status: string) {
  switch (status) {
    case 'APPROVED':
      return { variant: 'success' as const, label: 'Disetujui' }
    case 'REJECTED':
      return { variant: 'danger' as const, label: 'Ditolak' }
    case 'REVISION_REQUESTED':
      return { variant: 'warning' as const, label: 'Perlu Revisi' }
    case 'PENDING':
      return { variant: 'inka' as const, label: 'Sedang Diperiksa' }
    case 'WAITING':
    default:
      return { variant: 'neutral' as const, label: 'Menunggu Giliran' }
  }
}

onMounted(() => {
  fetchDetail()
})
</script>

<template>
  <div class="space-y-6">
    <!-- Back & Header -->
    <div class="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
      <div class="flex items-center gap-3">
        <button
          type="button"
          class="p-2 rounded-lg border border-slate-200 bg-white text-slate-600 hover:bg-slate-50 transition-colors"
          @click="router.push('/approvals')"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
          </svg>
        </button>
        <div>
          <div class="flex items-center gap-2">
            <span class="text-xs font-semibold text-slate-500 uppercase tracking-wider">Tinjauan Persetujuan</span>
            <span class="text-slate-300">•</span>
            <span class="font-mono text-sm font-bold text-slate-900">{{ detail?.simak?.simak_number || 'DRAFT' }}</span>
          </div>
          <h1 class="text-2xl font-bold text-slate-900 mt-0.5">
            {{ detail?.simak?.purpose || 'Memuat Permohonan...' }}
          </h1>
        </div>
      </div>
    </div>

    <!-- Alert / Notifications -->
    <div v-if="successMsg" class="p-4 rounded-xl bg-emerald-50 border border-emerald-200 text-emerald-800 text-sm flex items-center justify-between">
      <div class="flex items-center gap-2">
        <svg class="w-5 h-5 text-emerald-600 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
        </svg>
        <span>{{ successMsg }}</span>
      </div>
      <button class="text-emerald-700 hover:text-emerald-900 font-bold" @click="successMsg = ''">×</button>
    </div>

    <div v-if="errorMsg" class="p-4 rounded-xl bg-red-50 border border-red-200 text-red-800 text-sm flex items-center gap-2">
      <svg class="w-5 h-5 text-red-600 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
      </svg>
      <span>{{ errorMsg }}</span>
    </div>

    <!-- Segregation of Duties (SoD) Warning Banner -->
    <div
      v-if="detail && !detail.can_take_action && detail.restriction_note && detail.restriction_note.includes('Segregation of Duties')"
      class="p-4 rounded-xl bg-amber-50 border border-amber-200 flex items-start gap-3 shadow-sm"
    >
      <div class="p-2 rounded-lg bg-amber-100 text-amber-700 mt-0.5">
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
        </svg>
      </div>
      <div>
        <h4 class="text-sm font-bold text-amber-900">Kebijakan Segregation of Duties (SoD) Aktif</h4>
        <p class="text-xs text-amber-800 mt-1 leading-relaxed">
          {{ detail.restriction_note }} Anda dapat meninjau detail permohonan ini namun persetujuan harus diberikan oleh pejabat berwenang lain yang independen.
        </p>
      </div>
    </div>

    <!-- Main Content Layout -->
    <div v-if="!loading && detail" class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <!-- Left Column: Details & Items (2 Cols) -->
      <div class="lg:col-span-2 space-y-6">
        <!-- Applicant Profile Card -->
        <div class="bg-white rounded-xl border border-slate-200 shadow-sm p-5">
          <h3 class="text-sm font-bold text-slate-900 uppercase tracking-wider mb-4 flex items-center gap-2">
            <svg class="w-4 h-4 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
            </svg>
            Profil Pemohon & Instansi
          </h3>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div class="space-y-1">
              <p class="text-xs text-slate-500">Nama Lengkap</p>
              <p class="text-sm font-semibold text-slate-900">{{ detail.simak.worker?.full_name || 'Ahmad Wahyudi, A.Md' }}</p>
            </div>
            <div class="space-y-1">
              <p class="text-xs text-slate-500">ID Pekerja / Badge</p>
              <p class="text-sm font-mono font-medium text-slate-800">{{ detail.simak.worker?.worker_id || '-' }}</p>
            </div>
            <div class="space-y-1">
              <p class="text-xs text-slate-500">Perusahaan / Vendor</p>
              <p class="text-sm font-medium text-slate-800">{{ detail.simak.worker?.company?.name || 'PT INKA (Persero)' }}</p>
            </div>
            <div class="space-y-1">
              <p class="text-xs text-slate-500">Departemen / Divisi</p>
              <p class="text-sm font-medium text-slate-800">{{ detail.simak.worker?.department?.name || 'Divisi Fabrikasi Kereta' }}</p>
            </div>
            <div class="space-y-1">
              <p class="text-xs text-slate-500">Posisi / Jabatan</p>
              <p class="text-sm font-medium text-slate-800">{{ detail.simak.worker?.position || 'Teknisi Perakitan' }}</p>
            </div>
            <div class="space-y-1">
              <p class="text-xs text-slate-500">Nomor Telepon</p>
              <p class="text-sm font-medium text-slate-800">{{ detail.simak.worker?.phone || '-' }}</p>
            </div>
          </div>
        </div>

        <!-- Purpose & Location Card -->
        <div class="bg-white rounded-xl border border-slate-200 shadow-sm p-5">
          <h3 class="text-sm font-bold text-slate-900 uppercase tracking-wider mb-4 flex items-center gap-2">
            <svg class="w-4 h-4 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
            </svg>
            Ruang Lingkup & Waktu Izin
          </h3>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div class="space-y-1 md:col-span-2">
              <p class="text-xs text-slate-500">Keperluan Pekerjaan / Aktivitas</p>
              <p class="text-sm text-slate-800 bg-slate-50 p-3 rounded-lg border border-slate-100 leading-relaxed font-medium">
                {{ detail.simak.purpose }}
              </p>
            </div>
            <div class="space-y-1">
              <p class="text-xs text-slate-500">Area Plant / Lokasi Tujuan</p>
              <p class="text-sm font-semibold text-slate-900 flex items-center gap-1.5">
                <svg class="w-4 h-4 text-red-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
                </svg>
                {{ detail.simak.destination_area }}
              </p>
            </div>
            <div class="space-y-1">
              <p class="text-xs text-slate-500">Arah Aliran Peralatan</p>
              <span class="inline-flex items-center px-2.5 py-1 rounded-md text-xs font-semibold bg-blue-50 text-blue-700 border border-blue-200">
                {{ detail.simak.direction }}
              </span>
            </div>
            <div class="space-y-1">
              <p class="text-xs text-slate-500">Tanggal Mulai Berlaku</p>
              <p class="text-sm font-medium text-slate-800">{{ formatDate(detail.simak.start_date) }}</p>
            </div>
            <div class="space-y-1">
              <p class="text-xs text-slate-500">Tanggal Selesai Berlaku</p>
              <p class="text-sm font-medium text-slate-800">{{ formatDate(detail.simak.end_date) }}</p>
            </div>
          </div>
        </div>

        <!-- Carried Equipment / Items Table -->
        <div class="bg-white rounded-xl border border-slate-200 shadow-sm overflow-hidden">
          <div class="p-5 border-b border-slate-200 flex items-center justify-between">
            <div>
              <h3 class="text-sm font-bold text-slate-900 uppercase tracking-wider flex items-center gap-2">
                <svg class="w-4 h-4 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
                </svg>
                Daftar Peralatan & Barang Bawaan
              </h3>
              <p class="text-xs text-slate-500 mt-0.5">Total {{ detail.simak.items?.length || 0 }} item terdaftar</p>
            </div>
          </div>

          <div class="overflow-x-auto">
            <table class="w-full text-left text-xs">
              <thead class="bg-slate-50 text-slate-500 border-b border-slate-200">
                <tr>
                  <th class="py-3 px-4 font-semibold w-12 text-center">No</th>
                  <th class="py-3 px-4 font-semibold">Nama Alat & Kategori</th>
                  <th class="py-3 px-4 font-semibold">Merek / Tipe</th>
                  <th class="py-3 px-4 font-semibold">Nomor Seri</th>
                  <th class="py-3 px-4 font-semibold text-center">Jumlah</th>
                  <th class="py-3 px-4 font-semibold">Spesifikasi</th>
                  <th class="py-3 px-4 font-semibold text-center">Foto</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-100">
                <tr v-for="(item, idx) in detail.simak.items" :key="item.id || idx" class="hover:bg-slate-50/70 transition-colors">
                  <td class="py-3 px-4 text-center font-medium text-slate-400">{{ idx + 1 }}</td>
                  <td class="py-3 px-4">
                    <p class="font-semibold text-slate-900">{{ item.item_name }}</p>
                    <p class="text-[11px] text-slate-500">{{ item.category?.name || 'Peralatan Umum' }}</p>
                  </td>
                  <td class="py-3 px-4 text-slate-700">
                    <span v-if="item.brand">{{ item.brand }} {{ item.model || '' }}</span>
                    <span v-else class="text-slate-400 italic">-</span>
                  </td>
                  <td class="py-3 px-4">
                    <span v-if="item.serial_number" class="font-mono text-slate-800 bg-slate-100 px-1.5 py-0.5 rounded text-[11px]">
                      {{ item.serial_number }}
                    </span>
                    <span v-else class="text-slate-400 italic">-</span>
                  </td>
                  <td class="py-3 px-4 text-center">
                    <span class="inline-flex items-center px-2 py-0.5 rounded bg-slate-100 text-slate-800 font-semibold">
                      {{ item.quantity }} {{ item.unit }}
                    </span>
                  </td>
                  <td class="py-3 px-4 text-slate-600">
                    {{ item.size_spec || item.description || '-' }}
                  </td>
                  <td class="py-3 px-4 text-center">
                    <button
                      v-if="item.photo_url"
                      type="button"
                      class="text-red-600 hover:text-red-700 font-medium inline-flex items-center gap-1 text-[11px]"
                      @click="activePhotoUrl = item.photo_url"
                    >
                      <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                      </svg>
                      Lihat
                    </button>
                    <span v-else class="text-slate-300">-</span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <!-- Attachments Section -->
        <div v-if="detail.simak.attachments && detail.simak.attachments.length > 0" class="bg-white rounded-xl border border-slate-200 shadow-sm p-5">
          <h3 class="text-sm font-bold text-slate-900 uppercase tracking-wider mb-3 flex items-center gap-2">
            <svg class="w-4 h-4 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.172 7l-6.586 6.586a2 2 0 102.828 2.828l6.414-6.586a4 4 0 00-5.656-5.656l-6.415 6.585a6 6 0 108.486 8.486L20.5 13" />
            </svg>
            Dokumen Lampiran / Surat Tugas
          </h3>
          <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
            <div
              v-for="att in detail.simak.attachments"
              :key="att.id"
              class="p-3 rounded-lg border border-slate-200 bg-slate-50 flex items-center justify-between"
            >
              <div class="flex items-center gap-2.5">
                <div class="w-8 h-8 rounded bg-red-100 text-red-600 flex items-center justify-center font-bold text-xs">
                  PDF
                </div>
                <div>
                  <p class="text-xs font-semibold text-slate-800 line-clamp-1">{{ att.file_name }}</p>
                  <p class="text-[10px] text-slate-400">{{ att.attachment_type }}</p>
                </div>
              </div>
              <a
                :href="att.file_url"
                target="_blank"
                class="px-2.5 py-1 text-[11px] font-medium text-slate-600 hover:text-slate-900 border border-slate-200 rounded hover:bg-white transition-colors"
              >
                Buka
              </a>
            </div>
          </div>
        </div>
      </div>

      <!-- Right Column: Approval Stepper & Decision Action Card (1 Col) -->
      <div class="space-y-6">
        <!-- Visual Stepper Workflow -->
        <div class="bg-white rounded-xl border border-slate-200 shadow-sm p-5">
          <h3 class="text-sm font-bold text-slate-900 uppercase tracking-wider mb-4 flex items-center gap-2">
            <svg class="w-4 h-4 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
            </svg>
            Alur Persetujuan Bertingkat
          </h3>

          <div class="space-y-4">
            <div
              v-for="(st, i) in detail.workflow_steps"
              :key="st.id || i"
              class="relative pl-6 pb-4 border-l-2 last:border-l-0"
              :class="st.status === 'APPROVED' ? 'border-emerald-500' : st.status === 'PENDING' ? 'border-blue-500' : 'border-slate-200'"
            >
              <!-- Stepper Dot -->
              <div
                class="absolute -left-2 top-0 w-4 h-4 rounded-full border-2 bg-white flex items-center justify-center"
                :class="
                  st.status === 'APPROVED'
                    ? 'border-emerald-500 bg-emerald-50 text-emerald-600'
                    : st.status === 'PENDING'
                    ? 'border-blue-600 bg-blue-50 text-blue-600'
                    : st.status === 'REJECTED'
                    ? 'border-red-500 bg-red-50 text-red-600'
                    : st.status === 'REVISION_REQUESTED'
                    ? 'border-amber-500 bg-amber-50 text-amber-600'
                    : 'border-slate-300 bg-slate-100 text-slate-400'
                "
              >
                <div
                  v-if="st.status === 'APPROVED'"
                  class="w-1.5 h-1.5 rounded-full bg-emerald-500"
                ></div>
                <div
                  v-else-if="st.status === 'PENDING'"
                  class="w-1.5 h-1.5 rounded-full bg-blue-600 animate-ping"
                ></div>
              </div>

              <!-- Step Content -->
              <div class="space-y-1">
                <div class="flex items-center justify-between">
                  <span class="text-[10px] font-bold uppercase tracking-wider text-slate-400">
                    Tingkat {{ st.step_order }} • {{ st.required_role }}
                  </span>
                  <Badge :variant="getStepBadge(st.status).variant" size="sm">
                    {{ getStepBadge(st.status).label }}
                  </Badge>
                </div>
                <p class="text-xs font-bold text-slate-800">{{ st.step_name }}</p>

                <!-- Action summary if executed -->
                <div v-if="st.action_at" class="text-[11px] text-slate-500 bg-slate-50 p-2 rounded border border-slate-100 mt-1 space-y-0.5">
                  <p class="font-medium text-slate-700">Oleh: {{ st.action_by_name || 'Approver Terkait' }}</p>
                  <p class="text-slate-400 text-[10px]">{{ formatDateTime(st.action_at) }}</p>
                  <p v-if="st.comments" class="italic text-slate-600 mt-1">"{{ st.comments }}"</p>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Decision Box -->
        <div class="bg-white rounded-xl border border-slate-200 shadow-sm p-5 space-y-4">
          <h3 class="text-sm font-bold text-slate-900 uppercase tracking-wider flex items-center gap-2">
            <svg class="w-4 h-4 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
            </svg>
            Keputusan Pemeriksa
          </h3>

          <!-- When user is authorized to take action -->
          <div v-if="detail.can_take_action" class="space-y-4">
            <p class="text-xs text-slate-600 leading-relaxed">
              Anda ditugaskan memeriksa permohonan ini pada tahap
              <span class="font-bold text-slate-900">{{ detail.current_approval?.step_name }}</span>.
              Silakan periksa kesesuaian barang bawaan dan pilih tindakan:
            </p>

            <div class="space-y-2">
              <label class="block text-xs font-semibold text-slate-700">Catatan / Komentar Keputusan</label>
              <textarea
                v-model="comments"
                rows="3"
                placeholder="Tuliskan catatan verifikasi, instruksi revisi, atau alasan jika ditolak..."
                class="w-full p-2.5 text-xs bg-slate-50 border border-slate-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-red-500/20 focus:border-red-500"
              ></textarea>
            </div>

            <div class="flex flex-col gap-2 pt-2">
              <!-- Approve Button -->
              <Button
                variant="primary"
                class="w-full justify-center gap-2 shadow-sm bg-emerald-600 hover:bg-emerald-700 border-emerald-600"
                @click="handleOpenConfirm('APPROVE')"
              >
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                </svg>
                <span>Setujui Permohonan</span>
              </Button>

              <!-- Request Revision Button -->
              <Button
                variant="outline"
                class="w-full justify-center gap-2 text-amber-700 border-amber-300 hover:bg-amber-50"
                @click="handleOpenConfirm('REQUEST_REVISION')"
              >
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                </svg>
                <span>Minta Revisi ke Pemohon</span>
              </Button>

              <!-- Reject Button -->
              <Button
                variant="outline"
                class="w-full justify-center gap-2 text-red-700 border-red-200 hover:bg-red-50"
                @click="handleOpenConfirm('REJECT')"
              >
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
                <span>Tolak Permohonan</span>
              </Button>
            </div>
          </div>

          <!-- When user cannot take action -->
          <div v-else class="p-3 bg-slate-50 rounded-lg border border-slate-100 text-xs text-slate-500 space-y-2">
            <div class="flex items-center gap-2 text-slate-700 font-semibold">
              <svg class="w-4 h-4 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <span>Status Keputusan</span>
            </div>
            <p>{{ detail.restriction_note || 'Tidak ada aksi yang dapat dilakukan pada tahap ini.' }}</p>
          </div>
        </div>
      </div>
    </div>

    <!-- Confirmation Modal -->
    <div
      v-if="confirmModalOpen"
      class="fixed inset-0 z-50 bg-slate-900/60 backdrop-blur-xs flex items-center justify-center p-4"
    >
      <div class="bg-white rounded-2xl max-w-md w-full p-6 shadow-xl border border-slate-100 space-y-4">
        <div class="flex items-center gap-3">
          <div
            class="w-10 h-10 rounded-full flex items-center justify-center flex-shrink-0"
            :class="
              selectedDecision === 'APPROVE'
                ? 'bg-emerald-100 text-emerald-600'
                : selectedDecision === 'REQUEST_REVISION'
                ? 'bg-amber-100 text-amber-600'
                : 'bg-red-100 text-red-600'
            "
          >
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                v-if="selectedDecision === 'APPROVE'"
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M5 13l4 4L19 7"
              />
              <path
                v-else-if="selectedDecision === 'REQUEST_REVISION'"
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
              />
              <path
                v-else
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M6 18L18 6M6 6l12 12"
              />
            </svg>
          </div>
          <div>
            <h4 class="text-base font-bold text-slate-900">
              Konfirmasi {{ selectedDecision === 'APPROVE' ? 'Persetujuan' : selectedDecision === 'REJECT' ? 'Penolakan' : 'Permintaan Revisi' }}
            </h4>
            <p class="text-xs text-slate-500">
              SIMAK {{ detail?.simak?.simak_number }}
            </p>
          </div>
        </div>

        <div class="text-xs text-slate-600 bg-slate-50 p-3 rounded-lg border border-slate-100 space-y-1">
          <p class="font-semibold text-slate-800">Catatan yang akan disimpan:</p>
          <p class="italic text-slate-700">"{{ comments || '-' }}"</p>
        </div>

        <p class="text-xs text-slate-500">
          Apakah Anda yakin ingin menyimpan keputusan ini? Tindakan ini akan dicatat dalam audit trail persetujuan resmi.
        </p>

        <div class="flex items-center justify-end gap-2 pt-2">
          <Button
            variant="outline"
            size="sm"
            :disabled="submitting"
            @click="confirmModalOpen = false"
          >
            Batal
          </Button>
          <Button
            :variant="selectedDecision === 'APPROVE' ? 'primary' : 'danger'"
            size="sm"
            :loading="submitting"
            @click="executeDecision"
          >
            Ya, Lanjutkan
          </Button>
        </div>
      </div>
    </div>

    <!-- Photo Preview Modal -->
    <div
      v-if="activePhotoUrl"
      class="fixed inset-0 z-50 bg-slate-900/70 backdrop-blur-xs flex items-center justify-center p-4"
      @click.self="activePhotoUrl = null"
    >
      <div class="bg-white rounded-xl max-w-lg w-full p-4 shadow-xl space-y-3">
        <div class="flex items-center justify-between pb-2 border-b border-slate-200">
          <p class="text-xs font-bold text-slate-800">Foto Fisik Barang Bawaan</p>
          <button class="text-slate-400 hover:text-slate-600 text-lg leading-none" @click="activePhotoUrl = null">×</button>
        </div>
        <div class="rounded-lg overflow-hidden bg-slate-100 flex items-center justify-center max-h-96">
          <img :src="activePhotoUrl" alt="Foto Barang" class="max-h-96 w-auto object-contain" />
        </div>
      </div>
    </div>
  </div>
</template>
