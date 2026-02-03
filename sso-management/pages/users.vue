<template>
  <NuxtLayout name="default" title="User Management">
    <div class="space-y-6">
      <!-- Page Header -->
      <div class="flex items-center justify-between">
        <div>
          <p class="text-sm text-gray-600">Manage system users, roles, and permissions</p>
        </div>
        <button @click="showCreateModal = true" class="btn btn-primary">
          ➕ Create User
        </button>
      </div>

      <!-- Filters -->
      <div class="card p-4">
        <div class="flex gap-3">
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Search by name or email..."
            class="input flex-1"
            @input="debouncedSearch"
          >
          <button @click="loadUsers" class="btn btn-secondary">
            Search
          </button>
        </div>
      </div>

      <!-- Users Table -->
      <div class="card">
        <div class="overflow-x-auto">
          <table class="w-full">
            <thead>
              <tr class="border-b border-gray-200 bg-gray-50">
                <th class="text-left py-4 px-6 text-sm font-semibold text-gray-700">Name</th>
                <th class="text-left py-4 px-6 text-sm font-semibold text-gray-700">Email</th>
                <th class="text-left py-4 px-6 text-sm font-semibold text-gray-700">Status</th>
                <th class="text-left py-4 px-6 text-sm font-semibold text-gray-700">2FA</th>
                <th class="text-left py-4 px-6 text-sm font-semibold text-gray-700">Created</th>
                <th class="text-left py-4 px-6 text-sm font-semibold text-gray-700">Actions</th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="loading">
                <td colspan="6" class="py-8 text-center text-gray-500">Loading...</td>
              </tr>
              <tr v-else-if="users.length === 0">
                <td colspan="6" class="py-8 text-center text-gray-500">No users found</td>
              </tr>
              <tr v-else v-for="user in users" :key="user.id" class="border-b border-gray-100 hover:bg-gray-50">
                <td class="py-4 px-6">
                  <div class="font-medium text-gray-900">{{ user.name }}</div>
                </td>
                <td class="py-4 px-6 text-gray-600">{{ user.email }}</td>
                <td class="py-4 px-6">
                  <span :class="[
                    'badge',
                    user.is_active ? 'badge-success' : 'badge-error'
                  ]">
                    {{ user.is_active ? 'Active' : 'Inactive' }}
                  </span>
                </td>
                <td class="py-4 px-6">
                  {{ user.two_factor_enabled ? '✅' : '❌' }}
                </td>
                <td class="py-4 px-6 text-gray-600">
                  {{ formatDate(user.created_at) }}
                </td>
                <td class="py-4 px-6">
                  <div class="flex gap-2">
                    <button @click="viewUser(user)" class="text-primary-600 hover:text-primary-800" title="View">
                      👁️
                    </button>
                    <button @click="resetPassword(user)" class="text-primary-600 hover:text-primary-800" title="Reset Password">
                      🔑
                    </button>
                    <button v-if="user.locked_until" @click="unlockUser(user)" class="text-green-600 hover:text-green-800" title="Unlock">
                      🔓
                    </button>
                    <button @click="deactivateUser(user)" class="text-red-600 hover:text-red-800" title="Deactivate">
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

    <!-- Create/Edit User Modal -->
    <div v-if="showCreateModal" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4" @click.self="showCreateModal = false">
      <div class="bg-white rounded-xl shadow-xl max-w-lg w-full max-h-[90vh] overflow-y-auto">
        <div class="p-6 border-b border-gray-200 flex items-center justify-between sticky top-0 bg-white">
          <h3 class="text-xl font-semibold text-gray-900">{{ isEditing ? 'Edit User' : 'Create New User' }}</h3>
          <button @click="showCreateModal = false" class="text-gray-400 hover:text-gray-600 text-2xl">&times;</button>
        </div>
        <form @submit.prevent="saveUser" class="p-6 space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Name</label>
            <input v-model="form.name" type="text" required class="input">
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Email</label>
            <input v-model="form.email" type="email" required class="input">
          </div>
          <div v-if="!isEditing">
            <label class="block text-sm font-medium text-gray-700 mb-1">Password</label>
            <input v-model="form.password" type="password" required class="input">
          </div>
          
          <div v-if="isEditing">
             <label class="flex items-center">
                <input type="checkbox" v-model="form.is_active" class="mr-2 rounded text-primary-600">
                <span class="text-sm font-medium text-gray-700">Active Account</span>
             </label>
          </div>

          <!-- Roles Selection -->
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-2">Assign Roles</label>
            <div class="border border-gray-200 rounded-lg p-3 max-h-40 overflow-y-auto bg-gray-50">
               <div v-if="allRoles.length === 0" class="text-sm text-gray-500 text-center py-2">No roles available</div>
               <div v-else class="space-y-2">
                  <label v-for="role in allRoles" :key="role.id" class="flex items-center">
                    <input type="checkbox" :value="role.id" v-model="form.role_ids" class="mr-2 rounded text-primary-600">
                    <span class="text-sm text-gray-900">{{ role.name }}</span>
                  </label>
               </div>
            </div>
          </div>

          <div class="flex gap-3 pt-4">
            <button type="button" @click="showCreateModal = false" class="btn btn-secondary flex-1">Cancel</button>
            <button type="submit" class="btn btn-primary flex-1">{{ isEditing ? 'Update User' : 'Create User' }}</button>
          </div>
        </form>
      </div>
    </div>
  </NuxtLayout>
