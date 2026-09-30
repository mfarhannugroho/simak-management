<script setup lang="ts">
import { ref, onMounted } from 'vue'
import {
  masterDataService,
  type WorkerRegisteredItem,
  type ItemCategory,
} from '@/services/masterData'
import Button from '@/components/ui/Button.vue'
import Badge from '@/components/ui/Badge.vue'
import Modal from '@/components/ui/Modal.vue'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()

const items = ref<WorkerRegisteredItem[]>([])
const categories = ref<ItemCategory[]>([])
const loading = ref(false)
const search = ref('')

const showModal = ref(false)
const editingItem = ref<WorkerRegisteredItem | null>(null)
const form = ref({
  item_name: '',
  category_id: '',
  brand: '',
  model: '',
  serial_number: '',
  quantity: 1,
  unit: 'Unit',
  description: '',
  photo_url: '',
})

async function fetchMyItems() {
  loading.value = true
  try {
    const res = await masterDataService.getMyRegisteredItems()
    if (res.data) {
      items.value = res.data
    }
  } catch (err: any) {
    console.error('Failed to load registered items:', err)
  } finally {
    loading.value = false
  }
}

async function loadCategories() {
  try {
    const res = await masterDataService.getCategories({ active_only: true })
    if (res.data) {
      categories.value = res.data
    }
  } catch (err) {
    console.error('Failed to load categories:', err)
  }
}

function openCreate() {
  editingItem.value = null
  form.value = {
    item_name: '',
    category_id: categories.value[0]?.id || '',
    brand: '',
    model: '',
    serial_number: '',
    quantity: 1,
    unit: 'Unit',
    description: '',
    photo_url: '',
  }
  showModal.value = true
}

function openEdit(item: WorkerRegisteredItem) {
  editingItem.value = item
  form.value = {
    item_name: item.item_name,
    category_id: item.category_id || '',
    brand: item.brand || '',
    model: item.model || '',
    serial_number: item.serial_number || '',
    quantity: item.quantity || 1,
    unit: item.unit || 'Unit',
    description: item.description || '',
    photo_url: item.photo_url || '',
  }
  showModal.value = true
}

async function saveItem() {
  try {
    if (editingItem.value) {
      await masterDataService.updateRegisteredItem(editingItem.value.id, form.value)
    } else {
      await masterDataService.createRegisteredItem(form.value)
    }
    showModal.value = false
    fetchMyItems()
  } catch (err: any) {
    alert(err.response?.data?.message || 'Gagal menyimpan barang rutin')
  }
}

async function deleteItem(id: string) {
  if (confirm('Hapus barang bawaan rutin ini dari profil Anda?')) {
    try {
      await masterDataService.deleteRegisteredItem(id)
      fetchMyItems()
    } catch (err: any) {
      alert(err.response?.data?.message || 'Gagal menghapus barang')
    }
  }
}

const filteredItems = () => {
  if (!search.value) return items.value
  const q = search.value.toLowerCase()
  return items.value.filter(
    (i) =>
      i.item_name.toLowerCase().includes(q) ||
      (i.brand && i.brand.toLowerCase().includes(q)) ||
      (i.serial_number && i.serial_number.toLowerCase().includes(q)) ||
      (i.category && i.category.name.toLowerCase().includes(q)),
  )
}

onMounted(() => {
  fetchMyItems()
  loadCategories()
})
</script>

