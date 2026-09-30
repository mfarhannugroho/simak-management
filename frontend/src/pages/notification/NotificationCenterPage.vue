<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { notificationService, type NotificationItem } from '@/services/notification'
import { useNotificationStore } from '@/stores/notification'
import Button from '@/components/ui/Button.vue'
import Badge from '@/components/ui/Badge.vue'

const router = useRouter()
const notifStore = useNotificationStore()

const items = ref<NotificationItem[]>([])
const loading = ref(false)
const activeTab = ref<'ALL' | 'UNREAD' | 'APPROVAL' | 'GATE' | 'SIMAK'>('ALL')

const page = ref(1)
const totalItems = ref(0)
const totalPages = ref(1)

const tabs = [
  { key: 'ALL', label: 'Semua Notifikasi' },
  { key: 'UNREAD', label: 'Belum Dibaca' },
  { key: 'APPROVAL', label: 'Persetujuan (Approval)' },
  { key: 'SIMAK', label: 'SIMAK & Masa Berlaku' },
  { key: 'GATE', label: 'Pos Gerbang & Temuan' },
]

async function fetchNotifications(p = 1) {
  loading.value = true
  page.value = p
  try {
    let typeParam = ''
    let unreadOnly = false

    if (activeTab.value === 'UNREAD') {
      unreadOnly = true
    } else if (activeTab.value === 'APPROVAL') {
      typeParam = 'APPROVAL_REQUIRED'
    } else if (activeTab.value === 'GATE') {
      typeParam = 'DISCREPANCY_CREATED'
    } else if (activeTab.value === 'SIMAK') {
      typeParam = 'SIMAK_ISSUED'
    }

    const res = await notificationService.getNotifications({
      page: p,
      limit: 15,
      unread_only: unreadOnly,
      type: typeParam,
    })

    if (res.data) {
      items.value = res.data.items || []
      totalItems.value = res.data.pagination.total_items
      totalPages.value = res.data.pagination.total_pages
    }
  } catch (err) {
    console.error('Failed to load notifications:', err)
  } finally {
    loading.value = false
  }
}

function handleTabChange(tabKey: any) {
  activeTab.value = tabKey
  fetchNotifications(1)
}

async function handleMarkAsRead(notif: NotificationItem) {
  if (notif.is_read) return
  await notifStore.markAsRead(notif.id)
  notif.is_read = true
}

async function handleMarkAllAsRead() {
  await notifStore.markAllAsRead()
  items.value.forEach((n) => (n.is_read = true))
}

function navigateToRef(notif: NotificationItem) {
  handleMarkAsRead(notif)
  if (notif.action_url) {
    router.push(notif.action_url)
  }
}