</template>

<script setup lang="ts">
const config = useRuntimeConfig()
const loading = ref(false)
const showCreateModal = ref(false)
const isEditing = ref(false) // Track edit mode

const searchQuery = ref('')
const currentPage = ref(1)
const limit = ref(20)

const users = ref<any[]>([])
const allRoles = ref<any[]>([]) // Store available roles
const totalUsers = ref(0)
const totalPages = computed(() => Math.ceil(totalUsers.value / limit.value))

const form = ref({
  id: '',
  name: '',
  email: '',
  password: '',
  role_ids: [] as string[],
  is_active: true
})

// Store original roles for diffing during edit
const originalRoleIds = ref<string[]>([])

// Format date helper
const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleDateString()
}

// Debounced search
let searchTimeout: NodeJS.Timeout
const debouncedSearch = () => {
  clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    currentPage.value = 1
    loadUsers()
  }, 500)
}

// Load users
const loadUsers = async () => {
  loading.value = true
  try {
    const { data: userData, error } = await useAdminFetch<any>(
      `/admin/api/users?page=${currentPage.value}&limit=${limit.value}&search=${encodeURIComponent(searchQuery.value)}`
    )
    
    if (userData.value) {
      users.value = userData.value.users || []
      totalUsers.value = userData.value.total || 0
    } else if (error.value) {
        throw error.value
    }
  } catch (error) {
    console.error('Failed to load users:', error)
    // alert('Failed to load users') // useAdminFetch handles 401 redirect
  } finally {
    loading.value = false
  }
}

// Load roles for dropdown
const loadRoles = async () => {
  try {
    const { data } = await useAdminFetch<any>(`/admin/api/roles?limit=100`)
    if (data.value) {
      allRoles.value = data.value.roles || []
    }
  } catch (error) {
    console.error('Failed to load roles:', error)
  }
}

// Open Create Modal
const openCreateModal = async () => {
  isEditing.value = false
  form.value = {
    id: '',
    name: '',
    email: '',
    password: '',
    role_ids: [],
    is_active: true
  }
  showCreateModal.value = true
  if (allRoles.value.length === 0) await loadRoles()
}

// View User (alias to Edit for now)
const viewUser = (user: any) => {
  openEditModal(user)
}

// Open Edit Modal
const openEditModal = async (user: any) => {
  isEditing.value = true
  
  // Extract Role IDs from user object (assuming user.roles is populated by backend)
  // If backend doesn't return full roles in list, we might need to fetch single user. 
  // For now assuming list returns roles or we fetch single user here.
  // Actually, list usually doesn't return full details. Let's fetch single user to be safe.
  
  try {
    const { data: fullUser, error } = await useAdminFetch<any>(`/admin/api/users/${user.id}`)
    
    if (fullUser.value) {
      const currentRoleIds = fullUser.value.roles ? fullUser.value.roles.map((r: any) => r.id) : []
      
      form.value = {
        id: fullUser.value.id,
        name: fullUser.value.name,
        email: fullUser.value.email,
        password: '', // Password not editable directly here usually, or leave empty to keep
        role_ids: [...currentRoleIds], // Clone
        is_active: fullUser.value.is_active
      }
      originalRoleIds.value = [...currentRoleIds]
      
      showCreateModal.value = true
      if (allRoles.value.length === 0) await loadRoles()
    } else {
        throw error.value || new Error('Failed to load')
    }
  } catch (e) {
    alert('Failed to load user details')
  }
}

