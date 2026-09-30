<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { adminUserService, type AdminUser, type RoleItem } from '@/services/admin'
import DataTable, { type Column } from '@/components/ui/DataTable.vue'
import Button from '@/components/ui/Button.vue'
import Badge from '@/components/ui/Badge.vue'
import Modal from '@/components/ui/Modal.vue'

const activeTab = ref<'users' | 'rbac'>('users')
const users = ref<AdminUser[]>([])
const roles = ref<RoleItem[]>([])
const loading = ref(false)
const search = ref('')
const selectedRole = ref('')
const selectedStatus = ref('')

// Edit Role Modal State
const editModalOpen = ref(false)
const selectedUser = ref<AdminUser | null>(null)
const targetRole = ref('')
const savingRole = ref(false)
const actionMessage = ref<string | null>(null)

const columns: Column[] = [
  { key: 'user', label: 'Pengguna & Akun', sortable: true },
  { key: 'nip', label: 'NIP / ID Pekerja' },
  { key: 'company', label: 'Perusahaan / Departemen' },
  { key: 'roles', label: 'Peran Sistem (Role)', align: 'center' },
  { key: 'status', label: 'Status Akun', align: 'center' },
  { key: 'last_login', label: 'Login Terakhir', align: 'center' },
  { key: 'actions', label: 'Aksi', align: 'right' },
]

async function fetchData() {
  loading.value = true
  actionMessage.value = null
  try {
    const [usersRes, rolesRes] = await Promise.all([
      adminUserService.getUsers({
        search: search.value,
        role: selectedRole.value,
      }),
      adminUserService.getRoles(),
    ])
    if (usersRes.data) {
      users.value = usersRes.data
    }
    if (rolesRes.data) {
      roles.value = rolesRes.data
    }
  } catch (err: any) {
    console.error('Gagal mengambil data pengguna:', err)
  } finally {
    loading.value = false
  }
}

const filteredUsers = computed(() => {
  return users.value.filter(u => {
    if (selectedStatus.value === 'ACTIVE' && !u.is_active) return false
    if (selectedStatus.value === 'INACTIVE' && u.is_active) return false
    return true
  })
})

const stats = computed(() => {
  const total = users.value.length
  const active = users.value.filter(u => u.is_active).length
  const rolesCount = roles.value.length || 6
  let totalPerms = 0
  roles.value.forEach(r => {
    if (r.permissions) totalPerms = Math.max(totalPerms, r.permissions.length)
  })
  return { total, active, rolesCount, totalPerms: totalPerms || 25 }
})

function getRoleVariant(roleName?: string): 'inka' | 'success' | 'warning' | 'neutral' | 'danger' {
  switch (roleName) {
    case 'Admin': return 'danger'
    case 'Approver': return 'warning'
    case 'SIMAK Officer': return 'success'
    case 'Applicant': return 'inka'
    case 'Gate Officer': return 'neutral'
    case 'Supervisor': return 'success'
    default: return 'neutral'
  }
}

function openEditRoleModal(user: AdminUser) {
  selectedUser.value = user
  targetRole.value = user.roles[0] || 'Applicant'
  editModalOpen.value = true
}

async function handleSaveRole() {
  if (!selectedUser.value || !targetRole.value) return
  savingRole.value = true
  try {
    await adminUserService.updateUserRole(selectedUser.value.id, targetRole.value)
    editModalOpen.value = false
    actionMessage.value = `Peran untuk ${selectedUser.value.full_name || selectedUser.value.email} berhasil diubah ke ${targetRole.value}`
    fetchData()
  } catch (err: any) {
    alert('Gagal memperbarui peran: ' + (err.response?.data?.message || err.message))
  } finally {
    savingRole.value = false
  }
}

async function handleToggleStatus(user: AdminUser) {
  const newStatus = !user.is_active
  const actionText = newStatus ? 'mengaktifkan kembali' : 'menonaktifkan'
  if (!confirm(`Apakah Anda yakin ingin ${actionText} akun ${user.full_name || user.email}?`)) {
    return
  }

  try {
    await adminUserService.toggleUserStatus(user.id, newStatus)
    actionMessage.value = `Status akun ${user.full_name || user.email} berhasil diperbarui`
    fetchData()
  } catch (err: any) {
    alert('Gagal mengubah status akun: ' + (err.response?.data?.message || err.message))
  }
}

