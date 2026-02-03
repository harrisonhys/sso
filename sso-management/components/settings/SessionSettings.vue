<template>
  <div class="card p-6">
    <div class="flex justify-between items-center mb-4">
      <h3 class="text-lg font-semibold text-gray-900">Session Storage Configuration</h3>
      <span class="px-2 py-1 text-xs font-medium rounded-full" 
        :class="currentDriver === 'redis' ? 'bg-green-100 text-green-800' : 'bg-blue-100 text-blue-800'">
        Current Driver: {{ currentDriver.toUpperCase() }}
      </span>
    </div>

    <div class="space-y-6">
      <!-- Driver Selection -->
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-2">Storage Driver</label>
        <div class="flex gap-4">
          <label class="inline-flex items-center">
            <input type="radio" v-model="selectedDriver" value="db" class="form-radio text-indigo-600">
            <span class="ml-2">Database (MySQL)</span>
          </label>
          <label class="inline-flex items-center">
            <input type="radio" v-model="selectedDriver" value="redis" class="form-radio text-indigo-600">
            <span class="ml-2">Redis (In-Memory)</span>
          </label>
        </div>
        <p class="text-xs text-gray-500 mt-1">
          Database is simpler but slower. Redis is faster and recommended for high traffic.
        </p>
      </div>

      <!-- Redis Configuration -->
      <transition enter-active-class="transition ease-out duration-200" enter-from-class="opacity-0 -translate-y-2" enter-to-class="opacity-100 translate-y-0" leave-active-class="transition ease-in duration-150" leave-from-class="opacity-100 translate-y-0" leave-to-class="opacity-0 -translate-y-2">
        <div v-if="selectedDriver === 'redis'" class="bg-gray-50 p-4 rounded-md border border-gray-200">
          <h4 class="text-sm font-medium text-gray-900 mb-3">Redis Settings</h4>
          
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">Host</label>
              <input v-model="redisConfig.host" type="text" class="input text-sm" placeholder="127.0.0.1">
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">Port</label>
              <input v-model="redisConfig.port" type="text" class="input text-sm" placeholder="6379">
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">Password</label>
              <input v-model="redisConfig.password" type="password" class="input text-sm" placeholder="(Optional)">
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">DB Index</label>
              <input v-model.number="redisConfig.db" type="number" class="input text-sm" placeholder="0">
            </div>
          </div>

          <div class="mt-4 flex items-center gap-3">
             <button @click="testConnection" :disabled="testing" class="btn btn-sm btn-secondary">
               {{ testing ? 'Connecting...' : '🔌 Test Connection' }}
             </button>
             <span v-if="testResult" class="text-xs" :class="testResult.success ? 'text-green-600' : 'text-red-600'">
                {{ testResult.message }}
             </span>
          </div>
        </div>
      </transition>

      <!-- Actions -->
      <div class="flex justify-end pt-4 border-t border-gray-100">
        <button 
          @click="saveConfiguration" 
          :disabled="saving || (selectedDriver === 'redis' && !testResult?.success)"
          class="btn btn-primary"
        >
          {{ saving ? 'Applying...' : 'Apply & Save' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const config = useRuntimeConfig()
const props = defineProps<{
  initialData?: any
}>()

const currentDriver = ref('db') // fetched from config
const selectedDriver = ref('db')
const redisConfig = ref({
  host: 'host.docker.internal',
  port: '6379',
  password: '',
  db: 0
})

const testing = ref(false)
const saving = ref(false)
const testResult = ref<{success: boolean, message: string} | null>(null)

// Watch for driver change to reset test result
watch(selectedDriver, () => {
  testResult.value = null
})

const testConnection = async () => {
  testing.value = true
  testResult.value = null
  
  try {
    const res = await fetch(`${config.public.apiBase}/admin/api/config/session/test-redis`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(redisConfig.value)
    })
    
    const data = await res.json()
    
    if (res.ok) {
      testResult.value = { success: true, message: '✅ Connection Successful' }
    } else {
      testResult.value = { success: false, message: `❌ ${data.error}` }
    }
  } catch (error) {
    testResult.value = { success: false, message: '❌ Network Error' }
  } finally {
    testing.value = false
  }
}

const saveConfiguration = async () => {
  if (!confirm('Switching session driver might logout active users. Continue?')) return

  saving.value = true
  try {
    // 1. Save Redis Configs to System Config (Generic)
    if (selectedDriver.value === 'redis') {
        const configsToUpdate = {
            'redis.host': redisConfig.value.host,
            'redis.port': redisConfig.value.port,
            'redis.password': redisConfig.value.password,
            'redis.db': redisConfig.value.db.toString()
        }
        
        // This relies on the internal implementation of `SwitchSessionDriver` to possibly handle saving
        // OR we manually save each key. For robustness, let's send it to the switch endpoint.
    }

    const res = await fetch(`${config.public.apiBase}/admin/api/config/session/switch`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        driver: selectedDriver.value,
        redis: redisConfig.value // Sending this so backend can save it
      })
    })

    if (res.ok) {
      currentDriver.value = selectedDriver.value
      alert('Session storage configuration updated successfully.')
    } else {
      alert('Failed to update configuration.')
    }
  } catch (error) {
    console.error(error)
    alert('An unexpected error occurred.')
  } finally {
    saving.value = false
  }
}
</script>
