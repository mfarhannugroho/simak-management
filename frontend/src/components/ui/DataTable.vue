<script setup lang="ts">
import { ref, watch } from 'vue'
import Button from './Button.vue'

export interface Column {
  key: string
  label: string
  sortable?: boolean
  width?: string
  align?: 'left' | 'center' | 'right'
}

export interface Pagination {
  page: number
  limit: number
  total_items: number
  total_pages: number
}

interface Props {
  columns: Column[]
  items: any[]
  loading?: boolean
  pagination?: Pagination
  searchable?: boolean
  searchPlaceholder?: string
}

const props = withDefaults(defineProps<Props>(), {
  loading: false,
  searchable: true,
  searchPlaceholder: 'Cari data...',
})

const emit = defineEmits<{
  (e: 'search', query: string): void
  (e: 'page-change', page: number): void
  (e: 'sort', key: string, order: 'asc' | 'desc'): void
}>()

defineSlots<{
  toolbar?: (props?: any) => any
  [key: string]: ((scope: any) => any) | undefined
}>()

const searchQuery = ref('')
const sortKey = ref('')
const sortOrder = ref<'asc' | 'desc'>('asc')

let debounceTimer: any = null
watch(searchQuery, (newVal) => {
  clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    emit('search', newVal)
  }, 350)
})

function handleSort(key: string, sortable?: boolean) {
  if (!sortable) return
  if (sortKey.value === key) {
    sortOrder.value = sortOrder.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortKey.value = key
    sortOrder.value = 'asc'
  }
  emit('sort', sortKey.value, sortOrder.value)
}

function changePage(newPage: number) {
  if (!props.pagination) return
  if (newPage >= 1 && newPage <= props.pagination.total_pages) {
    emit('page-change', newPage)
  }
}
</script>