function formatDate(dateStr?: string | null) {
  if (!dateStr) return 'Belum Pernah'
  const d = new Date(dateStr)
  return d.toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

onMounted(() => {
  fetchData()
})
</script>

<template>
  <div class="space-y-6 max-w-7xl mx-auto pb-16">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div>
        <h1 class="text-xl sm:text-2xl font-bold tracking-tight text-slate-900">
          Manajemen Pengguna & Otorisasi RBAC
        </h1>
        <p class="text-xs text-slate-500 mt-1">
          Pengelolaan 6 role pengguna, akun kedinasan PT INKA, dan pengawasan hak akses granular sistem.
        </p>
      </div>

      <!-- Tab Switcher -->
      <div class="inline-flex rounded-xl p-1 bg-slate-100 border border-slate-200">
        <button
          type="button"
          @click="activeTab = 'users'"
          :class="[
            'px-4 py-1.5 text-xs font-semibold rounded-lg transition-all cursor-pointer',
            activeTab === 'users'
              ? 'bg-white text-inka-700 shadow-xs'
              : 'text-slate-600 hover:text-slate-900',
          ]"
        >
          Daftar Pengguna
        </button>
        <button
          type="button"
          @click="activeTab = 'rbac'"
          :class="[
            'px-4 py-1.5 text-xs font-semibold rounded-lg transition-all cursor-pointer',
            activeTab === 'rbac'
              ? 'bg-white text-inka-700 shadow-xs'
              : 'text-slate-600 hover:text-slate-900',
          ]"
        >
          Matriks Peran (RBAC)
        </button>
      </div>
    </div>

    <!-- Alert Success Notification -->
    <div
      v-if="actionMessage"
      class="p-3.5 rounded-xl bg-emerald-50 border border-emerald-200 text-xs text-emerald-800 flex items-center justify-between shadow-xs"
    >
      <div class="flex items-center gap-2">
        <svg class="w-4 h-4 text-emerald-600 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
        </svg>
        <span>{{ actionMessage }}</span>
      </div>
      <button @click="actionMessage = null" class="text-emerald-700 hover:text-emerald-900 p-1">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
        </svg>
      </button>
    </div>

    <!-- KPI Summary Row -->
    <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
      <div class="bg-white border border-slate-200 rounded-xl p-4 shadow-xs">
        <p class="text-[11px] font-bold text-slate-500 uppercase tracking-wider">Total Pengguna</p>
        <p class="text-2xl font-black text-slate-900 mt-1">{{ stats.total }}</p>
        <p class="text-[11px] text-slate-400 mt-0.5">Akun terdaftar di database</p>
      </div>
      <div class="bg-white border border-slate-200 rounded-xl p-4 shadow-xs">
        <p class="text-[11px] font-bold text-slate-500 uppercase tracking-wider">Akun Aktif</p>
        <p class="text-2xl font-black text-emerald-600 mt-1">{{ stats.active }}</p>
        <p class="text-[11px] text-slate-400 mt-0.5">Dapat login & beraktivitas</p>
      </div>
      <div class="bg-white border border-slate-200 rounded-xl p-4 shadow-xs">
        <p class="text-[11px] font-bold text-slate-500 uppercase tracking-wider">Struktur Peran (Roles)</p>
        <p class="text-2xl font-black text-inka-700 mt-1">{{ stats.rolesCount }}</p>
        <p class="text-[11px] text-slate-400 mt-0.5">Admin, Worker, Approver, dll.</p>
      </div>
      <div class="bg-white border border-slate-200 rounded-xl p-4 shadow-xs">
        <p class="text-[11px] font-bold text-slate-500 uppercase tracking-wider">Hak Akses Granular</p>
        <p class="text-2xl font-black text-slate-900 mt-1">{{ stats.totalPerms }}</p>
        <p class="text-[11px] text-slate-400 mt-0.5">Permissions terlindungi RBAC</p>
      </div>
    </div>

    <!-- TAB 1: User Directory List -->
    <div v-if="activeTab === 'users'" class="space-y-4">
      <!-- Filters Bar -->
      <div class="bg-white border border-slate-200 rounded-xl p-4 shadow-xs flex flex-col md:flex-row items-stretch md:items-center justify-between gap-3">
        <div class="flex-1 flex flex-col sm:flex-row items-stretch sm:items-center gap-3">
          <!-- Search Input -->
          <div class="relative flex-1">
            <input
              v-model="search"
              type="text"
              placeholder="Cari nama, email, NIP pekerja..."
              @input="fetchData"
              class="w-full pl-9 pr-3 py-2 text-xs bg-slate-50 border border-slate-200 rounded-lg text-slate-800 placeholder-slate-400 focus:outline-none focus:ring-1 focus:ring-inka-600 focus:bg-white"
            />
            <svg class="w-4 h-4 text-slate-400 absolute left-3 top-2.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
          </div>

          <!-- Role Filter -->
          <select
            v-model="selectedRole"
            @change="fetchData"
            class="px-3 py-2 text-xs bg-slate-50 border border-slate-200 rounded-lg text-slate-700 focus:outline-none focus:ring-1 focus:ring-inka-600 font-medium"
          >
            <option value="">Semua Peran (Roles)</option>
            <option v-for="r in roles" :key="r.name" :value="r.name">{{ r.name }}</option>
          </select>

          <!-- Status Filter -->
          <select
            v-model="selectedStatus"
            class="px-3 py-2 text-xs bg-slate-50 border border-slate-200 rounded-lg text-slate-700 focus:outline-none focus:ring-1 focus:ring-inka-600 font-medium"
          >
            <option value="">Semua Status</option>
            <option value="ACTIVE">Aktif Saja</option>
            <option value="INACTIVE">Nonaktif Saja</option>
          </select>
        </div>

        <div class="flex items-center gap-2">
          <Button variant="secondary" size="sm" @click="fetchData" class="gap-1.5">
            <svg class="w-3.5 h-3.5 text-slate-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
            </svg>
            Segarkan
          </Button>
        </div>
      </div>

      <!-- User DataTable -->
      <DataTable
        :columns="columns"
        :items="filteredUsers"
        :loading="loading"
      >
        <!-- Slot: User -->
        <template #user="{ row }">
          <div class="flex items-center gap-3">
            <div class="w-9 h-9 rounded-full bg-slate-100 text-slate-700 border border-slate-200 flex items-center justify-center font-bold text-xs shrink-0">
              {{ (row.full_name || row.email).charAt(0).toUpperCase() }}
            </div>
            <div class="min-w-0">
              <p class="text-xs font-bold text-slate-900 truncate">
                {{ row.full_name || 'Tanpa Nama Profil' }}
              </p>
              <p class="text-[11px] text-slate-500 font-mono truncate">
                {{ row.email }}
              </p>
            </div>
          </div>
        </template>

        <!-- Slot: NIP -->
        <template #nip="{ row }">
          <span v-if="row.worker_id" class="font-mono text-xs font-semibold px-2 py-0.5 rounded bg-slate-100 text-slate-800">
            {{ row.worker_id }}
          </span>
          <span v-else class="text-slate-400 text-xs">-</span>
        </template>

        <!-- Slot: Company / Department -->
        <template #company="{ row }">
          <div class="space-y-0.5">
            <p class="text-xs text-slate-800 font-medium truncate">
              {{ row.company_name || 'PT INKA (Persero)' }}
            </p>
            <p class="text-[11px] text-slate-400 truncate">
              {{ row.department_name || row.position || '-' }}
            </p>
          </div>
        </template>

        <!-- Slot: Roles -->
        <template #roles="{ row }">
          <div class="flex flex-wrap items-center justify-center gap-1">
            <Badge
              v-for="r in row.roles"
              :key="r"
              :variant="getRoleVariant(r)"
              size="sm"
            >
              {{ r }}
            </Badge>
          </div>
        </template>

        <!-- Slot: Status -->
        <template #status="{ row }">
          <span
            v-if="row.is_active"
            class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[11px] font-semibold bg-emerald-50 text-emerald-700 border border-emerald-200"
          >
            <span class="w-1.5 h-1.5 rounded-full bg-emerald-500"></span>
            Aktif
          </span>
          <span
            v-else
            class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[11px] font-semibold bg-rose-50 text-rose-700 border border-rose-200"
          >
            <span class="w-1.5 h-1.5 rounded-full bg-rose-500"></span>
            Nonaktif
          </span>
        </template>

        <!-- Slot: Last Login -->
        <template #last_login="{ row }">
          <span class="text-[11px] text-slate-500 font-medium">
            {{ formatDate(row.last_login_at) }}
          </span>
        </template>

        <!-- Slot: Actions -->
        <template #actions="{ row }">
          <div class="flex items-center justify-end gap-1.5">
            <Button
              variant="secondary"
              size="sm"
              @click="openEditRoleModal(row)"
              class="text-xs font-semibold text-inka-700 hover:text-inka-800"
            >
              Ubah Role
            </Button>
            <Button
              :variant="row.is_active ? 'secondary' : 'primary'"
              size="sm"
              @click="handleToggleStatus(row)"
              :class="row.is_active ? 'text-rose-600 hover:bg-rose-50' : 'text-xs'"
            >
              {{ row.is_active ? 'Nonaktifkan' : 'Aktifkan' }}
            </Button>
          </div>
        </template>
      </DataTable>
    </div>

    <!-- TAB 2: RBAC Matrix Specification -->
    <div v-else class="space-y-6">
      <div class="bg-white border border-slate-200 rounded-2xl p-6 shadow-xs space-y-4">
        <div class="flex items-center justify-between border-b border-slate-100 pb-4">
          <div>
            <h2 class="text-base font-bold text-slate-900">
              Matriks Peran Pengguna & Granular Permission
            </h2>
            <p class="text-xs text-slate-500 mt-0.5">
              Daftar izin operasional bawaan yang melekat pada setiap peran di sistem SIMAK PT INKA (Persero).
            </p>
          </div>
          <span class="text-xs font-bold px-2.5 py-1 rounded bg-slate-100 text-slate-700 font-mono">
            6 Standar Peran
          </span>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 pt-2">
          <div
            v-for="role in roles"
            :key="role.id"
            class="p-4 rounded-xl border border-slate-200 bg-slate-50/50 hover:bg-white hover:border-inka-300 transition-all space-y-3"
          >
            <div class="flex items-center justify-between">
              <Badge :variant="getRoleVariant(role.name)" size="sm">
                {{ role.name }}
              </Badge>
              <span class="text-[11px] font-bold text-slate-400">
                {{ role.permissions?.length || 0 }} Izin
              </span>
            </div>

            <p class="text-xs text-slate-600 leading-relaxed min-h-[36px]">
              {{ role.description }}
            </p>

            <div class="pt-2 border-t border-slate-200/60">
              <p class="text-[10px] font-bold text-slate-400 uppercase tracking-wider mb-1.5">
                Cakupan Izin Utama:
              </p>
              <div class="flex flex-wrap gap-1 max-h-36 overflow-y-auto pr-1">
                <span
                  v-for="p in role.permissions"
                  :key="p.id"
                  class="font-mono text-[10px] px-1.5 py-0.5 rounded bg-white border border-slate-200 text-slate-700"
                  :title="p.description"
                >
                  {{ p.name }}
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Edit Role Modal -->
    <Modal
      v-model="editModalOpen"
      title="Ubah Peran Pengguna"
      @close="editModalOpen = false"
    >
      <div v-if="selectedUser" class="space-y-4">
        <div class="p-3 bg-slate-50 rounded-xl border border-slate-200 space-y-1">
          <p class="text-xs text-slate-500">Pengguna Terpilih:</p>
          <p class="text-sm font-bold text-slate-900">{{ selectedUser.full_name || selectedUser.email }}</p>
          <p class="text-xs text-slate-500 font-mono">{{ selectedUser.email }}</p>
        </div>

        <div>
          <label class="block text-xs font-bold text-slate-700 uppercase tracking-wider mb-2">
            Pilih Peran Baru:
          </label>
          <select
            v-model="targetRole"
            class="w-full px-3.5 py-2.5 text-xs bg-white border border-slate-300 rounded-lg text-slate-900 focus:outline-none focus:ring-2 focus:ring-inka-600"
          >
            <option v-for="r in roles" :key="r.name" :value="r.name">
              {{ r.name }} — {{ r.description }}
            </option>
          </select>
        </div>

        <div class="pt-4 border-t border-slate-100 flex items-center justify-end gap-2">
          <Button variant="secondary" @click="editModalOpen = false">
            Batal
          </Button>
          <Button variant="primary" :loading="savingRole" @click="handleSaveRole">
            Simpan Perubahan Role
          </Button>
        </div>
      </div>
    </Modal>
  </div>
</template>
