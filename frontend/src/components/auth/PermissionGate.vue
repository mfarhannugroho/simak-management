<script setup lang="ts">
import { computed } from 'vue'
import { useAuthStore } from '@/stores/auth'

interface Props {
  permission?: string
  role?: string
  roles?: string[]
}

const props = defineProps<Props>()
const authStore = useAuthStore()

const isPermitted = computed(() => {
  // If no restrictions provided, allow render
  if (!props.permission && !props.role && (!props.roles || !props.roles.length)) {
    return true
  }

  // Check role requirement
  if (props.role && !authStore.hasRole(props.role)) {
    return false
  }

  // Check multiple roles requirement
  if (props.roles && props.roles.length && !authStore.hasAnyRole(props.roles)) {
    return false
  }

  // Check granular permission requirement
  if (props.permission && !authStore.hasPermission(props.permission)) {
    return false
  }

  return true
})
</script>

<template>
  <template v-if="isPermitted">
    <slot />
  </template>
  <template v-else>
    <slot name="fallback" />
  </template>
</template>
