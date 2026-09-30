<script setup lang="ts">
import { ref, onMounted } from 'vue'
import {
  approvalService,
  type ApprovalMatrix,
  type CreateMatrixPayload,
} from '@/services/approval'
import { masterDataService, type Department, type Company } from '@/services/masterData'
import Button from '@/components/ui/Button.vue'
import Badge from '@/components/ui/Badge.vue'

const matrices = ref<ApprovalMatrix[]>([])
const departments = ref<Department[]>([])
const companies = ref<Company[]>([])
const loading = ref(false)
const modalOpen = ref(false)
const saving = ref(false)
const editingId = ref<string | null>(null)

// Form state
const form = ref<CreateMatrixPayload>({
  name: '',
  description: '',
  department_id: null,
  company_id: null,
  is_active: true,
  steps: [
    {
      step_order: 1,
      step_name: 'Pemeriksaan Manajer / PIC Terkait',
      required_role: 'Approver',
      specific_user_id: null,
    },
    {
      step_order: 2,
      step_name: 'Verifikasi Supervisor Pengawas Mutu & Area',
      required_role: 'Supervisor',
      specific_user_id: null,
    },
  ],
})

async function fetchData() {
  loading.value = true
  try {
    const [matrixRes, deptRes, compRes] = await Promise.all([
      approvalService.getMatrices(),
      masterDataService.getDepartments(),
      masterDataService.getCompanies(),
    ])

    if (matrixRes.data) matrices.value = matrixRes.data
    if (deptRes.data) departments.value = deptRes.data.items || []
    if (compRes.data) companies.value = compRes.data.items || []
  } catch (err) {
    console.error('Failed to load matrix data:', err)
  } finally {
    loading.value = false
  }
}

function handleOpenCreate() {
  editingId.value = null
  form.value = {
    name: '',
    description: '',
    department_id: null,
    company_id: null,
    is_active: true,
    steps: [
      {
        step_order: 1,
        step_name: 'Pemeriksaan Manajer / PIC Terkait',
        required_role: 'Approver',
        specific_user_id: null,
      },
      {
        step_order: 2,
        step_name: 'Verifikasi Supervisor Pengawas Mutu & Area',
        required_role: 'Supervisor',
        specific_user_id: null,
      },
    ],
  }
  modalOpen.value = true
}

function handleOpenEdit(matrix: ApprovalMatrix) {
  editingId.value = matrix.id
  form.value = {
    name: matrix.name,
    description: matrix.description || '',
    department_id: matrix.department_id || null,
    company_id: matrix.company_id || null,
    is_active: matrix.is_active,
    steps: matrix.steps.map((s, idx) => ({
      step_order: idx + 1,
      step_name: s.step_name,
      required_role: s.required_role,
      specific_user_id: s.specific_user_id || null,
    })),
  }
  modalOpen.value = true
}

function addStep() {
  const newOrder = form.value.steps.length + 1
  form.value.steps.push({
    step_order: newOrder,
    step_name: `Verifikasi Tingkat ${newOrder}`,
    required_role: 'Supervisor',
    specific_user_id: null,
  })
}

function removeStep(index: number) {
  if (form.value.steps.length <= 1) {
    alert('Matriks persetujuan minimal harus memiliki 1 (satu) langkah.')
    return
  }
  form.value.steps.splice(index, 1)
  // Reorder
  form.value.steps.forEach((s, idx) => {
    s.step_order = idx + 1
  })
}

async function handleSave() {
  if (!form.value.name.trim()) {
    alert('Nama matriks persetujuan wajib diisi.')
    return
  }

  for (const st of form.value.steps) {
    if (!st.step_name.trim()) {
      alert('Nama setiap langkah verifikasi wajib diisi.')
      return
    }
  }

  saving.value = true
  try {
    if (editingId.value) {
      await approvalService.updateMatrix(editingId.value, form.value)
    } else {
      await approvalService.createMatrix(form.value)
    }
    modalOpen.value = false
    await fetchData()
  } catch (err: any) {
    alert(err?.response?.data?.message || 'Gagal menyimpan matriks persetujuan.')
  } finally {
    saving.value = false
  }
}

async function handleDelete(matrix: ApprovalMatrix) {
  if (!confirm(`Apakah Anda yakin ingin menghapus matriks persetujuan "${matrix.name}"?`)) {
    return
  }
  try {
    await approvalService.deleteMatrix(matrix.id)
    await fetchData()
  } catch (err: any) {
    alert(err?.response?.data?.message || 'Gagal menghapus matriks.')
  }
}

