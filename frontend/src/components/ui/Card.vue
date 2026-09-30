<script setup lang="ts">
interface Props {
  title?: string
  subtitle?: string
  noPadding?: boolean
}

defineProps<Props>()
</script>

<template>
  <div class="bg-white border border-slate-200 rounded-xl shadow-card transition-all overflow-hidden">
    <div
      v-if="title || $slots.header || $slots.actions"
      class="px-5 py-4 border-b border-slate-100 flex items-center justify-between"
    >
      <div>
        <h3 v-if="title" class="text-sm font-semibold text-slate-900">{{ title }}</h3>
        <p v-if="subtitle" class="text-xs text-slate-500 mt-0.5">{{ subtitle }}</p>
        <slot name="header" />
      </div>
      <div v-if="$slots.actions" class="flex items-center gap-2">
        <slot name="actions" />
      </div>
    </div>
    <div :class="noPadding ? '' : 'p-5'">
      <slot />
    </div>
    <div v-if="$slots.footer" class="px-5 py-3 bg-slate-50 border-t border-slate-100">
      <slot name="footer" />
    </div>
  </div>
</template>