function formatRelativeTime(dateStr: string) {
  if (!dateStr) return '-'
  const d = new Date(dateStr)
  const now = new Date()
  const diffSec = Math.floor((now.getTime() - d.getTime()) / 1000)

  if (diffSec < 60) return 'Baru saja'
  if (diffSec < 3600) return `${Math.floor(diffSec / 60)} menit yang lalu`
  if (diffSec < 86400) return `${Math.floor(diffSec / 3600)} jam yang lalu`
  if (diffSec < 172800) return 'Kemarin'
  return d.toLocaleDateString('id-ID', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function getTypeIcon(type: string) {
  switch (type) {
    case 'SIMAK_APPROVED':
    case 'SIMAK_ISSUED':
      return { type: 'success', bg: 'bg-emerald-50 text-emerald-600 border border-emerald-200' }
    case 'SIMAK_REJECTED':
    case 'DISCREPANCY_CREATED':
      return { type: 'danger', bg: 'bg-rose-50 text-rose-600 border border-rose-200' }
    case 'APPROVAL_REQUIRED':
    case 'REVISION_REQUIRED':
      return { type: 'warning', bg: 'bg-amber-50 text-amber-600 border border-amber-200' }
    case 'SIMAK_EXPIRING':
    case 'SIMAK_EXPIRED':
      return { type: 'clock', bg: 'bg-orange-50 text-orange-600 border border-orange-200' }
    default:
      return { type: 'info', bg: 'bg-blue-50 text-blue-600 border border-blue-200' }
  }
}

onMounted(() => {
  fetchNotifications(1)
  notifStore.fetchUnreadCount()
})
</script>

<template>
  <div class="space-y-6 max-w-5xl mx-auto pb-16">
    <!-- Header Section -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div>
        <div class="flex items-center gap-2">
          <span class="px-2 py-0.5 text-xs font-semibold uppercase tracking-wider rounded bg-slate-100 text-slate-700">
            Pemberitahuan
          </span>
          <span class="text-xs text-slate-500 font-medium">WebSocket & In-App Notification</span>
        </div>
        <h1 class="text-2xl font-bold text-slate-900 tracking-tight mt-1">
          Pusat Notifikasi & Pemberitahuan Sistem
        </h1>
        <p class="text-sm text-slate-600 mt-0.5">
          Pantau pergerakan pengajuan, persetujuan bertingkat, verifikasi gerbang, dan temuan selisih barang secara seketika.
        </p>
      </div>

      <div class="flex items-center gap-2">
        <Button variant="secondary" size="sm" @click="handleMarkAllAsRead">
          <svg class="w-4 h-4 mr-1.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
          </svg>
          Tandai Semua Dibaca
        </Button>
      </div>
    </div>

    <!-- Filter Tabs -->
    <div class="flex items-center gap-2 overflow-x-auto pb-1 border-b border-slate-200">
      <button
        v-for="tab in tabs"
        :key="tab.key"
        type="button"
        @click="handleTabChange(tab.key)"
        class="px-4 py-2 rounded-lg text-xs font-semibold whitespace-nowrap transition-colors flex items-center gap-2 cursor-pointer"
        :class="
          activeTab === tab.key
            ? 'bg-inka-600 text-white shadow-sm'
            : 'bg-white text-slate-600 hover:bg-slate-100 hover:text-slate-900 border border-slate-200'
        "
      >
        <span>{{ tab.label }}</span>
        <span
          v-if="tab.key === 'UNREAD' && notifStore.unreadCount > 0"
          class="px-1.5 py-0.5 rounded-full text-[10px] font-bold"
          :class="activeTab === tab.key ? 'bg-white/20 text-white' : 'bg-rose-100 text-rose-700'"
        >
          {{ notifStore.unreadCount }}
        </span>
      </button>
    </div>

    <!-- Notification Feed -->
    <div class="bg-white border border-slate-200 rounded-2xl shadow-sm overflow-hidden">
      <!-- Loading State -->
      <div v-if="loading" class="p-12 text-center text-slate-500">
        <div class="inline-block w-6 h-6 border-2 border-inka-600 border-t-transparent rounded-full animate-spin"></div>
        <p class="text-xs mt-2">Memuat notifikasi...</p>
      </div>

      <!-- Empty State -->
      <div v-else-if="items.length === 0" class="p-12 text-center text-slate-500">
        <div class="w-12 h-12 rounded-full bg-slate-100 flex items-center justify-center mx-auto mb-3 text-slate-400">
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
          </svg>
        </div>
        <p class="text-sm font-semibold text-slate-700">Tidak ada notifikasi pada kategori ini</p>
        <p class="text-xs text-slate-400 mt-0.5">Semua pemberitahuan dan status operasional Anda telah terbarui.</p>
      </div>

      <!-- Notification List -->
      <div v-else class="divide-y divide-slate-100">
        <div
          v-for="notif in items"
          :key="notif.id"
          class="p-4 sm:p-5 transition-colors flex items-start justify-between gap-4 cursor-pointer hover:bg-slate-50"
          :class="notif.is_read ? 'bg-white' : 'bg-rose-50/20'"
          @click="navigateToRef(notif)"
        >
          <div class="flex items-start gap-3.5 flex-1 min-w-0">
            <!-- Icon -->
            <div
              class="w-10 h-10 rounded-xl flex items-center justify-center shrink-0"
              :class="getTypeIcon(notif.type).bg"
            >
              <svg v-if="getTypeIcon(notif.type).type === 'success'" class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
              </svg>
              <svg v-else-if="getTypeIcon(notif.type).type === 'danger'" class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
              </svg>
              <svg v-else-if="getTypeIcon(notif.type).type === 'warning'" class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
              </svg>
              <svg v-else-if="getTypeIcon(notif.type).type === 'clock'" class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <svg v-else class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
              </svg>
            </div>

            <!-- Content -->
            <div class="flex-1 min-w-0">
              <div class="flex items-center gap-2 flex-wrap mb-1">
                <span class="font-bold text-slate-900 text-sm">
                  {{ notif.title }}
                </span>
                <span
                  v-if="!notif.is_read"
                  class="w-2 h-2 rounded-full bg-inka-600 inline-block"
                  title="Belum dibaca"
                ></span>
                <span class="text-[11px] text-slate-400 font-medium">
                  • {{ formatRelativeTime(notif.created_at) }}
                </span>
              </div>

              <p class="text-xs text-slate-600 leading-relaxed">
                {{ notif.message }}
              </p>

              <!-- Action Link if present -->
              <div v-if="notif.action_url" class="mt-2 flex items-center gap-2">
                <span class="text-[11px] font-semibold text-inka-600 hover:text-inka-700 flex items-center gap-1">
                  Buka Rincian
                  <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
                  </svg>
                </span>
              </div>
            </div>
          </div>

          <!-- Mark As Read Button -->
          <div class="shrink-0 flex items-center gap-2" @click.stop>
            <button
              v-if="!notif.is_read"
              type="button"
              @click="handleMarkAsRead(notif)"
              class="p-1.5 rounded-lg text-slate-400 hover:text-slate-700 hover:bg-slate-100 cursor-pointer"
              title="Tandai telah dibaca"
            >
              <svg class="w-4 h-4 text-emerald-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
              </svg>
            </button>
          </div>
        </div>
      </div>

      <!-- Pagination Footer -->
      <div v-if="totalPages > 1" class="p-4 border-t border-slate-200 flex items-center justify-between text-xs text-slate-600">
        <div>Halaman {{ page }} dari {{ totalPages }}</div>
        <div class="flex items-center gap-2">
          <Button
            variant="secondary"
            size="sm"
            :disabled="page <= 1"
            @click="fetchNotifications(page - 1)"
          >
            Sebelumnya
          </Button>
          <Button
            variant="secondary"
            size="sm"
            :disabled="page >= totalPages"
            @click="fetchNotifications(page + 1)"
          >
            Selanjutnya
          </Button>
        </div>
      </div>
    </div>
  </div>
</template>
