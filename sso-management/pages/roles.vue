<template>
  <NuxtLayout name="default" title="Role Management">
    <div class="space-y-6">
      <!-- Page Header -->
      <div class="flex items-center justify-between">
        <div>
          <p class="text-sm text-gray-600">Define user roles and assign permissions</p>
        </div>
        <button @click="openCreateModal" class="btn btn-primary">
          ➕ Create Role
        </button>
      </div>

      <!-- Roles List -->
      <div v-if="loading" class="text-center py-12 text-gray-500">
        Loading roles...
      </div>
      <div v-else-if="roles.length === 0" class="card p-12 text-center text-gray-500">
        No roles found
      </div>
      <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        <div v-for="role in roles" :key="role.id" class="card p-6 flex flex-col h-full">
          <div class="flex-1">
            <h3 class="text-xl font-semibold text-gray-900 mb-2">{{ role.name }}</h3>
            <p class="text-sm text-gray-600 mb-4">{{ role.description || 'No description' }}</p>
            
            <div class="bg-gray-50 rounded-lg p-3 mb-4">
              <span class="text-xs font-medium text-gray-500 uppercase tracking-wider">Permissions</span>
              <div class="mt-2 flex flex-wrap gap-1">
                <span v-if="!role.permissions || role.permissions.length === 0" class="text-sm text-gray-400 italic">
                  No permissions assigned
                </span>
                <template v-else>
                  <span v-for="perm in role.permissions.slice(0, 5)" :key="perm.id" class="badge bg-blue-50 text-blue-700 text-xs">
                    {{ perm.name }}
                  </span>
                  <span v-if="role.permissions.length > 5" class="badge bg-gray-100 text-gray-600 text-xs">
                    +{{ role.permissions.length - 5 }} more
                  </span>
                </template>
              </div>
            </div>
          </div>

          <div class="pt-4 border-t border-gray-100 flex gap-2 justify-end">
            <button @click="openEditModal(role)" class="btn btn-secondary text-sm">
              ✏️ Edit
            </button>
            <button @click="deleteRole(role)" class="btn btn-danger text-sm">
              🗑️ Delete
            </button>
          </div>
        </div>
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

    <!-- Create/Edit Modal -->
    <div v-if="showModal" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4" @click.self="closeModal">
      <div class="bg-white rounded-xl shadow-xl max-w-2xl w-full max-h-[90vh] flex flex-col">
        <div class="p-6 border-b border-gray-200 flex items-center justify-between">
          <h3 class="text-xl font-semibold text-gray-900">{{ isEditing ? 'Edit Role' : 'Create Role' }}</h3>
          <button @click="closeModal" class="text-gray-400 hover:text-gray-600 text-2xl">&times;</button>
        </div>
        
        <div class="p-6 overflow-y-auto flex-1">
          <form id="roleForm" @submit.prevent="saveRole" class="space-y-6">
            <!-- Basic Info -->
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">Role Name</label>
                <input v-model="form.name" type="text" required class="input" placeholder="e.g. Content Editor">
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">Description</label>
                <input v-model="form.description" type="text" class="input" placeholder="Role description">
              </div>
            </div>

            <!-- Permissions Selector -->
            <div>
              <div class="flex items-center justify-between mb-2">
                <label class="block text-sm font-medium text-gray-700">Permissions</label>
                <span class="text-xs text-gray-500">{{ selectedPermissionIds.length }} selected</span>
              </div>
              
              <div class="border border-gray-200 rounded-lg p-4 h-64 overflow-y-auto bg-gray-50">
                <div v-if="loadingPermissions" class="text-center py-4 text-gray-500">Loading permissions...</div>
                <div v-else class="grid grid-cols-1 md:grid-cols-2 gap-2">
                  <label v-for="perm in allPermissions" :key="perm.id" class="flex items-center p-2 hover:bg-white rounded cursor-pointer transition-colors border border-transparent hover:border-gray-200">
                    <input 
                      type="checkbox" 
                      :value="perm.id" 
                      v-model="selectedPermissionIds"
                      class="rounded text-primary-600 focus:ring-primary-500 mr-3"
                    >
                    <div>
                      <div class="text-sm font-medium text-gray-900">{{ perm.name }}</div>
                      <div class="text-xs text-gray-500">{{ perm.resource }}:{{ perm.action }}</div>
                    </div>
                  </label>
                </div>
              </div>
            </div>
          </form>
        </div>

        <div class="p-6 border-t border-gray-200 flex gap-3 bg-gray-50 rounded-b-xl">
          <button type="button" @click="closeModal" class="btn btn-secondary flex-1">Cancel</button>
          <button type="submit" form="roleForm" class="btn btn-primary flex-1">{{ isEditing ? 'Update Role' : 'Create Role' }}</button>
        </div>
      </div>
    </div>
  </NuxtLayout>
</template>

