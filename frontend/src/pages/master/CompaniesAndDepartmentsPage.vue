<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { masterDataService, type Company, type Department } from '@/services/masterData'
import DataTable, { type Column, type Pagination } from '@/components/ui/DataTable.vue'
import Button from '@/components/ui/Button.vue'
import Badge from '@/components/ui/Badge.vue'
import Modal from '@/components/ui/Modal.vue'

const activeTab = ref<'companies' | 'departments'>('companies')

// ─────────────────────────────────────────────────────────────────────────────
// Companies State
// ─────────────────────────────────────────────────────────────────────────────
const companies = ref<Company[]>([])
const loadingCompanies = ref(false)
const companySearch = ref('')
const companyPagination = ref<Pagination>({
  page: 1,
  limit: 10,
  total_items: 0,
  total_pages: 1,
})

const companyColumns: Column[] = [
  { key: 'name', label: 'Nama Perusahaan', sortable: true },
  { key: 'type', label: 'Tipe Instansi' },
  { key: 'contact_phone', label: 'Kontak Telepon' },
  { key: 'contact_email', label: 'Email' },
  { key: 'is_active', label: 'Status', align: 'center' },
  { key: 'actions', label: 'Aksi', align: 'right' },
]

const showCompanyModal = ref(false)
const editingCompany = ref<Company | null>(null)
const companyForm = ref({
  name: '',
  type: 'VENDOR',
  address: '',
  contact_phone: '',
  contact_email: '',
  is_active: true,
})

async function fetchCompanies(page = 1) {
  loadingCompanies.value = true
  try {
    const res = await masterDataService.getCompanies({
      page,
      limit: companyPagination.value.limit,
      search: companySearch.value,
    })
    if (res.data) {
      companies.value = res.data.items
      companyPagination.value = res.data.pagination
    }
  } catch (err: any) {
    console.error('Failed to fetch companies:', err)
  } finally {
    loadingCompanies.value = false
  }
}

function openCreateCompany() {
  editingCompany.value = null
  companyForm.value = {
    name: '',
    type: 'VENDOR',
    address: '',
    contact_phone: '',
    contact_email: '',
    is_active: true,
  }
  showCompanyModal.value = true
}

function openEditCompany(row: Company) {
  editingCompany.value = row
  companyForm.value = {
    name: row.name,
    type: row.type || 'VENDOR',
    address: row.address || '',
    contact_phone: row.contact_phone || '',
    contact_email: row.contact_email || '',
    is_active: row.is_active,
  }
  showCompanyModal.value = true
}

async function saveCompany() {
  try {
    if (editingCompany.value) {
      await masterDataService.updateCompany(editingCompany.value.id, companyForm.value)
    } else {
      await masterDataService.createCompany(companyForm.value)
    }
    showCompanyModal.value = false
    fetchCompanies(companyPagination.value.page)
  } catch (err: any) {
    alert(err.response?.data?.message || 'Gagal menyimpan perusahaan')
  }
}

async function deleteCompany(id: string) {
  if (confirm('Hapus perusahaan ini?')) {
    try {
      await masterDataService.deleteCompany(id)
      fetchCompanies(companyPagination.value.page)
    } catch (err: any) {
      alert(err.response?.data?.message || 'Gagal menghapus perusahaan')
    }
  }
}

// ─────────────────────────────────────────────────────────────────────────────
// Departments State
// ─────────────────────────────────────────────────────────────────────────────
const departments = ref<Department[]>([])
const loadingDepartments = ref(false)
const departmentSearch = ref('')
const departmentPagination = ref<Pagination>({
  page: 1,
  limit: 10,
  total_items: 0,
  total_pages: 1,
})

const departmentColumns: Column[] = [
  { key: 'name', label: 'Nama Divisi / Departemen', sortable: true },
  { key: 'code', label: 'Kode', align: 'center' },
  { key: 'is_active', label: 'Status', align: 'center' },
  { key: 'actions', label: 'Aksi', align: 'right' },
]

const showDepartmentModal = ref(false)
const editingDepartment = ref<Department | null>(null)
const departmentForm = ref({
  name: '',
  code: '',
  is_active: true,
})

