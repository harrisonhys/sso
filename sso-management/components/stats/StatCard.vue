<template>
  <div
    :class="[
      'stat-card',
      gradient && 'stat-card-gradient',
      gradientClass
    ]"
  >
    <div class="flex items-start justify-between">
      <div class="flex-1">
        <p :class="['stat-label', gradient && 'text-white']">{{ label }}</p>
        <p :class="['stat-value', gradient && 'text-white']">
          <span v-if="animated" ref="valueRef">{{ displayValue }}</span>
          <span v-else>{{ value }}</span>
        </p>
        <p v-if="change" :class="['text-sm mt-1', changeColor]">
          <span v-if="change > 0">↑</span>
          <span v-else-if="change < 0">↓</span>
          {{ Math.abs(change) }}% from last month
        </p>
      </div>
      
      <div :class="['stat-icon', iconBgClass]">
        <slot name="icon">
          <svg class="w-6 h-6" :class="iconColorClass" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" :d="defaultIconPath" />
          </svg>
        </slot>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'

interface Props {
  label: string
  value: number | string
  change?: number
  gradient?: boolean
  gradientClass?: string
  iconBgClass?: string
  iconColorClass?: string
  animated?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  change: 0,
  gradient: false,
  gradientClass: 'from-indigo-500 to-purple-600',
  iconBgClass: 'bg-indigo-100',
  iconColorClass: 'text-indigo-600',
  animated: true
})

const valueRef = ref<HTMLElement | null>(null)
const displayValue = ref(0)

const defaultIconPath = 'M13 7h8m0 0v8m0-8l-8 8-4-4-6 6'

const changeColor = computed(() => {
  if (props.change > 0) return 'text-green-600'
  if (props.change < 0) return 'text-red-600'
  return 'text-gray-600'
})

// Animate number count-up
const animateValue = (target: number) => {
  const duration = 1000
  const start = 0
  const startTime = Date.now()
  
  const updateValue = () => {
    const now = Date.now()
    const progress = Math.min((now - startTime) / duration, 1)
    const easeOutQuad = 1 - (1 - progress) * (1 - progress)
    displayValue.value = Math.floor(start + (target - start) * easeOutQuad)
    
    if (progress < 1) {
      requestAnimationFrame(updateValue)
    }
  }
  
  requestAnimationFrame(updateValue)
}

onMounted(() => {
  if (props.animated && typeof props.value === 'number') {
    animateValue(props.value)
  }
})

watch(() => props.value, (newValue) => {
  if (props.animated && typeof newValue === 'number') {
    animateValue(newValue)
  }
})
</script>