// Save User (Create or Update)
const saveUser = async () => {
  try {
    if (isEditing.value) {
      // Update User
      await updateUser()
    } else {
      // Create User
      await createUser()
    }
  } catch (error) {
    console.error('Save failed:', error)
    alert('Operation failed')
  }
}

const createUser = async () => {
  const { data, error } = await useAdminFetch<any>(`/admin/api/users`, {
    method: 'POST',
    body: {
      name: form.value.name,
      email: form.value.email,
      password: form.value.password,
      role_ids: form.value.role_ids
    }
  })
  
  if (!error.value) {
    showCreateModal.value = false
    await loadUsers()
    alert('User created successfully')
  } else {
    alert(error.value.message || 'Failed to create user')
  }
}

const updateUser = async () => {
  const userId = form.value.id
  
  // 1. Update Basic Info
  const { error } = await useAdminFetch(`/admin/api/users/${userId}`, {
    method: 'PUT',
    body: {
      name: form.value.name,
      email: form.value.email,
      is_active: form.value.is_active
    }
  })

  if (error.value) {
    throw new Error(error.value.message || 'Failed to update user info')
  }

  // 2. Handle Role Changes
  await handleRoleChanges(userId)
  
  showCreateModal.value = false
  await loadUsers()
  alert('User updated successfully')
}

// Diff roles and apply changes
const handleRoleChanges = async (userId: string) => {
  const current = new Set(originalRoleIds.value)
  const selected = new Set(form.value.role_ids)

  const toAdd = [...selected].filter(id => !current.has(id))
  const toRemove = [...current].filter(id => !selected.has(id))

  // Add Roles (One by one as per API)
  for (const roleId of toAdd) {
     await useAdminFetch(`/admin/api/users/${userId}/roles/${roleId}`, {
      method: 'POST'
    })
  }

  // Remove Roles
  for (const roleId of toRemove) {
    await useAdminFetch(`/admin/api/users/${userId}/roles/${roleId}`, {
      method: 'DELETE'
    })
  }
}

// Reset password
const resetPassword = async (user: any) => {
  if (!confirm(`Generate password reset token for ${user.name}?`)) return
  
  try {
    const { data } = await useAdminFetch<any>(`/admin/api/users/${user.id}/reset-password`, {
      method: 'POST'
    })
    
    if (data.value) {
      alert(`Reset token: ${data.value.token}`)
    }
  } catch (error) {
    alert('Failed to reset password')
  }
}

// Unlock user
const unlockUser = async (user: any) => {
  try {
    const { error } = await useAdminFetch(`/admin/api/users/${user.id}/unlock`, {
      method: 'POST'
    })
    
    if (!error.value) {
      alert('Account unlocked')
      await loadUsers()
    }
  } catch (error) {
    alert('Failed to unlock account')
  }
}

// Deactivate user
const deactivateUser = async (user: any) => {
  if (!confirm(`Are you sure you want to deactivate ${user.name}?`)) return
  
  try {
    const { error } = await useAdminFetch(`/admin/api/users/${user.id}`, {
      method: 'DELETE'
    })
    
    if (!error.value) {
      alert('User deactivated')
      await loadUsers()
    }
  } catch (error) {
    alert('Failed to deactivate user')
  }
}

// Change page
const changePage = (page: number) => {
  currentPage.value = page
  loadUsers()
}

// Change limit
const changeLimit = () => {
  currentPage.value = 1
  loadUsers()
}

// Load users on mount
onMounted(() => {
  loadUsers()
})

useHead({
  title: 'Users'
})
</script>
