<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import {
  simakService,
  type Simak,
  type SimakItem,
  type SimakAttachment,
} from '@/services/simak'
import {
  masterDataService,
  type ItemCategory,
  type WorkerRegisteredItem,
} from '@/services/masterData'
import Button from '@/components/ui/Button.vue'
import Badge from '@/components/ui/Badge.vue'
import Modal from '@/components/ui/Modal.vue'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const isEditMode = computed(() => !!route.params.id)
const pageTitle = computed(() =>
  isEditMode.value ? 'Ubah Permohonan SIMAK' : 'Formulir Permohonan SIMAK Baru'
)

// Main Form State
const purpose = ref('')
const direction = ref<'MASUK' | 'KELUAR' | 'MASUK_KELUAR'>('MASUK')
const destinationArea = ref('Workshop Perakitan Bogie PT INKA')
const startDate = ref(new Date().toISOString().substring(0, 10))
const endDate = ref(
  new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString().substring(0, 10)
)
const notes = ref('')

const items = ref<SimakItem[]>([])
const attachments = ref<SimakAttachment[]>([])

const categories = ref<ItemCategory[]>([])
const routineItems = ref<WorkerRegisteredItem[]>([])
const showImportModal = ref(false)
const selectedRoutineItemIds = ref<string[]>([])

const saving = ref(false)
const loadingInitial = ref(false)
const validationError = ref('')

// Pre-defined Workshop Areas at PT INKA
const standardWorkshopAreas = [
  'Workshop Fabrikasi Kereta (Carbody & Underframe)',
  'Workshop Perakitan Bogie PT INKA',
  'Workshop Finishing & Pengecatan Eksterior',
  'Workshop Interior & Sistem Elektrikal Kereta',
  'Area Pengujian Lintasan & Final Inspection Track',
  'Gudang Logistik & Penerimaan Komponen',
  'Workshop Pemeliharaan Fasilitas Pabrik (Utility)',
]

async function loadMasterData() {
  try {
    const [catRes, routineRes] = await Promise.all([
      masterDataService.getCategories({ active_only: true }),
      masterDataService.getMyRegisteredItems(),
    ])
    if (catRes.data) categories.value = catRes.data
    if (routineRes.data) routineItems.value = routineRes.data
  } catch (err) {
    console.error('Failed to load form master data:', err)
  }
}

async function loadExistingSimak() {
  if (!isEditMode.value) {
    // Add 1 default empty item row for convenience
    addNewItemRow()
    return
  }

  loadingInitial.value = true
  try {
    const res = await simakService.getSimak(route.params.id as string)
    if (res.data) {
      const data = res.data
      purpose.value = data.purpose
      direction.value = data.direction
      destinationArea.value = data.destination_area
      startDate.value = data.start_date.substring(0, 10)
      endDate.value = data.end_date.substring(0, 10)
      notes.value = data.notes || ''
      items.value = data.items || []
      attachments.value = data.attachments || []

      if (items.value.length === 0) {
        addNewItemRow()
      }
    }
  } catch (err: any) {
    alert(err.response?.data?.message || 'Gagal memuat data SIMAK')
    router.push('/simak')
  } finally {
    loadingInitial.value = false
  }
}

function addNewItemRow() {
  items.value.push({
    item_name: '',
    category_id: categories.value[0]?.id || '',
    brand: '',
    model: '',
    serial_number: '',
    size_spec: '',
    quantity: 1,
    unit: 'Unit',
    description: '',
  })
}

function removeItemRow(index: number) {
  if (items.value.length === 1) {
    if (!confirm('Hapus baris barang ini? Permohonan SIMAK wajib memiliki minimal 1 barang.')) {
      return
    }
  }
  items.value.splice(index, 1)
}

// Routine items import picker
function openImportRoutineItemsModal() {
  selectedRoutineItemIds.value = []
  showImportModal.value = true
}