<template>
  <div class="space-y-6">
    <!-- Header Page -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 bg-white p-5 rounded-xl border border-slate-200 shadow-card">
      <div>
        <div class="flex items-center gap-2">
          <h1 class="text-xl font-bold text-slate-900">Barang Bawaan Rutin Saya</h1>
        </div>
        <p class="text-xs text-slate-500 mt-1">
          Daftar peralatan kerja & perangkat rutin yang Anda bawa ke lingkungan workshop/pabrik PT INKA.
          Barang di sini dapat langsung dipilih saat mengajukan SIMAK baru tanpa perlu mengisi ulang.
        </p>
      </div>

      <div class="flex items-center gap-3">
        <Button variant="primary" size="sm" @click="openCreate">
          + Daftarkan Barang Baru
        </Button>
      </div>
    </div>

    <!-- Worker Profile Summary Box -->
    <div class="bg-gradient-to-r from-slate-900 to-slate-800 text-white rounded-xl p-5 shadow-sm flex flex-col md:flex-row md:items-center justify-between gap-4">
      <div class="flex items-center gap-4">
        <div class="w-12 h-12 rounded-xl bg-white/10 border border-white/20 flex items-center justify-center font-bold text-lg text-white">
          {{ (authStore.user?.full_name || 'P').charAt(0) }}
        </div>
        <div>
          <h2 class="text-sm font-bold text-white">{{ authStore.user?.full_name }}</h2>
          <div class="flex items-center gap-3 text-xs text-slate-300 mt-0.5">
            <span>NIP / ID: <strong class="text-white font-mono">{{ authStore.user?.worker_id || '-' }}</strong></span>
            <span>·</span>
            <span>Instansi: <strong class="text-white">{{ authStore.user?.company || 'PT INKA (Persero)' }}</strong></span>
          </div>
        </div>
      </div>

      <div class="flex items-center gap-4 bg-white/5 border border-white/10 rounded-lg px-4 py-2.5">
        <div>
          <p class="text-[11px] text-slate-400 uppercase tracking-wider font-semibold">Total Terdaftar</p>
          <p class="text-lg font-bold text-white">{{ items.length }} <span class="text-xs font-normal text-slate-300">Peralatan</span></p>
        </div>
      </div>
    </div>

    <!-- Search bar -->
    <div class="flex items-center gap-3">
      <div class="relative flex-1 max-w-md">
        <span class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none text-slate-400">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
        </span>
        <input
          v-model="search"
          type="text"
          placeholder="Cari berdasarkan nama alat, merk, atau nomor seri..."
          class="w-full pl-9 pr-4 py-2 text-xs bg-white border border-slate-200 rounded-lg text-slate-800 placeholder-slate-400 focus:outline-none focus:ring-1 focus:ring-inka-600 focus:border-inka-600"
        />
      </div>
    </div>

    <!-- Empty State -->
    <div
      v-if="!loading && filteredItems().length === 0"
      class="bg-white border border-slate-200 rounded-xl p-12 text-center"
    >
      <div class="w-12 h-12 rounded-full bg-slate-100 text-slate-400 flex items-center justify-center mx-auto mb-3">
        <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
        </svg>
      </div>
      <h3 class="text-sm font-bold text-slate-800">Belum Ada Barang Bawaan Rutin</h3>
      <p class="text-xs text-slate-500 mt-1 max-w-sm mx-auto">
        Daftarkan peralatan kerja (laptop kantor, toolset teknisi, alat ukur) yang sering Anda bawa agar permohonan SIMAK lebih cepat.
      </p>
      <Button variant="primary" size="sm" class="mt-4" @click="openCreate">
        + Daftarkan Barang Sekarang
      </Button>
    </div>

    <!-- Cards Grid -->
    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      <div
        v-for="item in filteredItems()"
        :key="item.id"
        class="bg-white rounded-xl border border-slate-200 p-4 shadow-sm hover:border-slate-300 transition-all flex flex-col justify-between"
      >
        <div>
          <!-- Category badge & qty -->
          <div class="flex items-center justify-between gap-2 mb-2.5">
            <span class="px-2 py-0.5 rounded-full text-[10px] font-bold bg-inka-50 text-inka-700 border border-inka-100 uppercase tracking-wide">
              {{ item.category?.name || 'Umum' }}
            </span>
            <span class="text-xs font-bold text-slate-800 bg-slate-100 px-2 py-0.5 rounded border border-slate-200">
              {{ item.quantity }} {{ item.unit }}
            </span>
          </div>

          <!-- Item Name & Brand -->
          <h3 class="text-sm font-bold text-slate-900 leading-snug">{{ item.item_name }}</h3>
          <p class="text-xs text-slate-500 mt-0.5">
            {{ item.brand || 'Tanpa Merk' }} {{ item.model ? `— ${item.model}` : '' }}
          </p>

          <!-- Serial Number -->
          <div class="mt-3 py-1.5 px-2.5 bg-slate-50 rounded-lg border border-slate-100 flex items-center justify-between">
            <span class="text-[11px] font-medium text-slate-500">No. Seri (S/N):</span>
            <span class="font-mono text-xs font-bold text-slate-800">{{ item.serial_number || '-' }}</span>
          </div>

          <!-- Description -->
          <p v-if="item.description" class="text-xs text-slate-600 mt-2 line-clamp-2">
            {{ item.description }}
          </p>
        </div>

        <!-- Action footer -->
        <div class="flex items-center justify-end gap-2 pt-3 mt-4 border-t border-slate-100">
          <button
            @click="openEdit(item)"
            class="px-2.5 py-1 text-xs font-semibold rounded text-slate-600 hover:bg-slate-100 hover:text-slate-900 cursor-pointer transition-colors"
          >
            Ubah
          </button>
          <button
            @click="deleteItem(item.id)"
            class="px-2.5 py-1 text-xs font-semibold rounded text-rose-600 hover:bg-rose-50 cursor-pointer transition-colors"
          >
            Hapus
          </button>
        </div>
      </div>
    </div>

    <!-- Modal Form Registered Item -->
    <Modal
      v-model="showModal"
      :title="editingItem ? 'Ubah Barang Bawaan Rutin' : 'Daftarkan Barang Bawaan Rutin'"
      subtitle="Barang ini akan tersimpan di profil Anda dan otomatis dapat dipilih pada formulir SIMAK"
    >
      <form @submit.prevent="saveItem" id="itemForm" class="space-y-4">
        <div>
          <label class="block text-xs font-semibold text-slate-700 mb-1">Nama Barang / Peralatan Kerja *</label>
          <input
            v-model="form.item_name"
            required
            type="text"
            placeholder="misal: Laptop Lenovo ThinkPad / Multimeter Digital"
            class="w-full px-3 py-2 text-xs bg-slate-50 border border-slate-300 rounded-lg text-slate-900 focus:outline-none focus:ring-1 focus:ring-inka-600 focus:bg-white"
          />
        </div>

        <div>
          <label class="block text-xs font-semibold text-slate-700 mb-1">Kategori Peralatan *</label>
          <select
            v-model="form.category_id"
            required
            class="w-full px-3 py-2 text-xs bg-slate-50 border border-slate-300 rounded-lg text-slate-900 focus:outline-none focus:ring-1 focus:ring-inka-600 focus:bg-white"
          >
            <option value="" disabled>-- Pilih Kategori Peralatan --</option>
            <option v-for="cat in categories" :key="cat.id" :value="cat.id">
              {{ cat.name }}
            </option>
          </select>
        </div>

        <div class="grid grid-cols-2 gap-3">
          <div>
            <label class="block text-xs font-semibold text-slate-700 mb-1">Merk / Brand</label>
            <input
              v-model="form.brand"
              type="text"
              placeholder="misal: Fluke / Makita / Dell"
              class="w-full px-3 py-2 text-xs bg-slate-50 border border-slate-300 rounded-lg text-slate-900 focus:outline-none focus:ring-1 focus:ring-inka-600 focus:bg-white"
            />
          </div>
          <div>
            <label class="block text-xs font-semibold text-slate-700 mb-1">Model / Tipe</label>
            <input
              v-model="form.model"
              type="text"
              placeholder="misal: T14 Gen 3 / 115 True-RMS"
              class="w-full px-3 py-2 text-xs bg-slate-50 border border-slate-300 rounded-lg text-slate-900 focus:outline-none focus:ring-1 focus:ring-inka-600 focus:bg-white"
            />
          </div>
        </div>

        <div class="grid grid-cols-3 gap-3">
          <div class="col-span-1">
            <label class="block text-xs font-semibold text-slate-700 mb-1">Jumlah</label>
            <input
              v-model.number="form.quantity"
              type="number"
              min="1"
              required
              class="w-full px-3 py-2 text-xs bg-slate-50 border border-slate-300 rounded-lg text-slate-900 focus:outline-none focus:ring-1 focus:ring-inka-600 focus:bg-white"
            />
          </div>
          <div class="col-span-1">
            <label class="block text-xs font-semibold text-slate-700 mb-1">Satuan</label>
            <select
              v-model="form.unit"
              required
              class="w-full px-3 py-2 text-xs bg-slate-50 border border-slate-300 rounded-lg text-slate-900 focus:outline-none focus:ring-1 focus:ring-inka-600 focus:bg-white"
            >
              <option value="Unit">Unit</option>
              <option value="Set">Set</option>
              <option value="Pcs">Pcs</option>
              <option value="Buah">Buah</option>
              <option value="Box">Box / Kotak</option>
            </select>
          </div>
          <div class="col-span-1">
            <label class="block text-xs font-semibold text-slate-700 mb-1">Nomor Seri (S/N) *</label>
            <input
              v-model="form.serial_number"
              type="text"
              required
              placeholder="S/N fisik"
              class="w-full px-3 py-2 text-xs bg-slate-50 border border-slate-300 rounded-lg text-slate-900 focus:outline-none focus:ring-1 focus:ring-inka-600 focus:bg-white font-mono"
            />
          </div>
        </div>

        <div>
          <label class="block text-xs font-semibold text-slate-700 mb-1">Keterangan / Kondisi Alat</label>
          <textarea
            v-model="form.description"
            rows="2"
            placeholder="Kondisi fisik barang, kelengkapan adaptor kabel, tas bawaan, dsb..."
            class="w-full px-3 py-2 text-xs bg-slate-50 border border-slate-300 rounded-lg text-slate-900 focus:outline-none focus:ring-1 focus:ring-inka-600 focus:bg-white"
          ></textarea>
        </div>
      </form>

      <template #footer>
        <Button variant="outline" size="sm" @click="showModal = false">Batal</Button>
        <Button variant="primary" size="sm" type="submit" form="itemForm">Simpan Barang Rutin</Button>
      </template>
    </Modal>
  </div>
</template>
