<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  variant?: 'primary' | 'secondary' | 'outline' | 'ghost' | 'danger'
  size?: 'sm' | 'md' | 'lg'
  disabled?: boolean
  loading?: boolean
  type?: 'button' | 'submit' | 'reset'
}

const props = withDefaults(defineProps<Props>(), {
  variant: 'primary',
  size: 'md',
  disabled: false,
  loading: false,
  type: 'button',
})

const variantClasses = computed(() => {
  switch (props.variant) {
    case 'primary':
      return 'bg-inka-700 hover:bg-inka-800 text-white shadow-sm border border-inka-800 active:scale-[0.98]'
    case 'secondary':
      return 'bg-slate-100 hover:bg-slate-200 text-slate-800 border border-slate-200'
    case 'outline':
      return 'bg-white hover:bg-slate-50 text-slate-700 border border-slate-300'
    case 'ghost':
      return 'bg-transparent hover:bg-slate-100 text-slate-700'
    case 'danger':
      return 'bg-rose-600 hover:bg-rose-700 text-white shadow-sm'
    default:
      return 'bg-inka-700 text-white'
  }
})

const sizeClasses = computed(() => {
  switch (props.size) {
    case 'sm':
      return 'px-3 py-1.5 text-xs rounded-lg gap-1.5'
    case 'lg':
      return 'px-5 py-3 text-base rounded-xl gap-2.5 font-semibold'
    case 'md':
    default:
      return 'px-4 py-2 text-sm rounded-lg gap-2'
  }
})
</script>

<template>
  <button
    :type="type"
    :disabled="disabled || loading"
    :class="[
      'inline-flex items-center justify-center font-medium transition-all focus:outline-none focus:ring-2 focus:ring-inka-500 focus:ring-offset-1 disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer',
      variantClasses,
      sizeClasses,
    ]"
  >
    <svg
      v-if="loading"
      class="animate-spin -ml-1 mr-2 h-4 w-4 text-current"
      xmlns="http://www.w3.org/2000/svg"
      fill="none"
      viewBox="0 0 24 24"
    >
      <circle
        class="opacity-25"
        cx="12"
        cy="12"
        r="10"
        stroke="currentColor"
        stroke-width="4"
      ></circle>
      <path
        class="opacity-75"
        fill="currentColor"
        d="M4 12a8 8 0 018-8v8H4z"
      ></path>
    </svg>
    <slot />
  </button>
</template>
