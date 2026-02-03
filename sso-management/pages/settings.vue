<template>
  <NuxtLayout name="default" title="System Settings">
    <div class="space-y-6">
      <!-- Loading State -->
      <div v-if="loading" class="text-center py-12 text-gray-500">
        Loading configuration...
      </div>

      <!-- Main Content -->
      <div v-else class="space-y-6">
        <!-- Header Actions -->
        <div class="flex justify-end gap-2">
           <button v-if="hasChanges" @click="saveChanges" class="btn btn-primary" :disabled="saving">
            {{ saving ? 'Saving...' : '💾 Save Changes' }}
          </button>
           <button @click="loadConfigs" class="btn btn-secondary">
            🔄 Refresh
          </button>
        </div>

        <!-- Specialized Config Components -->
        <SettingsSessionSettings />

        <!-- Dynamic Categories -->
        <div v-for="(group, name) in groupedConfigs" :key="name" class="card p-6">
          <h3 class="text-lg font-semibold text-gray-900 mb-4 capitalize">{{ name }} Configuration</h3>
          
          <div v-if="group.length === 0" class="text-sm text-gray-400 italic">
            No configurations set.
          </div>
          
          <div v-else class="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div v-for="conf in group" :key="conf.config_key">
              <label class="block text-sm font-medium text-gray-700 mb-1">
                {{ formatKey(conf.config_key) }}
              </label>
              <div class="flex gap-2">
                <input 
                  v-model="editedValues[conf.config_key]" 
                  type="text" 
                  class="input flex-1"
                  :class="{'border-yellow-400 bg-yellow-50': isChanged(conf.config_key)}"
                >
                <button 
                  v-if="isChanged(conf.config_key)" 
                  @click="revertChange(conf.config_key)"
                  class="text-gray-400 hover:text-red-500"
                  title="Revert"
                >
                  ↩️
                </button>
              </div>
              <p class="text-xs text-gray-500 mt-1 font-mono">{{ conf.config_key }}</p>
            </div>
          </div>
        </div>

        <!-- Add New Config (Optional, for power users) -->
        <div class="card p-6 bg-gray-50 border-dashed">
           <h3 class="text-sm font-semibold text-gray-700 mb-2">➕ Add New Configuration</h3>
           <form @submit.prevent="addNewConfig" class="flex gap-3 items-end">
             <div class="flex-1">
               <label class="block text-xs text-gray-500 mb-1">Key (e.g. app.banner_text)</label>
               <input v-model="newConfig.key" type="text" class="input text-sm" placeholder="category.key_name" required>
             </div>
             <div class="flex-1">
               <label class="block text-xs text-gray-500 mb-1">Value</label>
               <input v-model="newConfig.value" type="text" class="input text-sm" placeholder="Value" required>
             </div>
             <button type="submit" class="btn btn-secondary text-sm">Add</button>
           </form>
        </div>
      </div>
    </div>
  </NuxtLayout>
</template>

<script setup lang="ts">
const config = useRuntimeConfig()
const loading = ref(false)
const saving = ref(false)

interface SystemConfig {
  config_key: string
  config_value: string
}

const configs = ref<SystemConfig[]>([])
const editedValues = ref<Record<string, string>>({})
const newConfig = ref({ key: '', value: '' })

// Group configs by prefix (e.g. "oauth2.", "security.")
const groupedConfigs = computed(() => {
  const groups: Record<string, SystemConfig[]> = {
    'General': [],
    'Security': [],
    'OAuth2': [],
    'Session': [],
    'Email': []
  }

  configs.value.forEach(conf => {
    if (conf.config_key.startsWith('oauth2.')) groups['OAuth2'].push(conf)
    else if (conf.config_key.startsWith('security.') || conf.config_key.startsWith('password.')) groups['Security'].push(conf)
    // Session configs are handled by specialized component
    // else if (conf.config_key.startsWith('session.')) groups['Session'].push(conf)
    else if (conf.config_key.startsWith('email.') || conf.config_key.startsWith('smtp.')) groups['Email'].push(conf)
    else groups['General'].push(conf)
  })

  // Filter out empty groups if simpler view desired, but keeping all for now
  return groups
})

const hasChanges = computed(() => {
  return configs.value.some(c => c.config_value !== editedValues.value[c.config_key])
})

const isChanged = (key: string) => {
  const original = configs.value.find(c => c.config_key === key)
  return original && original.config_value !== editedValues.value[key]
}

const formatKey = (key: string) => {
  // session.timeout -> Timeout
  const parts = key.split('.')
  const name = parts[parts.length - 1]
  return name.split('_').map(w => w.charAt(0).toUpperCase() + w.slice(1)).join(' ')
}

const loadConfigs = async () => {
  loading.value = true
  try {
    const { data } = await useAdminFetch<any>(`/admin/api/config`)
    if (data.value) {
      configs.value = data.value || []
      
      // Initialize edited values
      editedValues.value = {}
      configs.value.forEach(c => {
        editedValues.value[c.config_key] = c.config_value
      })
    }
  } catch (error) {
    console.error('Failed to load configs:', error)
  } finally {
    loading.value = false
  }
}

const revertChange = (key: string) => {
  const original = configs.value.find(c => c.config_key === key)
  if (original) {
    editedValues.value[key] = original.config_value
  }
}

const saveChanges = async () => {
  if (!confirm('Apply configuration changes? This may affect system behavior immediately.')) return

  saving.value = true
  try {
    // Find changed keys
    const changed = configs.value.filter(c => c.config_value !== editedValues.value[c.config_key])
    
    for (const c of changed) {
      const newValue = editedValues.value[c.config_key]
      await useAdminFetch(`/admin/api/config/${c.config_key}`, {
        method: 'PUT',
        body: { value: newValue }
      })
    }
    
    await loadConfigs()
    alert('Configuration saved successfully')
  } catch (error) {
    console.error('Failed to save:', error)
    alert('Failed to save some changes')
  } finally {
    saving.value = false
  }
}

const addNewConfig = async () => {
  if (!newConfig.value.key || !newConfig.value.value) return
  
  try {
    const { error } = await useAdminFetch(`/admin/api/config/${newConfig.value.key}`, {
      method: 'PUT',
      body: { value: newConfig.value.value }
    })

    if (!error.value) {
        newConfig.value = { key: '', value: '' }
        await loadConfigs()
    } else {
        alert('Failed to add config')
    }
  } catch (error) {
    alert('Failed to add config')
  }
}

onMounted(loadConfigs)

useHead({
  title: 'System Settings'
})
</script>
