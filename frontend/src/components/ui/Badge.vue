<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  variant?: 'success' | 'warning' | 'danger' | 'neutral' | 'inka'
  size?: 'sm' | 'md'
  dot?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  variant: 'neutral',
  size: 'sm',
  dot: false,
})

const variantClasses = computed(() => {
  switch (props.variant) {
    case 'success':
      return 'bg-emerald-50 text-emerald-700 border-emerald-200'
    case 'warning':
      return 'bg-amber-50 text-amber-700 border-amber-200'
    case 'danger':
      return 'bg-rose-50 text-rose-700 border-rose-200'
    case 'inka':
      return 'bg-inka-50 text-inka-700 border-inka-200'
    case 'neutral':
    default:
      return 'bg-slate-100 text-slate-700 border-slate-200'
  }
})

const dotColorClasses = computed(() => {
  switch (props.variant) {
    case 'success':
      return 'bg-emerald-500'
    case 'warning':
      return 'bg-amber-500'
    case 'danger':
      return 'bg-rose-500'
    case 'inka':
      return 'bg-inka-600'
    case 'neutral':
    default:
      return 'bg-slate-500'
  }
})

const sizeClasses = computed(() => {
  return props.size === 'sm' ? 'px-2 py-0.5 text-xs' : 'px-2.5 py-1 text-sm'
})
</script>

<template>
  <span
    :class="[
      'inline-flex items-center gap-1.5 font-medium border rounded-full transition-colors',
      variantClasses,
      sizeClasses,
    ]"
  >
    <span v-if="dot" :class="['w-1.5 h-1.5 rounded-full', dotColorClasses]"></span>
    <slot />
  </span>
</template>
