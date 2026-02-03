<template>
  <NuxtLayout name="default" title="OAuth2 Clients">
    <div class="space-y-6">
      <!-- Page Header -->
      <div class="flex items-center justify-between">
        <div>
          <p class="text-sm text-gray-600">Manage OAuth2 client applications and credentials</p>
        </div>
        <button @click="showCreateModal = true" class="btn btn-primary">
          ➕ Register Client
        </button>
      </div>

      <!-- Clients Grid -->
      <div v-if="loading" class="text-center py-12 text-gray-500">
        Loading clients...
      </div>
      <div v-else-if="clients.length === 0" class="card p-12 text-center text-gray-500">
        No OAuth2 clients registered yet
      </div>
      <div v-else class="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div v-for="client in clients" :key="client.id" class="card p-6">
          <div class="flex items-start justify-between mb-4">
            <div class="flex-1">
              <h3 class="text-lg font-semibold text-gray-900">{{ client.name }}</h3>
              <p class="text-sm text-gray-600 mt-1">{{ client.description || 'No description' }}</p>
            </div>
            <span :class="[
              'badge',
              client.is_active ? 'badge-success' : 'badge-error'
            ]">
              {{ client.is_active ? 'Active' : 'Inactive' }}
            </span>
          </div>

          <div class="space-y-3 text-sm">
            <div>
              <span class="font-medium text-gray-700">Client ID:</span>
              <code class="ml-2 px-2 py-1 bg-gray-100 rounded text-xs">{{ client.client_id }}</code>
            </div>
            
            <div>
              <span class="font-medium text-gray-700">Grant Types:</span>
              <div class="flex flex-wrap gap-1 mt-1">
                <span v-for="grant in client.grant_types" :key="grant" class="badge bg-blue-100 text-blue-800">
                  {{ grant }}
                </span>
              </div>
            </div>

            <div>
              <span class="font-medium text-gray-700">Redirect URIs:</span>
              <ul class="mt-1 space-y-1">
                <li v-for="uri in client.redirect_uris" :key="uri" class="text-xs text-gray-600 truncate">
                  • {{ uri }}
                </li>
              </ul>
            </div>

            <div>
              <span class="font-medium text-gray-700">Scopes:</span>
              <div class="flex flex-wrap gap-1 mt-1">
                <span v-for="scope in client.allowed_scopes" :key="scope" class="badge bg-purple-100 text-purple-800">
                  {{ scope }}
                </span>
              </div>
            </div>
          </div>

          <div class="mt-4 pt-4 border-t border-gray-200 flex gap-2">
            <button @click="regenerateSecret(client)" class="btn btn-secondary text-sm flex-1">
              🔑 Regenerate Secret
            </button>
            <button @click="revokeClient(client)" class="btn btn-danger text-sm flex-1">
              🗑️ Revoke
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

    <!-- Create Client Modal -->
    <div v-if="showCreateModal" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4" @click.self="showCreateModal = false">
      <div class="bg-white rounded-xl shadow-xl max-w-2xl w-full max-h-[90vh] overflow-y-auto">
        <div class="p-6 border-b border-gray-200 flex items-center justify-between sticky top-0 bg-white">
          <h3 class="text-xl font-semibold text-gray-900">Register OAuth2 Client</h3>
          <button @click="showCreateModal = false" class="text-gray-400 hover:text-gray-600 text-2xl">&times;</button>
        </div>
        <form @submit.prevent="createClient" class="p-6 space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Name *</label>
            <input v-model="newClient.name" type="text" required class="input">
          </div>
          
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Description</label>
            <textarea v-model="newClient.description" rows="2" class="input"></textarea>
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Redirect URIs * (one per line)</label>
            <textarea v-model="redirectUrisText" rows="3" placeholder="https://app.example.com/callback" class="input"></textarea>
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-700 mb-2">Grant Types *</label>
            <div class="space-y-2">
              <label class="flex items-center">
                <input type="checkbox" v-model="newClient.grant_types" value="authorization_code" class="mr-2">
                <span class="text-sm">Authorization Code</span>
              </label>
              <label class="flex items-center">
                <input type="checkbox" v-model="newClient.grant_types" value="refresh_token" class="mr-2">
                <span class="text-sm">Refresh Token</span>
              </label>
              <label class="flex items-center">
                <input type="checkbox" v-model="newClient.grant_types" value="client_credentials" class="mr-2">
                <span class="text-sm">Client Credentials</span>
              </label>
            </div>
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-700 mb-2">Allowed Scopes *</label>
            <div class="space-y-2">
              <label class="flex items-center">
                <input type="checkbox" v-model="newClient.allowed_scopes" value="openid" class="mr-2">
                <span class="text-sm">openid</span>
              </label>
              <label class="flex items-center">
                <input type="checkbox" v-model="newClient.allowed_scopes" value="profile" class="mr-2">
                <span class="text-sm">profile</span>
              </label>
              <label class="flex items-center">
                <input type="checkbox" v-model="newClient.allowed_scopes" value="email" class="mr-2">
                <span class="text-sm">email</span>
              </label>
            </div>
          </div>

          <div>
            <label class="flex items-center">
              <input type="checkbox" v-model="newClient.is_public" class="mr-2">
              <span class="text-sm font-medium text-gray-700">Public Client (PKCE required)</span>
            </label>
          </div>

          <div class="flex gap-3 pt-4">
            <button type="button" @click="showCreateModal = false" class="btn btn-secondary flex-1">Cancel</button>
            <button type="submit" class="btn btn-primary flex-1">Register Client</button>
          </div>
        </form>
      </div>
    </div>

    <!-- Secret Display Modal -->
    <div v-if="showSecretModal" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4" @click.self="showSecretModal = false">
      <div class="bg-white rounded-xl shadow-xl max-w-lg w-full p-6">
        <h3 class="text-xl font-semibold text-gray-900 mb-4">Client Secret</h3>
        <div class="bg-yellow-50 border border-yellow-200 rounded-lg p-4 mb-4">
          <p class="text-sm text-yellow-800">⚠️ Save this secret now! It will not be shown again.</p>
        </div>
        <div class="bg-gray-100 p-4 rounded-lg mb-4">
          <code class="text-sm break-all">{{ clientSecret }}</code>
        </div>
        <button @click="showSecretModal = false" class="btn btn-primary w-full">I've Saved It</button>
      </div>
    </div>
  </NuxtLayout>
