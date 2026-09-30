<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Card from '@/components/ui/Card.vue'
import Button from '@/components/ui/Button.vue'
import Badge from '@/components/ui/Badge.vue'

const route = useRoute()
const router = useRouter()

const title = computed(() => (route.meta?.title as string) || 'Modul SIMAK')
const description = computed(() => (route.meta?.description as string) || 'Modul ini sedang disiapkan untuk operasional terpadu.')
</script>

<template>
  <div class="space-y-6">
    <Card>
      <div class="py-12 px-6 flex flex-col items-center text-center max-w-lg mx-auto">
        <div class="w-16 h-16 rounded-2xl bg-inka-50 text-inka-700 flex items-center justify-center mb-4 border border-inka-100 shadow-sm">
          <svg class="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
          </svg>
        </div>

        <h1 class="text-xl font-bold text-slate-900 mb-2">
          {{ title }}
        </h1>

        <p class="text-xs text-slate-500 mb-6 leading-relaxed">
          {{ description }}
        </p>

        <div class="flex items-center gap-3">
          <Button variant="outline" size="sm" @click="router.push('/')">
            Kembali ke Overview
          </Button>
          <a
            href="http://localhost:8080/swagger/index.html"
            target="_blank"
            rel="noopener"
            class="inline-flex items-center text-xs text-inka-700 hover:text-inka-800 font-medium p-2"
          >
            Lihat Spesifikasi API Swagger &rarr;
          </a>
        </div>
      </div>
    </Card>
  </div>
</template>
