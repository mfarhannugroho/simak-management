<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { masterDataService, type Worker, type Company, type Department } from '@/services/masterData'
import DataTable, { type Column, type Pagination } from '@/components/ui/DataTable.vue'
import Button from '@/components/ui/Button.vue'
import Badge from '@/components/ui/Badge.vue'
import Modal from '@/components/ui/Modal.vue'

const workers = ref<Worker[]>([])
const loading = ref(false)
const search = ref('')
const companies = ref<Company[]>([])
const departments = ref<Department[]>([])

const pagination = ref<Pagination>({
  page: 1,
  limit: 10,
  total_items: 0,
  total_pages: 1,
})

const columns: Column[] = [
  { key: 'full_name', label: 'Nama Lengkap Pekerja', sortable: true },
  { key: 'worker_id', label: 'NIP / Worker ID' },
  { key: 'company', label: 'Perusahaan' },
  { key: 'department', label: 'Departemen' },
  { key: 'position', label: 'Jabatan / Posisi' },
  { key: 'phone', label: 'No. Kontak' },
  { key: 'actions', label: 'Aksi', align: 'right' },
]

const showModal = ref(false)
const editingWorker = ref<Worker | null>(null)
const form = ref({
  user_id: '',
  worker_id: '',
  full_name: '',
  company_id: '',
  department_id: '',
  position: '',
  phone: '',
})

async function fetchWorkers(page = 1) {
  loading.value = true
  try {
    const res = await masterDataService.getWorkers({
      page,
      limit: pagination.value.limit,
      search: search.value,
    })
    if (res.data) {
      workers.value = res.data.items
      pagination.value = res.data.pagination
    }
  } catch (err: any) {
    console.error('Failed to load workers:', err)
  } finally {
    loading.value = false
  }
}

async function loadOptions() {
  const [cRes, dRes] = await Promise.all([
    masterDataService.getCompanies({ limit: 100 }),
    masterDataService.getDepartments({ limit: 100 }),
  ])
  if (cRes.data) companies.value = cRes.data.items
  if (dRes.data) departments.value = dRes.data.items
}

function openEdit(row: Worker) {
  editingWorker.value = row
  form.value = {
    user_id: row.user_id,
    worker_id: row.worker_id,
    full_name: row.full_name,
    company_id: row.company_id || '',
    department_id: row.department_id || '',
    position: row.position || '',
    phone: row.phone || '',
  }
  showModal.value = true
}

async function saveWorker() {
  try {
    if (editingWorker.value) {
      await masterDataService.updateWorker(editingWorker.value.id, form.value)
    } else {
      await masterDataService.createWorker(form.value)
    }
    showModal.value = false
    fetchWorkers(pagination.value.page)
  } catch (err: any) {
    alert(err.response?.data?.message || 'Gagal menyimpan data pekerja')
  }
}

async function deleteWorker(id: string) {
  if (confirm('Hapus profil pekerja ini?')) {
    try {
      await masterDataService.deleteWorker(id)
      fetchWorkers(pagination.value.page)
    } catch (err: any) {
      alert(err.response?.data?.message || 'Gagal menghapus pekerja')
    }
  }
}

onMounted(() => {
  fetchWorkers()
  loadOptions()
})
</script>