function toggleRoutineSelection(id: string) {
  const idx = selectedRoutineItemIds.value.indexOf(id)
  if (idx > -1) {
    selectedRoutineItemIds.value.splice(idx, 1)
  } else {
    selectedRoutineItemIds.value.push(id)
  }
}

function selectAllRoutineItems() {
  if (selectedRoutineItemIds.value.length === routineItems.value.length) {
    selectedRoutineItemIds.value = []
  } else {
    selectedRoutineItemIds.value = routineItems.value.map((i) => i.id)
  }
}

function applyImportRoutineItems() {
  const chosen = routineItems.value.filter((i) =>
    selectedRoutineItemIds.value.includes(i.id)
  )

  // Remove completely blank initial row if exists
  if (
    items.value.length === 1 &&
    !items.value[0].item_name &&
    !items.value[0].serial_number
  ) {
    items.value = []
  }

  for (const item of chosen) {
    items.value.push({
      item_name: item.item_name,
      category_id: item.category_id || '',
      brand: item.brand || '',
      model: item.model || '',
      serial_number: item.serial_number || '',
      size_spec: '',
      quantity: item.quantity || 1,
      unit: item.unit || 'Unit',
      description: item.description || '',
    })
  }

  showImportModal.value = false
}

// Add attachment simulator
const newAttachmentName = ref('')
const newAttachmentType = ref('SURAT_TUGAS')

function addSampleAttachment() {
  if (!newAttachmentName.value) {
    alert('Masukkan nama file dokumen')
    return
  }
  attachments.value.push({
    file_name: newAttachmentName.value,
    file_url: `/uploads/documents/${encodeURIComponent(newAttachmentName.value)}`,
    file_type: 'application/pdf',
    file_size: 154200,
    attachment_type: newAttachmentType.value,
  })
  newAttachmentName.value = ''
}

function removeAttachment(index: number) {
  attachments.value.splice(index, 1)
}

// Form Submission
async function handleSave(submitDirectly = false) {
  validationError.value = ''

  if (!purpose.value.trim()) {
    validationError.value = 'Keperluan atau tujuan izin membawa alat wajib diisi.'
    return
  }

  if (!destinationArea.value.trim()) {
    validationError.value = 'Area workshop tujuan di pabrik PT INKA wajib ditentukan.'
    return
  }

  if (new Date(endDate.value) < new Date(startDate.value)) {
    validationError.value = 'Tanggal selesai tidak boleh lebih awal dari tanggal mulai.'
    return
  }

  const validItems = items.value.filter((i) => i.item_name.trim().length > 0)
  if (submitDirectly && validItems.length === 0) {
    validationError.value = 'Permohonan SIMAK wajib memiliki minimal 1 (satu) barang bawaan.'
    return
  }

  saving.value = true
  try {
    const payload = {
      purpose: purpose.value,
      direction: direction.value,
      destination_area: destinationArea.value,
      start_date: startDate.value,
      end_date: endDate.value,
      notes: notes.value,
      submit_directly: submitDirectly,
      items: validItems,
      attachments: attachments.value,
    }

    if (isEditMode.value) {
      await simakService.updateSimak(route.params.id as string, payload)
      if (submitDirectly) {
        await simakService.submitSimak(route.params.id as string)
      }
      alert('Permohonan SIMAK berhasil diperbarui.')
      router.push(`/simak/${route.params.id}`)
    } else {
      const res = await simakService.createSimak(payload)
      if (res.data) {
        alert(
          submitDirectly
            ? 'Permohonan SIMAK berhasil dibuat dan diajukan untuk ditinjau.'
            : 'Draft permohonan SIMAK berhasil disimpan.'
        )
        router.push(`/simak/${res.data.id}`)
      }
    }
  } catch (err: any) {
    validationError.value = err.response?.data?.message || 'Gagal menyimpan SIMAK'
  } finally {
    saving.value = false
  }
}

onMounted(async () => {
  await loadMasterData()
  await loadExistingSimak()
})
</script>