onMounted(() => {
  fetchData()
})
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
      <div>
        <div class="flex items-center gap-2">
          <span class="inline-block w-2.5 h-2.5 rounded-full bg-red-600"></span>
          <p class="text-xs font-semibold uppercase tracking-wider text-red-600">Konfigurasi Sistem</p>
        </div>
        <h1 class="text-2xl font-bold text-slate-900 mt-1">Matriks Persetujuan SIMAK</h1>
        <p class="text-sm text-slate-500 mt-0.5">
          Atur hierarki rute dan tingkat persetujuan berjenjang per Departemen, Vendor, atau Global
        </p>
      </div>

      <Button variant="primary" class="gap-2 shadow-sm" @click="handleOpenCreate">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
        </svg>
        <span>Tambah Matriks Baru</span>
      </Button>
    </div>

    <!-- Matrix Cards List -->
    <div v-if="loading" class="text-center py-12 text-slate-400 text-sm">
      Memuat matriks persetujuan...
    </div>

    <div v-else class="grid grid-cols-1 gap-6">
      <div
        v-for="matrix in matrices"
        :key="matrix.id"
        class="bg-white rounded-xl border border-slate-200 shadow-sm overflow-hidden"
      >
        <!-- Card Header -->
        <div class="p-5 border-b border-slate-200 flex flex-col md:flex-row md:items-center md:justify-between gap-3 bg-slate-50/50">
          <div>
            <div class="flex items-center gap-2">
              <h3 class="text-base font-bold text-slate-900">{{ matrix.name }}</h3>
              <Badge :variant="matrix.is_active ? 'success' : 'neutral'" size="sm">
                {{ matrix.is_active ? 'Aktif' : 'Non-Aktif' }}
              </Badge>
              <span
                v-if="!matrix.department_id && !matrix.company_id"
                class="inline-flex items-center px-2 py-0.5 rounded text-[11px] font-semibold bg-red-50 text-red-700 border border-red-200"
              >
                Rute Standar PT INKA (Global)
              </span>
              <span
                v-else-if="matrix.department"
                class="inline-flex items-center px-2 py-0.5 rounded text-[11px] font-semibold bg-blue-50 text-blue-700 border border-blue-200"
              >
                Divisi: {{ matrix.department.name }}
              </span>
              <span
                v-else-if="matrix.company"
                class="inline-flex items-center px-2 py-0.5 rounded text-[11px] font-semibold bg-purple-50 text-purple-700 border border-purple-200"
              >
                Vendor: {{ matrix.company.name }}
              </span>
            </div>
            <p v-if="matrix.description" class="text-xs text-slate-500 mt-1 max-w-2xl">
              {{ matrix.description }}
            </p>
          </div>

          <div class="flex items-center gap-2">
            <Button variant="outline" size="sm" class="gap-1 text-slate-700" @click="handleOpenEdit(matrix)">
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
              </svg>
              <span>Ubah Rute</span>
            </Button>
            <Button
              v-if="matrix.department_id || matrix.company_id"
              variant="outline"
              size="sm"
              class="gap-1 text-red-600 border-red-200 hover:bg-red-50"
              @click="handleDelete(matrix)"
            >
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
              </svg>
              <span>Hapus</span>
            </Button>
          </div>
        </div>

        <!-- Steps Timeline / List -->
        <div class="p-5">
          <p class="text-xs font-bold text-slate-700 uppercase tracking-wider mb-3">
            Tahapan Rute Persetujuan ({{ matrix.steps?.length || 0 }} Tingkat)
          </p>

          <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
            <div
              v-for="st in matrix.steps"
              :key="st.id || st.step_order"
              class="p-3 rounded-lg border border-slate-200 bg-slate-50 flex items-start gap-3"
            >
              <div class="w-7 h-7 rounded-full bg-red-600 text-white font-bold text-xs flex items-center justify-center flex-shrink-0">
                {{ st.step_order }}
              </div>
              <div class="space-y-0.5">
                <p class="text-xs font-bold text-slate-800">{{ st.step_name }}</p>
                <div class="flex items-center gap-2 pt-0.5">
                  <span class="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-semibold bg-white border border-slate-200 text-slate-700">
                    Role: {{ st.required_role }}
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Create/Edit Modal -->
    <div
      v-if="modalOpen"
      class="fixed inset-0 z-50 bg-slate-900/60 backdrop-blur-xs flex items-center justify-center p-4 overflow-y-auto"
    >
      <div class="bg-white rounded-2xl max-w-xl w-full p-6 shadow-xl border border-slate-100 space-y-4 my-8">
        <div class="flex items-center justify-between pb-3 border-b border-slate-200">
          <div>
            <h3 class="text-base font-bold text-slate-900">
              {{ editingId ? 'Ubah Matriks Persetujuan' : 'Tambah Matriks Persetujuan Baru' }}
            </h3>
            <p class="text-xs text-slate-500">Konfigurasi rute hierarki persetujuan SIMAK</p>
          </div>
          <button class="text-slate-400 hover:text-slate-600 text-lg leading-none" @click="modalOpen = false">×</button>
        </div>

        <div class="space-y-4 text-xs">
          <!-- Name -->
          <div class="space-y-1">
            <label class="font-semibold text-slate-700">Nama Matriks <span class="text-red-600">*</span></label>
            <input
              v-model="form.name"
              type="text"
              placeholder="Contoh: Rute Khusus Divisi Fabrikasi"
              class="w-full p-2.5 bg-slate-50 border border-slate-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-red-500/20 focus:border-red-500"
            />
          </div>

          <!-- Description -->
          <div class="space-y-1">
            <label class="font-semibold text-slate-700">Keterangan / Catatan</label>
            <textarea
              v-model="form.description"
              rows="2"
              placeholder="Deskripsi singkat mengenai aturan matriks persetujuan ini..."
              class="w-full p-2.5 bg-slate-50 border border-slate-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-red-500/20 focus:border-red-500"
            ></textarea>
          </div>

          <!-- Scope: Department or Company -->
          <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
            <div class="space-y-1">
              <label class="font-semibold text-slate-700">Terapkan Khusus Divisi (Opsional)</label>
              <select
                v-model="form.department_id"
                class="w-full p-2.5 bg-slate-50 border border-slate-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-red-500/20 focus:border-red-500"
              >
                <option :value="null">-- Berlaku Semua Divisi --</option>
                <option v-for="d in departments" :key="d.id" :value="d.id">
                  {{ d.name }} ({{ d.code }})
                </option>
              </select>
            </div>

            <div class="space-y-1">
              <label class="font-semibold text-slate-700">Terapkan Khusus Vendor (Opsional)</label>
              <select
                v-model="form.company_id"
                class="w-full p-2.5 bg-slate-50 border border-slate-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-red-500/20 focus:border-red-500"
              >
                <option :value="null">-- Berlaku Semua Perusahaan --</option>
                <option v-for="c in companies" :key="c.id" :value="c.id">
                  {{ c.name }}
                </option>
              </select>
            </div>
          </div>

          <!-- Active checkbox -->
          <label class="flex items-center gap-2 cursor-pointer pt-1">
            <input
              v-model="form.is_active"
              type="checkbox"
              class="w-4 h-4 text-red-600 rounded border-slate-300 focus:ring-red-500"
            />
            <span class="text-slate-800 font-medium">Aktifkan matriks persetujuan ini</span>
          </label>

          <!-- Steps Builder -->
          <div class="pt-3 border-t border-slate-200 space-y-3">
            <div class="flex items-center justify-between">
              <label class="font-bold text-slate-900 uppercase tracking-wider text-[11px]">
                Urutan Langkah Persetujuan ({{ form.steps.length }} Tingkat)
              </label>
              <button
                type="button"
                class="text-red-600 hover:text-red-700 font-semibold text-xs inline-flex items-center gap-1"
                @click="addStep"
              >
                + Tambah Tingkat
              </button>
            </div>

            <div class="space-y-2.5 max-h-60 overflow-y-auto pr-1">
              <div
                v-for="(st, idx) in form.steps"
                :key="idx"
                class="p-3 rounded-lg border border-slate-200 bg-slate-50/70 flex items-center gap-2.5"
              >
                <div class="w-6 h-6 rounded-full bg-red-600 text-white font-bold text-xs flex items-center justify-center flex-shrink-0">
                  {{ st.step_order }}
                </div>
                <div class="flex-1 grid grid-cols-1 sm:grid-cols-2 gap-2">
                  <input
                    v-model="st.step_name"
                    type="text"
                    placeholder="Nama Langkah (misal: PIC Manajer)"
                    class="p-2 bg-white border border-slate-200 rounded text-xs focus:outline-none focus:ring-1 focus:ring-red-500"
                  />
                  <select
                    v-model="st.required_role"
                    class="p-2 bg-white border border-slate-200 rounded text-xs focus:outline-none focus:ring-1 focus:ring-red-500"
                  >
                    <option value="Approver">Role: Approver (Manajer/PIC)</option>
                    <option value="Supervisor">Role: Supervisor (Pengawas)</option>
                    <option value="Admin">Role: Admin</option>
                    <option value="SIMAK Officer">Role: SIMAK Officer</option>
                  </select>
                </div>
                <button
                  type="button"
                  class="p-1.5 text-slate-400 hover:text-red-600 rounded transition-colors"
                  title="Hapus tingkat ini"
                  @click="removeStep(idx)"
                >
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                  </svg>
                </button>
              </div>
            </div>
          </div>
        </div>

        <div class="flex items-center justify-end gap-2 pt-3 border-t border-slate-200">
          <Button variant="outline" size="sm" :disabled="saving" @click="modalOpen = false">
            Batal
          </Button>
          <Button variant="primary" size="sm" :loading="saving" @click="handleSave">
            Simpan Matriks
          </Button>
        </div>
      </div>
    </div>
  </div>
</template>
