<script setup lang="ts">
import { computed, ref, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'
import { useNotificationStore } from '@/stores/notification'
import Badge from '@/components/ui/Badge.vue'

const route = useRoute()
const router = useRouter()
const appStore = useAppStore()
const authStore = useAuthStore()
const notifStore = useNotificationStore()

const pageTitle = computed(() => {
  return (route.meta?.title as string) || 'Dashboard Overview'
})

const notifOpen = ref(false)
const popoverRef = ref<HTMLElement | null>(null)

function toggleNotif() {
  notifOpen.value = !notifOpen.value
  if (notifOpen.value) {
    notifStore.fetchRecent()
  }
}

function closeNotif(e: MouseEvent) {
  if (popoverRef.value && !popoverRef.value.contains(e.target as Node)) {
    notifOpen.value = false
  }
}

function handleNotifClick(notif: any) {
  notifStore.markAsRead(notif.id)
  notifOpen.value = false
  if (notif.action_url) {
    router.push(notif.action_url)
  }
}

function handleMarkAllAsRead() {
  notifStore.markAllAsRead()
}

function goToAllNotifications() {
  notifOpen.value = false
  router.push('/notifications')
}

onMounted(() => {
  document.addEventListener('click', closeNotif)
  notifStore.fetchUnreadCount()
  notifStore.fetchRecent()
  notifStore.connectWebSocket()
})

onUnmounted(() => {
  document.removeEventListener('click', closeNotif)
  notifStore.disconnectWebSocket()
})
</script>

<template>
  <header
    class="h-16 bg-white border-b border-slate-200 px-6 flex items-center justify-between select-none shrink-0 sticky top-0 z-20"
  >
    <!-- Left: Breadcrumbs & Page Title -->
    <div class="flex items-center gap-3">
      <button
        @click="appStore.toggleSidebar"
        class="p-1.5 rounded-lg text-slate-500 hover:text-slate-800 hover:bg-slate-100 transition-colors cursor-pointer"
        title="Toggle sidebar"
      >
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
        </svg>
      </button>

      <div class="flex items-center gap-2 text-xs text-slate-400">
        <span class="hover:text-slate-600 transition-colors">SIMAK PT INKA</span>
        <span>/</span>
        <span class="text-slate-800 font-semibold">{{ pageTitle }}</span>
      </div>
    </div>

    <!-- Right: Global Search, and Notifications -->
    <div class="flex items-center gap-3">

      <!-- Quick Search Bar (Design Polish) -->
      <div class="relative hidden md:block">
        <input
          type="text"
          placeholder="Cari no. SIMAK, pekerja, alat... (Ctrl+K)"
          class="w-64 pl-8 pr-3 py-1.5 text-xs bg-slate-50 border border-slate-200 rounded-lg text-slate-700 placeholder-slate-400 focus:outline-none focus:ring-1 focus:ring-inka-600 focus:bg-white transition-all"
        />
        <svg
          class="w-3.5 h-3.5 text-slate-400 absolute left-2.5 top-2.5"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
          />
        </svg>
      </div>

      <!-- Notification Bell with Interactive Popover -->
      <div class="relative" ref="popoverRef">
        <button
          type="button"
          @click.stop="toggleNotif"
          class="p-2 rounded-lg text-slate-500 hover:text-slate-800 hover:bg-slate-100 transition-colors relative cursor-pointer"
          title="Notifikasi"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9"
            />
          </svg>
          <span
            v-if="notifStore.unreadCount > 0"
            class="absolute top-1 right-1 min-w-[18px] h-[18px] px-1 rounded-full bg-inka-600 text-white text-[10px] font-bold flex items-center justify-center border-2 border-white animate-pulse"
          >
            {{ notifStore.unreadCount > 99 ? '99+' : notifStore.unreadCount }}
          </span>
        </button>

        <!-- Dropdown Popover -->
        <div
          v-if="notifOpen"
          class="absolute right-0 mt-2 w-80 sm:w-96 bg-white border border-slate-200 rounded-2xl shadow-2xl z-50 overflow-hidden"
        >
          <div class="p-3.5 bg-slate-900 text-white flex items-center justify-between">
            <div class="flex items-center gap-2">
              <span class="font-bold text-xs uppercase tracking-wider">Notifikasi</span>
              <span
                v-if="notifStore.unreadCount > 0"
                class="px-1.5 py-0.2 rounded-full bg-inka-600 text-[10px] font-bold"
              >
                {{ notifStore.unreadCount }} Baru
              </span>
            </div>
            <button
              v-if="notifStore.unreadCount > 0"
              type="button"
              @click="handleMarkAllAsRead"
              class="text-[11px] text-slate-300 hover:text-white transition-colors cursor-pointer"
            >
              Tandai Dibaca
            </button>
          </div>

          <!-- Items list -->
          <div class="max-h-80 overflow-y-auto divide-y divide-slate-100">
            <div
              v-if="notifStore.recentNotifications.length === 0"
              class="p-6 text-center text-xs text-slate-500"
            >
              Tidak ada notifikasi terbaru.
            </div>

            <div
              v-for="item in notifStore.recentNotifications"
              :key="item.id"
              @click="handleNotifClick(item)"
              class="p-3 hover:bg-slate-50 transition-colors cursor-pointer flex items-start gap-2.5"
              :class="item.is_read ? 'bg-white' : 'bg-rose-50/20'"
            >
              <span
                class="w-2 h-2 rounded-full mt-1.5 shrink-0"
                :class="item.is_read ? 'bg-transparent' : 'bg-inka-600'"
              ></span>
              <div class="flex-1 min-w-0">
                <div class="font-bold text-xs text-slate-900 truncate">{{ item.title }}</div>
                <div class="text-[11px] text-slate-600 line-clamp-2 mt-0.5">{{ item.message }}</div>
              </div>
            </div>
          </div>

          <!-- Footer link to all notifications -->
          <div class="p-2.5 bg-slate-50 border-t border-slate-200 text-center">
            <button
              type="button"
              @click="goToAllNotifications"
              class="text-xs font-bold text-inka-600 hover:text-inka-700 transition-colors cursor-pointer"
            >
              Buka Semua Notifikasi →
            </button>
          </div>
        </div>
      </div>

      <!-- Role & Department Badge -->
      <Badge variant="inka" size="sm" class="hidden lg:inline-flex">
        {{ authStore.user?.department || 'Workshop Madiun' }}
      </Badge>
      <Badge variant="neutral" size="sm" class="hidden sm:inline-flex">
        {{ authStore.primaryRole }}
      </Badge>
    </div>
  </header>
</template>