<script setup lang="ts">
const config = useRuntimeConfig()
const loading = ref(false)
const loadingPermissions = ref(false)
const showModal = ref(false)
const isEditing = ref(false)

const roles = ref<any[]>([])
const allPermissions = ref<any[]>([])
const currentPage = ref(1)
const limit = ref(20)
const totalRoles = ref(0)
const totalPages = computed(() => Math.ceil(totalRoles.value / limit.value))

const form = ref({
  id: '',
  name: '',
  description: ''
})
const selectedPermissionIds = ref<string[]>([])
const originalPermissionIds = ref<string[]>([]) // For diffing in edit mode

// Load roles
const loadRoles = async () => {
  loading.value = true
  try {
    const { data } = await useAdminFetch<any>(`/admin/api/roles?page=${currentPage.value}&limit=${limit.value}`)
    if (data.value) {
      roles.value = data.value.roles || []
      totalRoles.value = data.value.total || 0
    }
  } catch (error) {
    console.error('Failed to load roles:', error)
  } finally {
    loading.value = false
  }
}

const changePage = (page: number) => {
  currentPage.value = page
  loadRoles()
}

const changeLimit = () => {
  currentPage.value = 1
  loadRoles()
}

// Load all permissions for selector
const loadPermissions = async () => {
  loadingPermissions.value = true
  try {
    const { data } = await useAdminFetch<any>(`/admin/api/permissions?limit=1000`)
    if (data.value) {
      allPermissions.value = data.value.permissions || []
    }
  } catch (error) {
    console.error('Failed to load permissions:', error)
  } finally {
    loadingPermissions.value = false
  }
}

// Modal actions
const openCreateModal = async () => {
  isEditing.value = false
  form.value = { id: '', name: '', description: '' }
  selectedPermissionIds.value = []
  originalPermissionIds.value = []
  showModal.value = true
  if (allPermissions.value.length === 0) await loadPermissions()
}

const openEditModal = async (role: any) => {
  isEditing.value = true
  form.value = { id: role.id, name: role.name, description: role.description }
  
  // Extract existing permission IDs
  const currentIds = role.permissions ? role.permissions.map((p: any) => p.id) : []
  selectedPermissionIds.value = [...currentIds]
  originalPermissionIds.value = [...currentIds]
  
  showModal.value = true
  if (allPermissions.value.length === 0) await loadPermissions()
}

const closeModal = () => {
  showModal.value = false
}

// Save Role
const saveRole = async () => {
  try {
    let roleId = form.value.id

    // 1. Create/Update Role Basic Info
    const url = isEditing.value 
      ? `/admin/api/roles/${roleId}`
      : `/admin/api/roles`
    
    const method = isEditing.value ? 'PUT' : 'POST'
    
    // Payload for role info
    const rolePayload = { name: form.value.name, description: form.value.description }

    const { data, error } = await useAdminFetch<any>(url, {
      method,
      body: rolePayload
    })
    
    if (error.value) {
      alert(error.value.message || 'Failed to save role')
      return
    }

    if (!isEditing.value && data.value) {
      roleId = data.value.id
    }

    // 2. Handle Permissions
    await handlePermissionChanges(roleId)

    closeModal()
    await loadRoles()
    alert(isEditing.value ? 'Role updated' : 'Role created')
    
  } catch (error) {
    console.error('Save failed:', error)
    alert('Operation failed')
  }
}

// Handle Permission Diffing & assignment
const handlePermissionChanges = async (roleId: string) => {
  const current = new Set(originalPermissionIds.value)
  const selected = new Set(selectedPermissionIds.value)

  // To Add: In selected but not in current
  const toAdd = [...selected].filter(id => !current.has(id))
  
  // To Remove: In current but not in selected
  const toRemove = [...current].filter(id => !selected.has(id))

  // Execute Additions
  if (toAdd.length > 0) {
    await useAdminFetch(`/admin/api/roles/${roleId}/permissions`, {
      method: 'POST',
      body: { permission_ids: toAdd }
    })
  }

  // Execute Removals (one by one as API is single delete)
  // Backend API: DELETE /roles/:id/permissions/:permission_id
  for (const permId of toRemove) {
    await useAdminFetch(`/admin/api/roles/${roleId}/permissions/${permId}`, {
      method: 'DELETE'
    })
  }
}

const deleteRole = async (role: any) => {
  if (!confirm(`Delete role ${role.name}?`)) return
  
  try {
  try {
    const { error } = await useAdminFetch(`/admin/api/roles/${role.id}`, {
      method: 'DELETE'
    })
    
    if (!error.value) {
      await loadRoles()
    } else {
      alert('Failed to delete role')
    }
  } catch (error) {
    alert('Failed to delete role')
  }
  } catch (error) {
    alert('Failed to delete role')
  }
}

onMounted(loadRoles)

useHead({
  title: 'Roles'
})
</script>
