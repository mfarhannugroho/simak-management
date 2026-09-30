<script setup lang="ts">
interface Props {
  label: string
  value: string | number
  trend?: string
  trendType?: 'positive' | 'negative' | 'neutral'
  description?: string
  accentColor?: string
  highlight?: boolean
}

withDefaults(defineProps<Props>(), {
  trendType: 'positive',
  highlight: false,
})
</script>

<template>
  <div
    :class="[
      'bg-white rounded-xl border border-slate-200 p-5 shadow-card relative transition-all hover:border-slate-300',
      highlight ? 'ring-1 ring-inka-500/20' : '',
    ]"
  >
    <div class="flex items-center justify-between">
      <span class="text-xs font-semibold uppercase tracking-wider text-slate-500">
        {{ label }}
      </span>
      <slot name="icon">
        <div class="w-8 h-8 rounded-lg bg-slate-50 flex items-center justify-center text-slate-500">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" />
          </svg>
        </div>
      </slot>
    </div>

    <div class="mt-3 flex items-baseline gap-2">
      <span class="text-3xl font-bold text-slate-900 tracking-tight">{{ value }}</span>
      <span
        v-if="trend"
        :class="[
          'inline-flex items-center text-xs font-medium px-1.5 py-0.5 rounded',
          trendType === 'positive' ? 'bg-emerald-50 text-emerald-700' : '',
          trendType === 'negative' ? 'bg-rose-50 text-rose-700' : '',
          trendType === 'neutral' ? 'bg-slate-100 text-slate-600' : '',
        ]"
      >
        {{ trend }}
      </span>
    </div>

    <p v-if="description" class="text-xs text-slate-500 mt-1">
      {{ description }}
    </p>

    <!-- Subtle bottom decorative sparkline/accent indicator -->
    <div class="mt-4 pt-3 border-t border-slate-100 flex items-center justify-between text-[11px] text-slate-400">
      <slot name="footer">
        <span>Diperbarui otomatis</span>
        <span class="w-2 h-2 rounded-full" :class="highlight ? 'bg-inka-600' : 'bg-slate-300'"></span>
      </slot>
    </div>
  </div>
</template>