</template>

<script setup lang="ts">
const config = useRuntimeConfig()
const loading = ref(false)
const showCreateModal = ref(false)
const showSecretModal = ref(false)
const clientSecret = ref('')

const clients = ref<any[]>([])
const currentPage = ref(1)
const limit = ref(20)
const totalClients = ref(0)
const totalPages = computed(() => Math.ceil(totalClients.value / limit.value))

const redirectUrisText = ref('')

const newClient = ref({
  name: '',
  description: '',
  redirect_uris: [] as string[],
  allowed_scopes: [] as string[],
  grant_types: [] as string[],
  is_public: false
})

const loadClients = async () => {
  loading.value = true
  try {
    const { data } = await useAdminFetch<any>(`/admin/api/oauth2-clients?page=${currentPage.value}&limit=${limit.value}`)
    if (data.value) {
      clients.value = data.value.clients || []
      totalClients.value = data.value.total || 0
    }
  } catch (error) {
    console.error('Failed to load clients:', error)
  } finally {
    loading.value = false
  }
}

const changePage = (page: number) => {
  currentPage.value = page
  loadClients()
}

const changeLimit = () => {
  currentPage.value = 1
  loadClients()
}

const createClient = async () => {
  // Parse redirect URIs from textarea
  const uris = redirectUrisText.value.split('\n').map(u => u.trim()).filter(u => u)
  newClient.value.redirect_uris = uris

  try {
    const { data, error } = await useAdminFetch<any>(`/admin/oauth2/clients`, {
      method: 'POST',
      body: newClient.value
    })
    
    if (data.value) {
      clientSecret.value = data.value.client_secret
      showCreateModal.value = false
      showSecretModal.value = true
      
      // Reset form
      newClient.value = {
        name: '',
        description: '',
        redirect_uris: [],
        allowed_scopes: [],
        grant_types: [],
        is_public: false
      }
      redirectUrisText.value = ''
      
      await loadClients()
    } else {
      alert(error.value?.message || 'Failed to create client')
    }
  } catch (error) {
    console.error('Failed to create client:', error)
    alert('Failed to create client')
  }
}

const regenerateSecret = async (client: any) => {
  if (!confirm(`Regenerate secret for ${client.name}? Old secret will be invalidated.`)) return
  
  try {
    const { data } = await useAdminFetch<any>(`/admin/oauth2/clients/${client.client_id}/regenerate-secret`, {
      method: 'POST'
    })
    
    if (data.value) {
      clientSecret.value = data.value.client_secret
      showSecretModal.value = true
    }
  } catch (error) {
    alert('Failed to regenerate secret')
  }
}

const revokeClient = async (client: any) => {
  if (!confirm(`Revoke client ${client.name}?`)) return
  
  try {
    const { error } = await useAdminFetch(`/admin/oauth2/clients/${client.client_id}`, {
      method: 'DELETE'
    })
    
    if (!error.value) {
      alert('Client revoked')
      await loadClients()
    }
  } catch (error) {
    alert('Failed to revoke client')
  }
}

onMounted(loadClients)

useHead({
  title: 'OAuth2 Clients'
})
</script>
