<script setup lang="ts">
import { onMounted } from 'vue'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'
import AppSidebar from '@/layouts/AppSidebar.vue'
import AppTopbar from '@/layouts/AppTopbar.vue'

const appStore = useAppStore()
const authStore = useAuthStore()

onMounted(() => {
  // Initialize auth session and fetch health
  authStore.initAuth()
  appStore.fetchHealth()
})
</script>

<template>
  <div class="flex h-screen w-screen overflow-hidden bg-canvas">
    <!-- Collapsible Sidebar -->
    <AppSidebar v-show="appStore.sidebarOpen" />

    <!-- Main Application Column -->
    <div class="flex-1 flex flex-col min-w-0 h-screen overflow-hidden">
      <!-- Topbar Header -->
      <AppTopbar />

      <!-- Scrollable Main Content Area -->
      <main class="flex-1 overflow-y-auto p-6">
        <div class="max-w-7xl mx-auto">
          <router-view />
        </div>
      </main>
    </div>
  </div>
</template>