<template>
  <div class="space-y-6 max-w-5xl mx-auto pb-16">
    <!-- Breadcrumb & Header Title -->
    <div>
      <div class="flex items-center gap-2 text-xs text-slate-500 mb-2">
        <router-link to="/" class="hover:text-slate-800">Dashboard</router-link>
        <span>/</span>
        <router-link to="/simak" class="hover:text-slate-800">Permohonan SIMAK</router-link>
        <span>/</span>
        <span class="text-slate-800 font-semibold">{{ isEditMode ? 'Ubah' : 'Baru' }}</span>
      </div>

      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 bg-white p-5 rounded-xl border border-slate-200 shadow-card">
        <div>
          <div class="flex items-center gap-2">
            <h1 class="text-xl font-bold text-slate-900">{{ pageTitle }}</h1>
            <Badge variant="inka" size="sm">Formulir Resmi</Badge>
          </div>
          <p class="text-xs text-slate-500 mt-1">
            Lengkapi data izin membawa peralatan kerja, laptop, alat ukur, atau perkakas teknisi masuk/keluar area workshop PT INKA.
          </p>
        </div>

        <router-link to="/simak">
          <Button variant="outline" size="sm">
            &larr; Kembali ke Daftar
          </Button>
        </router-link>
      </div>
    </div>

    <!-- Error Alert Banner -->
    <div
      v-if="validationError"
      class="p-4 bg-rose-50 border-l-4 border-rose-500 rounded-r-xl text-rose-800 text-xs flex items-center justify-between shadow-sm"
    >
      <div class="flex items-center gap-2">
        <svg class="w-5 h-5 text-rose-600 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <span class="font-medium">{{ validationError }}</span>
      </div>
      <button @click="validationError = ''" class="text-rose-500 hover:text-rose-800 font-bold">&times;</button>
    </div>

    <!-- SECTION 1: APPLICANT PROFILE SUMMARY (READ-ONLY) -->
    <div class="bg-white rounded-xl border border-slate-200 shadow-card p-5 space-y-4">
      <div class="flex items-center justify-between pb-3 border-b border-slate-100">
        <div class="flex items-center gap-2.5">
          <div class="w-7 h-7 rounded-lg bg-inka-50 text-inka-700 flex items-center justify-center font-bold text-xs">
            1
          </div>
          <div>
            <h2 class="text-sm font-bold text-slate-900">Identitas Pemohon (Applicant Profile)</h2>
            <p class="text-[11px] text-slate-500">Data otomatis ditarik dari akun dan profil pekerja Anda</p>
          </div>
        </div>
        <span class="text-[11px] font-semibold text-emerald-700 bg-emerald-50 px-2 py-0.5 rounded border border-emerald-200">
          Akun Terverifikasi
        </span>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-3 gap-4 text-xs">
        <div class="p-3 rounded-lg bg-slate-50 border border-slate-100 space-y-1">
          <span class="text-[10px] text-slate-400 font-semibold uppercase">Nama Pekerja</span>
          <p class="font-bold text-slate-900 text-xs">{{ authStore.user?.full_name || 'Ahmad Wahyudi' }}</p>
          <p class="text-[11px] text-slate-500 font-mono">NIP / ID: {{ authStore.user?.worker_id || 'REK-FAB-104' }}</p>
        </div>

        <div class="p-3 rounded-lg bg-slate-50 border border-slate-100 space-y-1">
          <span class="text-[10px] text-slate-400 font-semibold uppercase">Perusahaan / Instansi</span>
          <p class="font-bold text-slate-900 text-xs">{{ authStore.user?.company || 'PT Rekaindo Global Jasa (Vendor)' }}</p>
          <p class="text-[11px] text-slate-500">{{ authStore.user?.position || 'Teknisi Perakitan Bogie' }}</p>
        </div>

        <div class="p-3 rounded-lg bg-slate-50 border border-slate-100 space-y-1">
          <span class="text-[10px] text-slate-400 font-semibold uppercase">Divisi & Kontak</span>
          <p class="font-bold text-slate-900 text-xs">{{ authStore.user?.department || 'Divisi Rekayasa & Fabrikasi' }}</p>
          <p class="text-[11px] text-slate-500">{{ authStore.user?.phone || '0821-9876-5432' }}</p>
        </div>
      </div>
    </div>

    <!-- SECTION 2: PERMIT DETAILS -->
    <div class="bg-white rounded-xl border border-slate-200 shadow-card p-5 space-y-5">
      <div class="flex items-center gap-2.5 pb-3 border-b border-slate-100">
        <div class="w-7 h-7 rounded-lg bg-inka-50 text-inka-700 flex items-center justify-center font-bold text-xs">
          2
        </div>
        <div>
          <h2 class="text-sm font-bold text-slate-900">Rincian Perizinan & Area Workshop</h2>
          <p class="text-[11px] text-slate-500">Tentukan tujuan pekerjaan dan area pabrik PT INKA yang dituju</p>
        </div>
      </div>

      <div class="space-y-4">
        <!-- Purpose / Keperluan -->
        <div>
          <label class="block text-xs font-semibold text-slate-800 mb-1">
            Keperluan / Tujuan Membawa Barang Kerja *
          </label>
          <textarea
            v-model="purpose"
            required
            rows="2"
            placeholder="misal: Melakukan pengujian sistem pengereman pneumatik gerbong K3 dan kalibrasi sensor bogie di workshop INKA"
            class="w-full px-3 py-2 text-xs bg-slate-50 border border-slate-300 rounded-lg text-slate-900 placeholder-slate-400 focus:outline-none focus:ring-1 focus:ring-inka-600 focus:bg-white"
          ></textarea>
        </div>

        <!-- Direction (Arah Aliran Barang) -->
        <div>
          <label class="block text-xs font-semibold text-slate-800 mb-2">
            Arah Aliran Pergerakan Barang *
          </label>
          <div class="grid grid-cols-1 sm:grid-cols-3 gap-3">
            <label
              :class="[
                'flex items-center gap-3 p-3 rounded-xl border cursor-pointer transition-all',
                direction === 'MASUK'
                  ? 'border-inka-600 bg-inka-50/40 ring-1 ring-inka-600 text-slate-900'
                  : 'border-slate-200 bg-white hover:bg-slate-50 text-slate-700',
              ]"
            >
              <input
                type="radio"
                v-model="direction"
                value="MASUK"
                class="text-inka-600 focus:ring-inka-500"
              />
              <div>
                <p class="text-xs font-bold">Membawa Masuk</p>
                <p class="text-[10px] text-slate-500">Barang dari luar dibawa masuk pabrik</p>
              </div>
            </label>

            <label
              :class="[
                'flex items-center gap-3 p-3 rounded-xl border cursor-pointer transition-all',
                direction === 'KELUAR'
                  ? 'border-inka-600 bg-inka-50/40 ring-1 ring-inka-600 text-slate-900'
                  : 'border-slate-200 bg-white hover:bg-slate-50 text-slate-700',
              ]"
            >
              <input
                type="radio"
                v-model="direction"
                value="KELUAR"
                class="text-inka-600 focus:ring-inka-500"
              />
              <div>
                <p class="text-xs font-bold">Membawa Keluar</p>
                <p class="text-[10px] text-slate-500">Peralatan dibawa keluar dari pabrik</p>
              </div>
            </label>

            <label
              :class="[
                'flex items-center gap-3 p-3 rounded-xl border cursor-pointer transition-all',
                direction === 'MASUK_KELUAR'
                  ? 'border-inka-600 bg-inka-50/40 ring-1 ring-inka-600 text-slate-900'
                  : 'border-slate-200 bg-white hover:bg-slate-50 text-slate-700',
              ]"
            >
              <input
                type="radio"
                v-model="direction"
                value="MASUK_KELUAR"
                class="text-inka-600 focus:ring-inka-500"
              />
              <div>
                <p class="text-xs font-bold">Masuk & Keluar</p>
                <p class="text-[10px] text-slate-500">Pekerjaan harian / proyek berulang</p>
              </div>
            </label>
          </div>
        </div>

        <!-- Destination Area & Validity Dates -->
        <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div class="md:col-span-1">
            <label class="block text-xs font-semibold text-slate-800 mb-1">
              Area Workshop Tujuan di PT INKA *
            </label>
            <input
              v-model="destinationArea"
              list="areaSuggestions"
              required
              type="text"
              placeholder="Pilih atau ketik workshop..."
              class="w-full px-3 py-2 text-xs bg-slate-50 border border-slate-300 rounded-lg text-slate-900 focus:outline-none focus:ring-1 focus:ring-inka-600 focus:bg-white"
            />
            <datalist id="areaSuggestions">
              <option v-for="a in standardWorkshopAreas" :key="a" :value="a" />
            </datalist>
          </div>

          <div>
            <label class="block text-xs font-semibold text-slate-800 mb-1">
              Tanggal Mulai Berlaku *
            </label>
            <input
              v-model="startDate"
              required
              type="date"
              class="w-full px-3 py-2 text-xs bg-slate-50 border border-slate-300 rounded-lg text-slate-900 focus:outline-none focus:ring-1 focus:ring-inka-600 focus:bg-white font-mono"
            />
          </div>

          <div>
            <label class="block text-xs font-semibold text-slate-800 mb-1">
              Tanggal Selesai Berlaku *
            </label>
            <input
              v-model="endDate"
              required
              type="date"
              class="w-full px-3 py-2 text-xs bg-slate-50 border border-slate-300 rounded-lg text-slate-900 focus:outline-none focus:ring-1 focus:ring-inka-600 focus:bg-white font-mono"
            />
          </div>
        </div>

        <!-- Notes -->
        <div>
          <label class="block text-xs font-semibold text-slate-800 mb-1">Catatan Tambahan (Opsional)</label>
          <input
            v-model="notes"
            type="text"
            placeholder="misal: Didampingi pengawas mutu pabrik INKA Bpk. Hendro"
            class="w-full px-3 py-2 text-xs bg-slate-50 border border-slate-300 rounded-lg text-slate-900 focus:outline-none focus:ring-1 focus:ring-inka-600 focus:bg-white"
          />
        </div>
      </div>
    </div>

    <!-- SECTION 3: ITEMS TABLE -->
    <div class="bg-white rounded-xl border border-slate-200 shadow-card p-5 space-y-4">
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pb-3 border-b border-slate-100">
        <div class="flex items-center gap-2.5">
          <div class="w-7 h-7 rounded-lg bg-inka-50 text-inka-700 flex items-center justify-center font-bold text-xs">
            3
          </div>
          <div>
            <h2 class="text-sm font-bold text-slate-900">Rincian Barang Bawaan & Peralatan Kerja</h2>
            <p class="text-[11px] text-slate-500">Cantumkan nomor seri (S/N) dan jumlah fisik secara akurat untuk verifikasi pos gerbang</p>
          </div>
        </div>

        <!-- Action buttons -->
        <div class="flex items-center gap-2">
          <!-- Button Import Routine Items -->
          <Button
            v-if="routineItems.length > 0"
            variant="outline"
            size="sm"
            type="button"
            @click="openImportRoutineItemsModal"
          >
            <svg class="w-3.5 h-3.5 mr-1.5 text-inka-700" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7v8a2 2 0 002 2h6M8 7V5a2 2 0 012-2h4.586a1 1 0 01.707.293l4.414 4.414a1 1 0 01.293.707V15a2 2 0 01-2 2h-2M8 7H6a2 2 0 00-2 2v10a2 2 0 002 2h8a2 2 0 002-2v-2" />
            </svg>
            Impor dari Barang Rutin Saya
          </Button>

          <Button
            variant="primary"
            size="sm"
            type="button"
            @click="addNewItemRow"
          >
            + Tambah Baris Manual
          </Button>
        </div>
      </div>

      <!-- Table of items -->
      <div class="overflow-x-auto border border-slate-200 rounded-lg">
        <table class="w-full text-left text-xs border-collapse">
          <thead>
            <tr class="bg-slate-50 border-b border-slate-200 text-slate-600 font-bold uppercase tracking-wider text-[11px]">
              <th class="py-2.5 px-3 w-10 text-center">No</th>
              <th class="py-2.5 px-3 min-w-[180px]">Nama Barang / Alat *</th>
              <th class="py-2.5 px-3 min-w-[140px]">Kategori</th>
              <th class="py-2.5 px-3 min-w-[120px]">Merk & Model</th>
              <th class="py-2.5 px-3 min-w-[130px]">No. Seri (S/N) *</th>
              <th class="py-2.5 px-3 w-20 text-center">Jumlah</th>
              <th class="py-2.5 px-3 w-24">Satuan</th>
              <th class="py-2.5 px-3 min-w-[140px]">Keterangan</th>
              <th class="py-2.5 px-2 w-12 text-center">Aksi</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100">
            <tr v-if="items.length === 0">
              <td colspan="9" class="py-8 text-center text-slate-400">
                Belum ada barang yang didaftarkan. Klik tombol "+ Tambah Baris Manual" atau "+ Impor dari Barang Rutin".
              </td>
            </tr>
            <tr
              v-for="(item, idx) in items"
              :key="idx"
              class="hover:bg-slate-50/60 transition-colors"
            >
              <td class="py-2 px-3 text-center text-slate-400 font-mono text-[11px]">
                {{ idx + 1 }}
              </td>

              <!-- Nama Barang -->
              <td class="py-2 px-3">
                <input
                  v-model="item.item_name"
                  required
                  type="text"
                  placeholder="Nama alat kerja"
                  class="w-full px-2 py-1 text-xs bg-white border border-slate-200 rounded text-slate-900 focus:outline-none focus:ring-1 focus:ring-inka-600 font-medium"
                />
              </td>

              <!-- Kategori -->
              <td class="py-2 px-3">
                <select
                  v-model="item.category_id"
                  class="w-full px-2 py-1 text-xs bg-white border border-slate-200 rounded text-slate-800 focus:outline-none focus:ring-1 focus:ring-inka-600"
                >
                  <option value="">-- Pilih --</option>
                  <option v-for="cat in categories" :key="cat.id" :value="cat.id">
                    {{ cat.name }}
                  </option>
                </select>
              </td>

              <!-- Merk & Model -->
              <td class="py-2 px-3">
                <input
                  v-model="item.brand"
                  type="text"
                  placeholder="Merk / Tipe"
                  class="w-full px-2 py-1 text-xs bg-white border border-slate-200 rounded text-slate-800 focus:outline-none focus:ring-1 focus:ring-inka-600"
                />
              </td>

              <!-- No. Seri -->
              <td class="py-2 px-3">
                <input
                  v-model="item.serial_number"
                  type="text"
                  placeholder="No. Seri Fisik"
                  class="w-full px-2 py-1 text-xs bg-white border border-slate-200 rounded text-slate-900 focus:outline-none focus:ring-1 focus:ring-inka-600 font-mono"
                />
              </td>

              <!-- Jumlah -->
              <td class="py-2 px-3">
                <input
                  v-model.number="item.quantity"
                  type="number"
                  min="1"
                  required
                  class="w-full px-2 py-1 text-xs bg-white border border-slate-200 rounded text-slate-900 focus:outline-none focus:ring-1 focus:ring-inka-600 text-center font-bold"
                />
              </td>

              <!-- Satuan -->
              <td class="py-2 px-3">
                <select
                  v-model="item.unit"
                  class="w-full px-2 py-1 text-xs bg-white border border-slate-200 rounded text-slate-800 focus:outline-none focus:ring-1 focus:ring-inka-600"
                >
                  <option value="Unit">Unit</option>
                  <option value="Set">Set</option>
                  <option value="Pcs">Pcs</option>
                  <option value="Buah">Buah</option>
                  <option value="Box">Box</option>
                </select>
              </td>

              <!-- Keterangan -->
              <td class="py-2 px-3">
                <input
                  v-model="item.description"
                  type="text"
                  placeholder="Catatan kondisi"
                  class="w-full px-2 py-1 text-xs bg-white border border-slate-200 rounded text-slate-700 focus:outline-none focus:ring-1 focus:ring-inka-600"
                />
              </td>

              <!-- Hapus -->
              <td class="py-2 px-2 text-center">
                <button
                  type="button"
                  @click="removeItemRow(idx)"
                  class="p-1 rounded text-slate-400 hover:text-rose-600 hover:bg-rose-50 transition-colors cursor-pointer"
                  title="Hapus baris"
                >
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                  </svg>
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- SECTION 4: ATTACHMENTS -->
    <div class="bg-white rounded-xl border border-slate-200 shadow-card p-5 space-y-4">
      <div class="flex items-center gap-2.5 pb-3 border-b border-slate-100">
        <div class="w-7 h-7 rounded-lg bg-inka-50 text-inka-700 flex items-center justify-center font-bold text-xs">
          4
        </div>
        <div>
          <h2 class="text-sm font-bold text-slate-900">Dokumen Lampiran (Surat Tugas / Dokumen Pendukung)</h2>
          <p class="text-[11px] text-slate-500">Lampirkan surat perintah kerja atau surat tugas resmi instansi rekanan (format PDF/Gambar)</p>
        </div>
      </div>

      <!-- Add attachment input -->
      <div class="flex flex-col sm:flex-row items-center gap-2 bg-slate-50 p-3 rounded-lg border border-slate-200">
        <input
          v-model="newAttachmentName"
          type="text"
          placeholder="Nama file, misal: Surat_Tugas_Vendor_2026.pdf"
          class="flex-1 px-3 py-1.5 text-xs bg-white border border-slate-300 rounded-lg text-slate-900 focus:outline-none focus:ring-1 focus:ring-inka-600"
        />
        <select
          v-model="newAttachmentType"
          class="px-3 py-1.5 text-xs bg-white border border-slate-300 rounded-lg text-slate-900 focus:outline-none focus:ring-1 focus:ring-inka-600"
        >
          <option value="SURAT_TUGAS">Surat Tugas Resmi</option>
          <option value="IDENTITAS">Kartu Identitas Pekerja</option>
          <option value="DOKUMEN_PENDUKUNG">Dokumen Pendukung Lain</option>
        </select>
        <Button variant="outline" size="sm" type="button" @click="addSampleAttachment">
          + Lampirkan Dokumen
        </Button>
      </div>

      <!-- Attachments list -->
      <div v-if="attachments.length > 0" class="space-y-2">
        <div
          v-for="(att, aIdx) in attachments"
          :key="aIdx"
          class="flex items-center justify-between p-2.5 rounded-lg border border-slate-200 bg-white"
        >
          <div class="flex items-center gap-2.5">
            <svg class="w-5 h-5 text-red-600 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z" />
            </svg>
            <div>
              <p class="text-xs font-bold text-slate-900">{{ att.file_name }}</p>
              <p class="text-[10px] text-slate-400">{{ att.attachment_type }} · {{ (att.file_size ? att.file_size / 1024 : 120).toFixed(0) }} KB</p>
            </div>
          </div>
          <button
            type="button"
            @click="removeAttachment(aIdx)"
            class="text-rose-600 hover:text-rose-800 text-xs font-semibold cursor-pointer"
          >
            Hapus
          </button>
        </div>
      </div>
    </div>

    <!-- BOTTOM STICKY ACTION BAR -->
    <div class="sticky bottom-4 bg-white/95 backdrop-blur border border-slate-300 rounded-xl p-4 shadow-lg flex flex-col sm:flex-row sm:items-center justify-between gap-3 z-20">
      <div class="text-xs text-slate-500">
        Total terdata: <strong class="text-slate-800">{{ items.length }} Barang</strong> ·
        Masa izin: <strong class="text-slate-800 font-mono">{{ startDate }} s.d {{ endDate }}</strong>
      </div>

      <div class="flex items-center gap-3">
        <router-link to="/simak">
          <Button variant="outline" size="sm" type="button">
            Batal
          </Button>
        </router-link>

        <Button
          variant="secondary"
          size="sm"
          type="button"
          :disabled="saving"
          @click="handleSave(false)"
        >
          {{ saving ? 'Menyimpan...' : 'Simpan sebagai Draft' }}
        </Button>

        <Button
          variant="primary"
          size="sm"
          type="button"
          :disabled="saving"
          @click="handleSave(true)"
        >
          {{ saving ? 'Mengajukan...' : 'Ajukan Permohonan SIMAK' }}
        </Button>
      </div>
    </div>

    <!-- MODAL: IMPORT DARI BARANG RUTIN SAYA -->
    <Modal
      v-model="showImportModal"
      title="Impor dari Barang Rutin Saya"
      subtitle="Pilih peralatan kerja yang tersimpan di profil Anda untuk langsung disalin ke permohonan SIMAK ini"
    >
      <div class="space-y-3">
        <div class="flex items-center justify-between pb-2 border-b border-slate-100 text-xs">
          <button
            type="button"
            @click="selectAllRoutineItems"
            class="text-inka-700 hover:text-inka-800 font-semibold cursor-pointer"
          >
            {{ selectedRoutineItemIds.length === routineItems.length ? 'Batalkan Semua' : 'Pilih Semua' }}
          </button>
          <span class="text-slate-400">
            {{ selectedRoutineItemIds.length }} dari {{ routineItems.length }} terpilih
          </span>
        </div>

        <div class="max-h-72 overflow-y-auto space-y-2">
          <div
            v-for="r in routineItems"
            :key="r.id"
            @click="toggleRoutineSelection(r.id)"
            :class="[
              'p-3 rounded-lg border cursor-pointer transition-all flex items-start gap-3',
              selectedRoutineItemIds.includes(r.id)
                ? 'bg-inka-50/40 border-inka-600'
                : 'bg-white border-slate-200 hover:bg-slate-50',
            ]"
          >
            <input
              type="checkbox"
              :checked="selectedRoutineItemIds.includes(r.id)"
              class="w-4 h-4 rounded text-inka-600 focus:ring-inka-500 border-slate-300 mt-0.5"
            />
            <div class="flex-1 min-w-0">
              <div class="flex items-center justify-between gap-2">
                <p class="text-xs font-bold text-slate-900 truncate">{{ r.item_name }}</p>
                <span class="text-[10px] font-semibold text-slate-500 bg-slate-100 px-1.5 py-0.5 rounded">
                  {{ r.quantity }} {{ r.unit }}
                </span>
              </div>
              <p class="text-[11px] text-slate-500 mt-0.5 font-mono">
                S/N: {{ r.serial_number || '-' }} · {{ r.brand || 'Tanpa Merk' }}
              </p>
            </div>
          </div>
        </div>
      </div>

      <template #footer>
        <Button variant="outline" size="sm" @click="showImportModal = false">
          Batal
        </Button>
        <Button
          variant="primary"
          size="sm"
          :disabled="selectedRoutineItemIds.length === 0"
          @click="applyImportRoutineItems"
        >
          Impor {{ selectedRoutineItemIds.length }} Barang Terpilih
        </Button>
      </template>
    </Modal>
  </div>
</template>