async function fetchDepartments(page = 1) {
  loadingDepartments.value = true
  try {
    const res = await masterDataService.getDepartments({
      page,
      limit: departmentPagination.value.limit,
      search: departmentSearch.value,
    })
    if (res.data) {
      departments.value = res.data.items
      departmentPagination.value = res.data.pagination
    }
  } catch (err: any) {
    console.error('Failed to fetch departments:', err)
  } finally {
    loadingDepartments.value = false
  }
}

function openCreateDepartment() {
  editingDepartment.value = null
  departmentForm.value = {
    name: '',
    code: '',
    is_active: true,
  }
  showDepartmentModal.value = true
}

function openEditDepartment(row: Department) {
  editingDepartment.value = row
  departmentForm.value = {
    name: row.name,
    code: row.code || '',
    is_active: row.is_active,
  }
  showDepartmentModal.value = true
}

async function saveDepartment() {
  try {
    if (editingDepartment.value) {
      await masterDataService.updateDepartment(editingDepartment.value.id, departmentForm.value)
    } else {
      await masterDataService.createDepartment(departmentForm.value)
    }
    showDepartmentModal.value = false
    fetchDepartments(departmentPagination.value.page)
  } catch (err: any) {
    alert(err.response?.data?.message || 'Gagal menyimpan departemen')
  }
}

async function deleteDepartment(id: string) {
  if (confirm('Hapus departemen ini?')) {
    try {
      await masterDataService.deleteDepartment(id)
      fetchDepartments(departmentPagination.value.page)
    } catch (err: any) {
      alert(err.response?.data?.message || 'Gagal menghapus departemen')
    }
  }
}

onMounted(() => {
  fetchCompanies()
  fetchDepartments()
})
</script>

