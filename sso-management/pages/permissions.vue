<template>
  <NuxtLayout name="default" title="Permission Management">
    <div class="space-y-6">
      <!-- Page Header -->
      <div class="flex items-center justify-between">
        <div>
          <p class="text-sm text-gray-600">Define system permissions (Resource + Action)</p>
        </div>
        <button @click="openCreateModal" class="btn btn-primary">
          ➕ Create Permission
        </button>
      </div>

      <!-- Filters -->
      <div class="card p-4">
        <div class="flex gap-3">
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Search permissions..."
            class="input flex-1"
            @input="debouncedSearch"
          >
          <button @click="loadPermissions" class="btn btn-secondary">
            Search
          </button>
        </div>
      </div>

      <!-- Permissions Table -->
      <div class="card">
        <div class="overflow-x-auto">
          <table class="w-full">
            <thead>
              <tr class="border-b border-gray-200 bg-gray-50">
                <th class="text-left py-4 px-6 text-sm font-semibold text-gray-700">Name</th>
                <th class="text-left py-4 px-6 text-sm font-semibold text-gray-700">Resource</th>
                <th class="text-left py-4 px-6 text-sm font-semibold text-gray-700">Action</th>
                <th class="text-left py-4 px-6 text-sm font-semibold text-gray-700">Description</th>
                <th class="text-left py-4 px-6 text-sm font-semibold text-gray-700">Actions</th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="loading">
                <td colspan="5" class="py-8 text-center text-gray-500">Loading...</td>
              </tr>
              <tr v-else-if="permissions.length === 0">
                <td colspan="5" class="py-8 text-center text-gray-500">No permissions found</td>
              </tr>
              <tr v-else v-for="permission in permissions" :key="permission.id" class="border-b border-gray-100 hover:bg-gray-50">
                <td class="py-4 px-6 font-medium text-gray-900">{{ permission.name }}</td>
                <td class="py-4 px-6">
                  <span class="badge bg-blue-50 text-blue-700">{{ permission.resource }}</span>
                </td>
                <td class="py-4 px-6">
                  <span class="badge bg-green-50 text-green-700">{{ permission.action }}</span>
                </td>
                <td class="py-4 px-6 text-gray-600">{{ permission.description || '-' }}</td>
                <td class="py-4 px-6">
                  <div class="flex gap-2">
                    <button @click="openEditModal(permission)" class="text-primary-600 hover:text-primary-800" title="Edit">
                      ✏️
                    </button>
                    <button @click="deletePermission(permission)" class="text-red-600 hover:text-red-800" title="Delete">
                      🗑️
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Pagination -->
        <div v-if="totalPages > 1 || limit !== 20" class="p-4 border-t border-gray-200 flex items-center justify-between">
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

    <!-- Create/Edit Modal -->
    <div v-if="showModal" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" @click.self="closeModal">
      <div class="bg-white rounded-xl shadow-xl max-w-md w-full mx-4">
        <div class="p-6 border-b border-gray-200 flex items-center justify-between">
          <h3 class="text-xl font-semibold text-gray-900">{{ isEditing ? 'Edit Permission' : 'Create Permission' }}</h3>
          <button @click="closeModal" class="text-gray-400 hover:text-gray-600 text-2xl">&times;</button>
        </div>
        <form @submit.prevent="savePermission" class="p-6 space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Name (Unique)</label>
            <input v-model="form.name" type="text" required class="input" placeholder="e.g. users:read">
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Resource</label>
              <input v-model="form.resource" type="text" required class="input" placeholder="users">
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Action</label>
              <input v-model="form.action" type="text" required class="input" placeholder="read">
            </div>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Description</label>
            <textarea v-model="form.description" rows="2" class="input"></textarea>
          </div>
          
          <div class="flex gap-3 pt-4">
            <button type="button" @click="closeModal" class="btn btn-secondary flex-1">Cancel</button>
            <button type="submit" class="btn btn-primary flex-1">{{ isEditing ? 'Update' : 'Create' }}</button>
          </div>
        </form>
      </div>
    </div>
  </NuxtLayout>
</template>

<script setup lang="ts">
const config = useRuntimeConfig()
const loading = ref(false)
const showModal = ref(false)
const isEditing = ref(false)

const searchQuery = ref('')
const currentPage = ref(1)
const limit = ref(20)

const permissions = ref<any[]>([])
const totalPermissions = ref(0)
const totalPages = computed(() => Math.ceil(totalPermissions.value / limit.value))

const form = ref({
  id: '',
  name: '',
  resource: '',
  action: '',
  description: ''
})

// Debounced search
let searchTimeout: NodeJS.Timeout
const debouncedSearch = () => {
  clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    currentPage.value = 1
    loadPermissions()
  }, 500)
}

// Load permissions
const loadPermissions = async () => {
  loading.value = true
  try {
    const { data } = await useAdminFetch<any>(
      `/admin/api/permissions?page=${currentPage.value}&limit=${limit.value}&search=${encodeURIComponent(searchQuery.value)}`
    )
    if (data.value) {
      permissions.value = data.value.permissions || []
      totalPermissions.value = data.value.total || 0
    }
  } catch (error) {
    console.error('Failed to load permissions:', error)
  } finally {
    loading.value = false
  }
}

// Modal actions
const openCreateModal = () => {
  isEditing.value = false
  form.value = { id: '', name: '', resource: '', action: '', description: '' }
  showModal.value = true
}

const openEditModal = (permission: any) => {
  isEditing.value = true
  form.value = { ...permission }
  showModal.value = true
}

const closeModal = () => {
  showModal.value = false
}

// Save (Create/Update)
const savePermission = async () => {
  try {
    const url = isEditing.value 
      ? `${config.public.apiBase}/admin/api/permissions/${form.value.id}`
      : `${config.public.apiBase}/admin/api/permissions`
    
    const method = isEditing.value ? 'PUT' : 'POST'
    
    // Payload (exclude ID for create)
    const payload = { ...form.value }
    if (!isEditing.value) delete (payload as any).id

    const { error } = await useAdminFetch(url, {
      method,
      body: payload
    })
    
    if (!error.value) {
      closeModal()
      await loadPermissions()
      alert(isEditing.value ? 'Permission updated' : 'Permission created')
    } else {
      alert(error.value.message || 'Operation failed')
    }
  } catch (error) {
    console.error('Save failed:', error)
    alert('Operation failed')
  }
}

// Delete
const deletePermission = async (permission: any) => {
  if (!confirm(`Delete permission ${permission.name}?`)) return
  
  try {
    const { error } = await useAdminFetch(`/admin/api/permissions/${permission.id}`, {
      method: 'DELETE'
    })
    
    if (!error.value) {
      await loadPermissions()
    } else {
      alert('Failed to delete permission')
    }
  } catch (error) {
    alert('Failed to delete permission')
  }
}

const changePage = (page: number) => {
  currentPage.value = page
  loadPermissions()
}

const changeLimit = () => {
  currentPage.value = 1
  loadPermissions()
}

onMounted(() => {
  loadPermissions()
})

useHead({
  title: 'Permissions'
})
</script>
