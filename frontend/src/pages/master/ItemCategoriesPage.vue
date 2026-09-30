<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { masterDataService, type ItemCategory } from '@/services/masterData'
import DataTable, { type Column } from '@/components/ui/DataTable.vue'
import Button from '@/components/ui/Button.vue'
import Badge from '@/components/ui/Badge.vue'
import Modal from '@/components/ui/Modal.vue'

const categories = ref<ItemCategory[]>([])
const loading = ref(false)
const search = ref('')

const columns: Column[] = [
  { key: 'name', label: 'Nama Kategori Barang', sortable: true },
  { key: 'description', label: 'Deskripsi & Lingkup Penggunaan' },
  { key: 'is_active', label: 'Status', align: 'center' },
  { key: 'created_at', label: 'Tanggal Dibuat', align: 'center' },
  { key: 'actions', label: 'Aksi', align: 'right' },
]

const showModal = ref(false)
const editingCategory = ref<ItemCategory | null>(null)
const form = ref({
  name: '',
  description: '',
  is_active: true,
})

async function fetchCategories() {
  loading.value = true
  try {
    const res = await masterDataService.getCategories({ search: search.value })
    if (res.data) {
      categories.value = res.data
    }
  } catch (err: any) {
    console.error('Failed to load item categories:', err)
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editingCategory.value = null
  form.value = {
    name: '',
    description: '',
    is_active: true,
  }
  showModal.value = true
}

function openEdit(row: ItemCategory) {
  editingCategory.value = row
  form.value = {
    name: row.name,
    description: row.description || '',
    is_active: row.is_active,
  }
  showModal.value = true
}

async function saveCategory() {
  try {
    if (editingCategory.value) {
      await masterDataService.updateCategory(editingCategory.value.id, form.value)
    } else {
      await masterDataService.createCategory(form.value)
    }
    showModal.value = false
    fetchCategories()
  } catch (err: any) {
    alert(err.response?.data?.message || 'Gagal menyimpan kategori barang')
  }
}

async function deleteCategory(id: string) {
  if (confirm('Hapus kategori barang ini? Pastikan tidak ada barang bawaan yang terhubung.')) {
    try {
      await masterDataService.deleteCategory(id)
      fetchCategories()
    } catch (err: any) {
      alert(err.response?.data?.message || 'Gagal menghapus kategori barang')
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

onMounted(() => {
  fetchCategories()
})
</script>

<template>
  <div class="space-y-6">
    <!-- Header Page -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 bg-white p-5 rounded-xl border border-slate-200 shadow-card">
      <div>
        <div class="flex items-center gap-2">
          <h1 class="text-xl font-bold text-slate-900">Kategori Alat Kerja & Peralatan</h1>
        </div>
        <p class="text-xs text-slate-500 mt-1">
          Klasifikasi standar barang bawaan pekerja (alat kerja teknisi, IT/elektronik, safety/APD, alat ukur presisi).
        </p>
      </div>

      <div class="flex items-center gap-3">
        <Button variant="primary" size="sm" @click="openCreate">
          + Tambah Kategori
        </Button>
      </div>
    </div>

    <!-- Data Table -->
    <DataTable
      :columns="columns"
      :items="categories"
      :loading="loading"
      search-placeholder="Cari kategori barang..."
      @search="(q) => { search = q; fetchCategories() }"
    >
      <template #name="{ row, value }">
        <div class="flex items-center gap-2.5">
          <div class="w-8 h-8 rounded-lg bg-inka-50 text-inka-700 border border-inka-100 flex items-center justify-center shrink-0">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
            </svg>
          </div>
          <div>
            <span class="font-bold text-slate-900 text-xs">{{ value }}</span>
            <span class="block text-[10px] text-slate-400 font-mono">ID: {{ row.id.substring(0, 8) }}...</span>
          </div>
        </div>
      </template>

      <template #description="{ value }">
        <span class="text-xs text-slate-600 line-clamp-1 max-w-md">
          {{ value || 'Tidak ada keterangan tambahan.' }}
        </span>
      </template>

      <template #is_active="{ value }">
        <Badge :variant="value ? 'success' : 'neutral'" size="sm" dot>
          {{ value ? 'Aktif Digunakan' : 'Dinonaktifkan' }}
        </Badge>
      </template>

      <template #created_at="{ value }">
        <span class="text-xs text-slate-500 font-mono">{{ formatDate(value) }}</span>
      </template>

      <template #actions="{ row }">
        <div class="flex items-center justify-end gap-1.5">
          <button
            @click="openEdit(row)"
            class="px-2.5 py-1 text-[11px] font-semibold rounded text-slate-600 hover:bg-slate-100 hover:text-slate-900 cursor-pointer transition-colors"
          >
            Ubah
          </button>
          <button
            @click="deleteCategory(row.id)"
            class="px-2.5 py-1 text-[11px] font-semibold rounded text-rose-600 hover:bg-rose-50 cursor-pointer transition-colors"
          >
            Hapus
          </button>
        </div>
      </template>
    </DataTable>

    <!-- Modal Form -->
    <Modal
      v-model="showModal"
      :title="editingCategory ? 'Ubah Kategori Barang' : 'Tambah Kategori Barang'"
      subtitle="Kategori ini akan muncul pada opsi klasifikasi barang SIMAK dan profil pekerja"
    >
      <form @submit.prevent="saveCategory" id="catForm" class="space-y-4">
        <div>
          <label class="block text-xs font-semibold text-slate-700 mb-1">Nama Kategori Barang</label>
          <input
            v-model="form.name"
            required
            type="text"
            placeholder="misal: Alat Ukur & Kalibrasi"
            class="w-full px-3 py-2 text-xs bg-slate-50 border border-slate-300 rounded-lg text-slate-900 focus:outline-none focus:ring-1 focus:ring-inka-600 focus:bg-white"
          />
        </div>
        <div>
          <label class="block text-xs font-semibold text-slate-700 mb-1">Deskripsi & Ruang Lingkup</label>
          <textarea
            v-model="form.description"
            rows="3"
            placeholder="Keterangan jenis barang atau pedoman keselamatan untuk kategori ini..."
            class="w-full px-3 py-2 text-xs bg-slate-50 border border-slate-300 rounded-lg text-slate-900 focus:outline-none focus:ring-1 focus:ring-inka-600 focus:bg-white"
          ></textarea>
        </div>
        <div class="flex items-center gap-2 pt-2">
          <input
            id="catActive"
            v-model="form.is_active"
            type="checkbox"
            class="w-4 h-4 rounded text-inka-600 focus:ring-inka-500 border-slate-300"
          />
          <label for="catActive" class="text-xs text-slate-700 font-medium">Kategori aktif dan dapat dipilih oleh pemohon</label>
        </div>
      </form>

      <template #footer>
        <Button variant="outline" size="sm" @click="showModal = false">Batal</Button>
        <Button variant="primary" size="sm" type="submit" form="catForm">Simpan Kategori</Button>
      </template>
    </Modal>
  </div>
</template>
