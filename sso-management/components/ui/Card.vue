<template>
  <div 
    :class="[
      'card',
      variant === 'gradient' && 'card-gradient',
      variant === 'bordered' && 'card-bordered',
      hover && 'hover-lift',
      className
    ]"
  >
    <div v-if="$slots.header" class="px-6 py-4 border-b border-gray-200">
      <slot name="header" />
    </div>
    
    <div :class="['p-6', bodyClass]">
      <slot />
    </div>
    
    <div v-if="$slots.footer" class="px-6 py-4 border-t border-gray-200">
      <slot name="footer" />
    </div>
  </div>
</template>

<script setup lang="ts">
interface Props {
  variant?: 'default' | 'gradient' | 'bordered'
  hover?: boolean
  className?: string
  bodyClass?: string
}

withDefaults(defineProps<Props>(), {
  variant: 'default',
  hover: true,
  className: '',
  bodyClass: ''
})
</script>
