<template>
  <span
    :class="badgeClasses"
  >
    <span v-if="pulse" class="flex h-2 w-2 mr-1">
      <span class="animate-ping absolute inline-flex h-2 w-2 rounded-full opacity-75" :class="pulseColor"></span>
      <span class="relative inline-flex rounded-full h-2 w-2" :class="pulseColor"></span>
    </span>
    <slot />
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  variant?: 'primary' | 'success' | 'warning' | 'danger' | 'error' | 'info'
  pulse?: boolean
  className?: string
}

const props = withDefaults(defineProps<Props>(), {
  variant: 'primary',
  pulse: false,
  className: ''
})

const badgeClasses = computed(() => {
  const classes = ['badge']
  
  switch (props.variant) {
    case 'primary':
      classes.push('badge-primary')
      break
    case 'success':
      classes.push('badge-success')
      break
    case 'warning':
      classes.push('badge-warning')
      break
    case 'danger':
    case 'error':
      classes.push('badge-danger')
      break
    case 'info':
      classes.push('badge-info')
      break
  }
  
  if (props.className) classes.push(props.className)
  
  return classes.join(' ')
})

const pulseColor = computed(() => {
  switch (props.variant) {
    case 'success':
      return 'bg-green-500'
    case 'warning':
      return 'bg-yellow-500'
    case 'danger':
    case 'error':
      return 'bg-red-500'
    case 'info':
      return 'bg-blue-500'
    default:
      return 'bg-indigo-500'
  }
})
</script>