<template>
  <div class="space-y-6">
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 bg-white p-5 rounded-xl border border-slate-200 shadow-card">
      <div>
        <div class="flex items-center gap-2">
          <h1 class="text-xl font-bold text-slate-900">Manajemen Pekerja & Profil</h1>
        </div>
        <p class="text-xs text-slate-500 mt-1">
          Daftar seluruh pekerja PT INKA maupun teknisi vendor rekanan yang berwenang membawa alat kerja ke lingkungan pabrik.
        </p>
      </div>

      <div class="flex items-center gap-2">
        <span class="text-xs text-slate-500">Total: <strong>{{ pagination.total_items }}</strong> Pekerja Terdaftar</span>
      </div>
    </div>

    <!-- Data Table Workers -->
    <DataTable
      :columns="columns"
      :items="workers"
      :loading="loading"
      :pagination="pagination"
      search-placeholder="Cari nama pekerja, NIP, no. telepon..."
      @search="(q) => { search = q; fetchWorkers(1) }"
      @page-change="(p) => fetchWorkers(p)"
    >
      <template #full_name="{ row, value }">
        <div class="flex items-center gap-2.5">
          <div class="w-7 h-7 rounded-full bg-slate-100 text-slate-700 flex items-center justify-center font-bold text-xs ring-1 ring-slate-200 shrink-0">
            {{ value.charAt(0) }}
          </div>
          <div>
            <p class="font-bold text-slate-900 text-xs">{{ value }}</p>
            <p class="text-[10px] text-slate-400">{{ row.position || 'Staf Teknis' }}</p>
          </div>
        </div>
      </template>

      <template #worker_id="{ value }">
        <span class="font-mono text-xs font-semibold px-2 py-0.5 rounded bg-slate-100 text-slate-800 border border-slate-200">
          {{ value }}
        </span>
      </template>

      <template #company="{ row }">
        <span class="text-xs text-slate-700">{{ row.company?.name || '-' }}</span>
      </template>

      <template #department="{ row }">
        <span class="text-xs text-slate-600">{{ row.department?.name || '-' }}</span>
      </template>

      <template #actions="{ row }">
        <div class="flex items-center justify-end gap-1.5">
          <button
            @click="openEdit(row)"
            class="px-2 py-1 text-[11px] font-semibold rounded text-slate-600 hover:bg-slate-100 hover:text-slate-900 cursor-pointer"
          >
            Ubah
          </button>
          <button
            @click="deleteWorker(row.id)"
            class="px-2 py-1 text-[11px] font-semibold rounded text-rose-600 hover:bg-rose-50 cursor-pointer"
          >
            Hapus
          </button>
        </div>
      </template>
    </DataTable>

    <!-- Modal Form Edit Worker -->
    <Modal
      v-model="showModal"
      title="Ubah Profil Pekerja"
      subtitle="Perbarui jabatan, penempatan departemen, dan kontak pekerja"
    >
      <form @submit.prevent="saveWorker" id="workerForm" class="space-y-4">
        <div>
          <label class="block text-xs font-semibold text-slate-700 mb-1">Nama Lengkap</label>
          <input
            v-model="form.full_name"
            required
            type="text"
            class="w-full px-3 py-2 text-xs bg-slate-50 border border-slate-300 rounded-lg text-slate-900 focus:outline-none focus:ring-1 focus:ring-inka-600 focus:bg-white"
          />
        </div>
        <div>
          <label class="block text-xs font-semibold text-slate-700 mb-1">NIP / Worker ID</label>
          <input
            v-model="form.worker_id"
            required
            type="text"
            class="w-full px-3 py-2 text-xs bg-slate-50 border border-slate-300 rounded-lg text-slate-900 focus:outline-none focus:ring-1 focus:ring-inka-600 focus:bg-white font-mono"
          />
        </div>
        <div class="grid grid-cols-2 gap-3">
          <div>
            <label class="block text-xs font-semibold text-slate-700 mb-1">Perusahaan</label>
            <select
              v-model="form.company_id"
              class="w-full px-3 py-2 text-xs bg-slate-50 border border-slate-300 rounded-lg text-slate-900 focus:outline-none focus:ring-1 focus:ring-inka-600 focus:bg-white"
            >
              <option value="">-- Pilih Perusahaan --</option>
              <option v-for="c in companies" :key="c.id" :value="c.id">{{ c.name }}</option>
            </select>
          </div>
          <div>
            <label class="block text-xs font-semibold text-slate-700 mb-1">Departemen / Divisi</label>
            <select
              v-model="form.department_id"
              class="w-full px-3 py-2 text-xs bg-slate-50 border border-slate-300 rounded-lg text-slate-900 focus:outline-none focus:ring-1 focus:ring-inka-600 focus:bg-white"
            >
              <option value="">-- Pilih Departemen --</option>
              <option v-for="d in departments" :key="d.id" :value="d.id">{{ d.name }}</option>
            </select>
          </div>
        </div>
        <div class="grid grid-cols-2 gap-3">
          <div>
            <label class="block text-xs font-semibold text-slate-700 mb-1">Jabatan / Posisi</label>
            <input
              v-model="form.position"
              type="text"
              class="w-full px-3 py-2 text-xs bg-slate-50 border border-slate-300 rounded-lg text-slate-900 focus:outline-none focus:ring-1 focus:ring-inka-600 focus:bg-white"
            />
          </div>
          <div>
            <label class="block text-xs font-semibold text-slate-700 mb-1">Nomor Telepon / WhatsApp</label>
            <input
              v-model="form.phone"
              type="text"
              class="w-full px-3 py-2 text-xs bg-slate-50 border border-slate-300 rounded-lg text-slate-900 focus:outline-none focus:ring-1 focus:ring-inka-600 focus:bg-white"
            />
          </div>
        </div>
      </form>

      <template #footer>
        <Button variant="outline" size="sm" @click="showModal = false">Batal</Button>
        <Button variant="primary" size="sm" type="submit" form="workerForm">Simpan Profil</Button>
      </template>
    </Modal>
  </div>
</template>
