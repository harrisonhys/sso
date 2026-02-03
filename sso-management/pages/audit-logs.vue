<template>
  <NuxtLayout name="default" title="Audit Logs">
    <div class="space-y-6">
      <!-- Filters -->
      <div class="card p-4">
        <div class="grid grid-cols-1 md:grid-cols-3 gap-3">
          <input
            v-model="filters.search"
            type="text"
            placeholder="Search user or action..."
            class="input"
          >
          <select v-model="filters.action" class="input">
            <option value="">All Actions</option>
            <option value="login_success">Login Success</option>
            <option value="login_failed">Login Failed</option>
            <option value="password_changed">Password Changed</option>
            <option value="user_created">User Created</option>
            <option value="user_updated">User Updated</option>
          </select>
          <button @click="loadLogs" class="btn btn-primary">
            Apply Filters
          </button>
        </div>
      </div>

      <!-- Logs Table -->
      <div class="card">
        <div class="overflow-x-auto">
          <table class="w-full">
            <thead>
              <tr class="border-b border-gray-200 bg-gray-50">
                <th class="text-left py-4 px-6 text-sm font-semibold text-gray-700">Timestamp</th>
                <th class="text-left py-4 px-6 text-sm font-semibold text-gray-700">User</th>
                <th class="text-left py-4 px-6 text-sm font-semibold text-gray-700">Action</th>
                <th class="text-left py-4 px-6 text-sm font-semibold text-gray-700">Resource</th>
                <th class="text-left py-4 px-6 text-sm font-semibold text-gray-700">IP Address</th>
                <th class="text-left py-4 px-6 text-sm font-semibold text-gray-700">Status</th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="loading">
                <td colspan="6" class="py-8 text-center text-gray-500">Loading logs...</td>
              </tr>
              <tr v-else-if="logs.length === 0">
                <td colspan="6" class="py-8 text-center text-gray-500">No audit logs found</td>
              </tr>
              <tr v-else v-for="log in logs" :key="log.id" class="border-b border-gray-100 hover:bg-gray-50">
                <td class="py-3 px-6 text-sm text-gray-900">
                  {{ formatDateTime(log.created_at) }}
                </td>
                <td class="py-3 px-6 text-sm text-gray-900">
                  {{ log.User?.name || log.user_id || 'System' }}
                </td>
                <td class="py-3 px-6">
                  <code class="text-xs bg-gray-100 px-2 py-1 rounded">{{ log.action }}</code>
                </td>
                <td class="py-3 px-6 text-sm text-gray-600">
                  {{ log.resource || '-' }}
                </td>
                <td class="py-3 px-6 text-sm text-gray-600">
                  {{ log.ip_address || '-' }}
                </td>
                <td class="py-3 px-6">
                  <span :class="[
                    'badge',
                    log.status === 'success' ? 'badge-success' : 'badge-error'
                  ]">
                    {{ log.status || 'N/A' }}
                  </span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Pagination -->
        <div v-if="totalPages > 1 || limit !== 50" class="p-4 border-t border-gray-200 flex items-center justify-between">
           <div class="flex items-center gap-2 text-sm text-gray-600">
              <span>Show</span>
              <select v-model="limit" @change="changeLimit" class="border border-gray-300 rounded px-2 py-1 focus:ring-primary-500 focus:border-primary-500">
                <option :value="10">10</option>
                <option :value="20">20</option>
                <option :value="50">50</option>
                <option :value="100">100</option>
              </select>
              <span>per page</span>
           </div>

           <div class="flex gap-2">
             <button
            v-for="page in totalPages"
            :key="page"
            @click="changePage(page)"
            :class="[
              'px-4 py-2 rounded-lg font-medium transition-colors',
              page === currentPage ? 'bg-primary-600 text-white' : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
            ]"
          >
            {{ page }}
          </button>
        </div>
        </div>
      </div>
    </div>
  </NuxtLayout>
</template>

<script setup lang="ts">
const config = useRuntimeConfig()
const loading = ref(false)

const filters = ref({
  search: '',
  action: '',
  user_id: ''
})

const currentPage = ref(1)
const limit = ref(50)

const logs = ref<any[]>([])
const totalLogs = ref(0)
const totalPages = computed(() => Math.ceil(totalLogs.value / limit.value))

const formatDateTime = (dateString: string) => {
  return new Date(dateString).toLocaleString()
}

const loadLogs = async () => {
  loading.value = true
  try {
    const params = new URLSearchParams({
      page: currentPage.value.toString(),
      limit: limit.value.toString()
    })
    
    if (filters.value.action) params.append('action', filters.value.action)
    if (filters.value.user_id) params.append('user_id', filters.value.user_id)
    
    const { data } = await useAdminFetch<any>(`/admin/api/audit-logs?${params.toString()}`)
    
    if (data.value) {
      logs.value = data.value.logs || []
      totalLogs.value = data.value.total || 0
    }
  } catch (error) {
    console.error('Failed to load logs:', error)
  } finally {
    loading.value = false
  }
}

const changePage = (page: number) => {
  currentPage.value = page
  loadLogs()
}

const changeLimit = () => {
  currentPage.value = 1
  loadLogs()
}

onMounted(loadLogs)

useHead({
  title: 'Audit Logs'
})
</script>