<template>
  <div class="bg-white border border-slate-200 rounded-xl shadow-card overflow-hidden">
    <!-- Top Table Header: Search & Custom Actions -->
    <div
      v-if="searchable || $slots.toolbar || $slots.actions"
      class="p-4 border-b border-slate-100 flex flex-col sm:flex-row sm:items-center justify-between gap-3 bg-white"
    >
      <div v-if="searchable" class="relative max-w-sm w-full">
        <input
          v-model="searchQuery"
          type="text"
          :placeholder="searchPlaceholder"
          class="w-full pl-9 pr-3 py-1.5 text-xs bg-slate-50 border border-slate-200 rounded-lg text-slate-800 placeholder-slate-400 focus:outline-none focus:ring-1 focus:ring-inka-600 focus:bg-white transition-all"
        />
        <svg
          class="w-4 h-4 text-slate-400 absolute left-2.5 top-2"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
        </svg>
      </div>
      <div v-else></div>

      <div v-if="$slots.toolbar" class="flex items-center gap-2">
        <slot name="toolbar" />
      </div>
    </div>

    <!-- Table Container -->
    <div class="overflow-x-auto">
      <table class="w-full text-left border-collapse">
        <thead>
          <tr class="bg-slate-50/80 border-b border-slate-200">
            <th
              v-for="col in columns"
              :key="col.key"
              :style="col.width ? { width: col.width } : {}"
              :class="[
                'px-4 py-3 text-[11px] font-bold text-slate-500 uppercase tracking-wider select-none',
                col.sortable ? 'cursor-pointer hover:text-slate-900 transition-colors' : '',
                col.align === 'center' ? 'text-center' : col.align === 'right' ? 'text-right' : 'text-left',
              ]"
              @click="handleSort(col.key, col.sortable)"
            >
              <div
                class="inline-flex items-center gap-1.5"
                :class="col.align === 'center' ? 'justify-center' : col.align === 'right' ? 'justify-end' : ''"
              >
                <span>{{ col.label }}</span>
                <span v-if="col.sortable" class="text-slate-400">
                  <svg
                    v-if="sortKey !== col.key"
                    class="w-3 h-3"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16V4m0 0L3 8m4-4l4 4m6 0v12m0 0l4-4m-4 4l-4-4" />
                  </svg>
                  <svg
                    v-else-if="sortOrder === 'asc'"
                    class="w-3 h-3 text-inka-700"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 15l7-7 7 7" />
                  </svg>
                  <svg
                    v-else
                    class="w-3 h-3 text-inka-700"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                  </svg>
                </span>
              </div>
            </th>
          </tr>
        </thead>
        <tbody class="divide-y divide-slate-100">
          <!-- Skeleton Loading -->
          <template v-if="loading">
            <tr v-for="i in 5" :key="i" class="animate-pulse">
              <td v-for="col in columns" :key="col.key" class="px-4 py-3.5">
                <div class="h-3.5 bg-slate-200 rounded w-3/4"></div>
              </td>
            </tr>
          </template>

          <!-- Empty State -->
          <tr v-else-if="!items || items.length === 0">
            <td :colspan="columns.length" class="px-4 py-12 text-center text-slate-400">
              <div class="flex flex-col items-center justify-center space-y-2">
                <svg class="w-8 h-8 text-slate-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4" />
                </svg>
                <p class="text-xs font-medium text-slate-500">Tidak ada data ditemukan</p>
                <p class="text-[11px] text-slate-400">Silakan ubah kata kunci pencarian atau tambahkan data baru</p>
              </div>
            </td>
          </tr>

          <!-- Data Rows -->
          <template v-else>
            <tr
              v-for="(row, idx) in items"
              :key="row.id || idx"
              class="hover:bg-slate-50/80 transition-colors"
            >
              <td
                v-for="col in columns"
                :key="col.key"
                :class="[
                  'px-4 py-3.5 text-xs text-slate-700',
                  col.align === 'center' ? 'text-center' : col.align === 'right' ? 'text-right' : 'text-left',
                ]"
              >
                <slot :name="col.key" :row="row" :value="row[col.key]">
                  <slot :name="'cell-' + col.key" :row="row" :value="row[col.key]">
                    <template v-if="typeof row[col.key] === 'object' && row[col.key] !== null">
                      {{ row[col.key].full_name || row[col.key].name || '-' }}
                    </template>
                    <template v-else>
                      {{ row[col.key] !== null && row[col.key] !== undefined ? row[col.key] : '-' }}
                    </template>
                  </slot>
                </slot>
              </td>
            </tr>
          </template>
        </tbody>
      </table>
    </div>

    <!-- Table Pagination Footer -->
    <div
      v-if="pagination && pagination.total_items > 0"
      class="px-4 py-3 border-t border-slate-100 bg-slate-50/60 flex flex-col sm:flex-row sm:items-center justify-between gap-3 text-xs text-slate-500"
    >
      <div>
        Menampilkan
        <span class="font-bold text-slate-700">
          {{ (pagination.page - 1) * pagination.limit + 1 }}
        </span>
        -
        <span class="font-bold text-slate-700">
          {{ Math.min(pagination.page * pagination.limit, pagination.total_items) }}
        </span>
        dari
        <span class="font-bold text-slate-700">{{ pagination.total_items }}</span>
        total data
      </div>

      <div class="flex items-center gap-1.5">
        <button
          @click="changePage(pagination.page - 1)"
          :disabled="pagination.page <= 1"
          class="px-2.5 py-1 rounded border border-slate-200 bg-white text-slate-600 hover:bg-slate-100 disabled:opacity-40 disabled:cursor-not-allowed cursor-pointer transition-colors text-xs"
        >
          &larr; Sebelumnya
        </button>

        <span class="px-2 py-1 font-semibold text-slate-700">
          Hal. {{ pagination.page }} / {{ pagination.total_pages }}
        </span>

        <button
          @click="changePage(pagination.page + 1)"
          :disabled="pagination.page >= pagination.total_pages"
          class="px-2.5 py-1 rounded border border-slate-200 bg-white text-slate-600 hover:bg-slate-100 disabled:opacity-40 disabled:cursor-not-allowed cursor-pointer transition-colors text-xs"
        >
          Berikutnya &rarr;
        </button>
      </div>
    </div>
  </div>
</template>