<template>
  <div class="space-y-6">
    <!-- Header Page -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 bg-white p-5 rounded-xl border border-slate-200 shadow-card">
      <div>
        <div class="flex items-center gap-2">
          <h1 class="text-xl font-bold text-slate-900">Master Data Organisasi</h1>
        </div>
        <p class="text-xs text-slate-500 mt-1">
          Pengelolaan entitas perusahaan rekanan/vendor, internal PT INKA, dan divisi/departemen operasional.
        </p>
      </div>

      <!-- Tab Switcher -->
      <div class="flex items-center p-1 bg-slate-100 rounded-lg text-xs font-semibold">
        <button
          @click="activeTab = 'companies'"
          :class="[
            'px-3 py-1.5 rounded-md transition-all cursor-pointer',
            activeTab === 'companies' ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-500 hover:text-slate-800',
          ]"
        >
          Perusahaan / Vendor ({{ companyPagination.total_items }})
        </button>
        <button
          @click="activeTab = 'departments'"
          :class="[
            'px-3 py-1.5 rounded-md transition-all cursor-pointer',
            activeTab === 'departments' ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-500 hover:text-slate-800',
          ]"
        >
          Divisi & Departemen ({{ departmentPagination.total_items }})
        </button>
      </div>
    </div>

    <!-- TAB 1: COMPANIES TABLE -->
    <div v-show="activeTab === 'companies'">
      <DataTable
        :columns="companyColumns"
        :items="companies"
        :loading="loadingCompanies"
        :pagination="companyPagination"
        search-placeholder="Cari nama perusahaan, alamat..."
        @search="(q) => { companySearch = q; fetchCompanies(1) }"
        @page-change="(p) => fetchCompanies(p)"
      >
        <template #toolbar>
          <Button variant="primary" size="sm" @click="openCreateCompany">
            + Tambah Perusahaan
          </Button>
        </template>

        <template #type="{ value }">
          <span
            :class="[
              'px-2 py-0.5 rounded text-[11px] font-semibold uppercase',
              value === 'INTERNAL' ? 'bg-blue-50 text-blue-700 border border-blue-200' : 'bg-amber-50 text-amber-800 border border-amber-200',
            ]"
          >
            {{ value || 'VENDOR' }}
          </span>
        </template>

        <template #is_active="{ value }">
          <Badge :variant="value ? 'success' : 'neutral'" size="sm" dot>
            {{ value ? 'Aktif' : 'Non-Aktif' }}
          </Badge>
        </template>

        <template #actions="{ row }">
          <div class="flex items-center justify-end gap-1.5">
            <button
              @click="openEditCompany(row)"
              class="px-2 py-1 text-[11px] font-semibold rounded text-slate-600 hover:bg-slate-100 hover:text-slate-900 cursor-pointer"
            >
              Ubah
            </button>
            <button
              @click="deleteCompany(row.id)"
              class="px-2 py-1 text-[11px] font-semibold rounded text-rose-600 hover:bg-rose-50 cursor-pointer"
            >
              Hapus
            </button>
          </div>
        </template>
      </DataTable>
    </div>

    <!-- TAB 2: DEPARTMENTS TABLE -->
    <div v-show="activeTab === 'departments'">
      <DataTable
        :columns="departmentColumns"
        :items="departments"
        :loading="loadingDepartments"
        :pagination="departmentPagination"
        search-placeholder="Cari nama divisi, kode..."
        @search="(q) => { departmentSearch = q; fetchDepartments(1) }"
        @page-change="(p) => fetchDepartments(p)"
      >
        <template #toolbar>
          <Button variant="primary" size="sm" @click="openCreateDepartment">
            + Tambah Departemen
          </Button>
        </template>

        <template #code="{ value }">
          <span class="font-mono text-xs font-bold text-slate-700 px-2 py-0.5 rounded bg-slate-100 border border-slate-200">
            {{ value || '-' }}
          </span>
        </template>

        <template #is_active="{ value }">
          <Badge :variant="value ? 'success' : 'neutral'" size="sm" dot>
            {{ value ? 'Aktif' : 'Non-Aktif' }}
          </Badge>
        </template>

        <template #actions="{ row }">
          <div class="flex items-center justify-end gap-1.5">
            <button
              @click="openEditDepartment(row)"
              class="px-2 py-1 text-[11px] font-semibold rounded text-slate-600 hover:bg-slate-100 hover:text-slate-900 cursor-pointer"
            >
              Ubah
            </button>
            <button
              @click="deleteDepartment(row.id)"
              class="px-2 py-1 text-[11px] font-semibold rounded text-rose-600 hover:bg-rose-50 cursor-pointer"
            >
              Hapus
            </button>
          </div>
        </template>
      </DataTable>
    </div>

    <!-- Company Modal Form -->
    <Modal
      v-model="showCompanyModal"
      :title="editingCompany ? 'Ubah Data Perusahaan' : 'Tambah Perusahaan Baru'"
      subtitle="Pastikan data rekanan atau internal terisi lengkap"
    >
      <form @submit.prevent="saveCompany" id="companyForm" class="space-y-4">
        <div>
          <label class="block text-xs font-semibold text-slate-700 mb-1">Nama Perusahaan / Rekanan</label>
          <input
            v-model="companyForm.name"
            required
            type="text"
            placeholder="misal: PT Rekaindo Global Jasa"
            class="w-full px-3 py-2 text-xs bg-slate-50 border border-slate-300 rounded-lg text-slate-900 focus:outline-none focus:ring-1 focus:ring-inka-600 focus:bg-white"
          />
        </div>
        <div>
          <label class="block text-xs font-semibold text-slate-700 mb-1">Tipe Instansi</label>
          <select
            v-model="companyForm.type"
            class="w-full px-3 py-2 text-xs bg-slate-50 border border-slate-300 rounded-lg text-slate-900 focus:outline-none focus:ring-1 focus:ring-inka-600 focus:bg-white"
          >
            <option value="INTERNAL">INTERNAL (PT INKA Persero)</option>
            <option value="VENDOR">VENDOR (Penyedia Jasa / Barang)</option>
            <option value="KONTRAKTOR">KONTRAKTOR (Pekerjaan Khusus)</option>
            <option value="PT">PT (Perseroan Terbatas)</option>
            <option value="CV">CV</option>
          </select>
        </div>
        <div>
          <label class="block text-xs font-semibold text-slate-700 mb-1">Alamat Kantor / Workshop</label>
          <textarea
            v-model="companyForm.address"
            rows="2"
            placeholder="Alamat lengkap instansi"
            class="w-full px-3 py-2 text-xs bg-slate-50 border border-slate-300 rounded-lg text-slate-900 focus:outline-none focus:ring-1 focus:ring-inka-600 focus:bg-white"
          ></textarea>
        </div>
        <div class="grid grid-cols-2 gap-3">
          <div>
            <label class="block text-xs font-semibold text-slate-700 mb-1">Nomor Telepon</label>
            <input
              v-model="companyForm.contact_phone"
              type="text"
              placeholder="0351-xxxxxx"
              class="w-full px-3 py-2 text-xs bg-slate-50 border border-slate-300 rounded-lg text-slate-900 focus:outline-none focus:ring-1 focus:ring-inka-600 focus:bg-white"
            />
          </div>
          <div>
            <label class="block text-xs font-semibold text-slate-700 mb-1">Email Resmi</label>
            <input
              v-model="companyForm.contact_email"
              type="email"
              placeholder="kontak@perusahaan.com"
              class="w-full px-3 py-2 text-xs bg-slate-50 border border-slate-300 rounded-lg text-slate-900 focus:outline-none focus:ring-1 focus:ring-inka-600 focus:bg-white"
            />
          </div>
        </div>
        <div class="flex items-center gap-2 pt-2">
          <input
            id="companyActive"
            v-model="companyForm.is_active"
            type="checkbox"
            class="w-4 h-4 rounded text-inka-600 focus:ring-inka-500 border-slate-300"
          />
          <label for="companyActive" class="text-xs text-slate-700 font-medium">Perusahaan aktif beroperasi</label>
        </div>
      </form>

      <template #footer>
        <Button variant="outline" size="sm" @click="showCompanyModal = false">Batal</Button>
        <Button variant="primary" size="sm" type="submit" form="companyForm">Simpan Perusahaan</Button>
      </template>
    </Modal>

    <!-- Department Modal Form -->
    <Modal
      v-model="showDepartmentModal"
      :title="editingDepartment ? 'Ubah Data Departemen' : 'Tambah Departemen Baru'"
      subtitle="Input divisi atau unit operasional pabrik PT INKA"
    >
      <form @submit.prevent="saveDepartment" id="departmentForm" class="space-y-4">
        <div>
          <label class="block text-xs font-semibold text-slate-700 mb-1">Nama Divisi / Departemen</label>
          <input
            v-model="departmentForm.name"
            required
            type="text"
            placeholder="misal: Divisi Fabrikasi Kereta"
            class="w-full px-3 py-2 text-xs bg-slate-50 border border-slate-300 rounded-lg text-slate-900 focus:outline-none focus:ring-1 focus:ring-inka-600 focus:bg-white"
          />
        </div>
        <div>
          <label class="block text-xs font-semibold text-slate-700 mb-1">Kode Singkatan</label>
          <input
            v-model="departmentForm.code"
            required
            type="text"
            placeholder="misal: FAB"
            class="w-full px-3 py-2 text-xs bg-slate-50 border border-slate-300 rounded-lg text-slate-900 focus:outline-none focus:ring-1 focus:ring-inka-600 focus:bg-white uppercase font-mono"
          />
        </div>
        <div class="flex items-center gap-2 pt-2">
          <input
            id="deptActive"
            v-model="departmentForm.is_active"
            type="checkbox"
            class="w-4 h-4 rounded text-inka-600 focus:ring-inka-500 border-slate-300"
          />
          <label for="deptActive" class="text-xs text-slate-700 font-medium">Departemen aktif</label>
        </div>
      </form>

      <template #footer>
        <Button variant="outline" size="sm" @click="showDepartmentModal = false">Batal</Button>
        <Button variant="primary" size="sm" type="submit" form="departmentForm">Simpan Departemen</Button>
      </template>
    </Modal>
  </div>
</template>
