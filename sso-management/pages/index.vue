<template>
  <NuxtLayout name="default" title="Dashboard" subtitle="Welcome back! Here's what's happening today">
    <div class="space-y-6 animate-fade-in">
      <!-- Welcome Banner -->
      <Card class="bg-gradient-to-r from-indigo-600 via-purple-600 to-pink-500 text-white">
        <div class="flex items-center justify-between">
          <div>
            <h2 class="text-2xl font-bold mb-2">{{ greeting }}, Admin! 👋</h2>
            <p class="text-indigo-100">{{ currentDate }}</p>
          </div>
          <div class="hidden md:block">
            <Icon name="chart" class="w-20 h-20 opacity-20" />
          </div>
        </div>
      </Card>

      <!-- Statistics Cards -->
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <StatCard
          label="Total Users"
          :value="stats.totalUsers"
          :change="12.5"
          gradient
          gradient-class="from-blue-500 to-cyan-600"
          icon-bg-class="bg-blue-100"
          icon-color-class="text-blue-600"
        >
          <template #icon>
            <Icon name="users" class="w-6 h-6 text-white" />
          </template>
        </StatCard>

        <StatCard
          label="Active Sessions"
          :value="stats.activeSessions"
          :change="8.2"
          gradient
          gradient-class="from-green-500 to-emerald-600"
          icon-bg-class="bg-green-100"
          icon-color-class="text-green-600"
        >
          <template #icon>
            <Icon name="lock" class="w-6 h-6 text-white" />
          </template>
        </StatCard>

        <StatCard
          label="OAuth2 Clients"
          :value="stats.oauth2Clients"
          :change="5.1"
          gradient
          gradient-class="from-purple-500 to-pink-600"
          icon-bg-class="bg-purple-100"
          icon-color-class="text-purple-600"
        >
          <template #icon>
            <Icon name="key" class="w-6 h-6 text-white" />
          </template>
        </StatCard>

        <StatCard
          label="Failed Logins (24h)"
          :value="stats.failedLogins"
          :change="-15.3"
          gradient
          gradient-class="from-red-500 to-rose-600"
          icon-bg-class="bg-red-100"
          icon-color-class="text-red-600"
        >
          <template #icon>
            <Icon name="exclamation" class="w-6 h-6 text-white" />
          </template>
        </StatCard>
      </div>

      <!-- Quick Actions -->
      <Card>
        <template #header>
          <div class="flex items-center justify-between">
            <h3 class="text-lg font-semibold text-gray-900">Quick Actions</h3>
            <Icon name="plus" class="w-5 h-5 text-gray-400" />
          </div>
        </template>
        
        <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
          <NuxtLink
            to="/users"
            class="flex items-center gap-4 p-4 rounded-lg border-2 border-gray-200 hover:border-indigo-500 hover:bg-indigo-50 transition-all duration-200 group"
          >
            <div class="w-12 h-12 bg-indigo-100 rounded-lg flex items-center justify-center group-hover:bg-indigo-200 transition-colors">
              <Icon name="users" class="w-6 h-6 text-indigo-600" />
            </div>
            <div>
              <p class="font-semibold text-gray-900">Create User</p>
              <p class="text-sm text-gray-500">Add new user account</p>
            </div>
          </NuxtLink>

          <NuxtLink
            to="/oauth2-clients"
            class="flex items-center gap-4 p-4 rounded-lg border-2 border-gray-200 hover:border-purple-500 hover:bg-purple-50 transition-all duration-200 group"
          >
            <div class="w-12 h-12 bg-purple-100 rounded-lg flex items-center justify-center group-hover:bg-purple-200 transition-colors">
              <Icon name="key" class="w-6 h-6 text-purple-600" />
            </div>
            <div>
              <p class="font-semibold text-gray-900">Register Client</p>
              <p class="text-sm text-gray-500">New OAuth2 client</p>
            </div>
          </NuxtLink>

          <NuxtLink
            to="/audit-logs"
            class="flex items-center gap-4 p-4 rounded-lg border-2 border-gray-200 hover:border-green-500 hover:bg-green-50 transition-all duration-200 group"
          >
            <div class="w-12 h-12 bg-green-100 rounded-lg flex items-center justify-center group-hover:bg-green-200 transition-colors">
              <Icon name="clipboard" class="w-6 h-6 text-green-600" />
            </div>
            <div>
              <p class="font-semibold text-gray-900">View Logs</p>
              <p class="text-sm text-gray-500">System audit logs</p>
            </div>
          </NuxtLink>
        </div>
      </Card>

      <!-- Recent Activity -->
      <Card>
        <template #header>
          <div class="flex items-center justify-between">
            <div>
              <h3 class="text-lg font-semibold text-gray-900">Recent Activity</h3>
              <p class="text-sm text-gray-500 mt-1">Latest system events and user actions</p>
            </div>
            <Button variant="secondary" size="sm" @click="refreshLogs">
              <Icon name="refresh" class="w-4 h-4" />
              Refresh
            </Button>
          </div>
        </template>

        <div v-if="loading" class="space-y-3">
          <div v-for="i in 5" :key="i" class="skeleton h-16 rounded-lg"></div>
        </div>

        <div v-else-if="recentLogs.length === 0" class="text-center py-12">
          <Icon name="clipboard" class="w-16 h-16 text-gray-300 mx-auto mb-4" />
          <p class="text-gray-500 font-medium">No recent activity</p>
          <p class="text-sm text-gray-400 mt-1">Activity will appear here as it happens</p>
        </div>

        <div v-else class="overflow-x-auto">
          <table class="table">
            <thead>
              <tr>
                <th>Time</th>
                <th>User</th>
                <th>Action</th>
                <th>Resource</th>
                <th>Status</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="log in recentLogs" :key="log.id" class="animate-slide-in-up">
                <td class="text-sm text-gray-600">
                  {{ formatDate(log.created_at) }}
                </td>
                <td>
                  <div class="flex items-center gap-2">
                    <div class="w-8 h-8 bg-gradient-to-br from-indigo-500 to-purple-500 rounded-full flex items-center justify-center text-white text-xs font-semibold">
                      {{ getUserInitials(log.User?.name || 'System') }}
                    </div>
                    <span class="font-medium text-gray-900">{{ log.User?.name || 'System' }}</span>
                  </div>
                </td>
                <td>
                  <span class="font-medium text-gray-900">{{ log.action }}</span>
                </td>
                <td>
                  <span class="text-sm text-gray-600">{{ log.resource }}</span>
                </td>
                <td>
                  <Badge
                    :variant="log.status === 'success' ? 'success' : 'danger'"
                    :pulse="log.status === 'success'"
                  >
                    {{ log.status || 'N/A' }}
                  </Badge>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <template #footer>
          <div class="flex justify-between items-center">
            <p class="text-sm text-gray-500">Showing {{ recentLogs.length }} recent activities</p>
            <NuxtLink to="/audit-logs" class="text-sm font-medium text-indigo-600 hover:text-indigo-700">
              View all logs →
            </NuxtLink>
          </div>
        </template>
      </Card>
    </div>
  </NuxtLayout>
</template>

<script setup lang="ts">
const config = useRuntimeConfig()
const loading = ref(true)

// Dashboard statistics
const stats = ref({
  totalUsers: 0,
  activeSessions: 0,
  oauth2Clients: 0,
  failedLogins: 0
})

// Recent audit logs
const recentLogs = ref<any[]>([])

// Greeting based on time
const greeting = computed(() => {
  const hour = new Date().getHours()
  if (hour < 12) return 'Good Morning'
  if (hour < 18) return 'Good Afternoon'
  return 'Good Evening'
})

// Current date
const currentDate = computed(() => {
  return new Date().toLocaleDateString('en-US', {
    weekday: 'long',
    year: 'numeric',
    month: 'long',
    day: 'numeric'
  })
})

// Format date helper
const formatDate = (dateString: string) => {
  const date = new Date(dateString)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const minutes = Math.floor(diff / 60000)
  const hours = Math.floor(diff / 3600000)
  const days = Math.floor(diff / 86400000)

  if (minutes < 1) return 'Just now'
  if (minutes < 60) return `${minutes}m ago`
  if (hours < 24) return `${hours}h ago`
  if (days < 7) return `${days}d ago`
  return date.toLocaleDateString()
}

// Get user initials
const getUserInitials = (name: string) => {
  return name
    .split(' ')
    .map(n => n[0])
    .join('')
    .toUpperCase()
    .slice(0, 2)
}

// Load dashboard data
const loadData = async () => {
  loading.value = true
  try {
    // Fetch statistics
    const statsRes = await fetch(`${config.public.apiBase}/admin/api/stats`, {
      credentials: 'include'
    })
    if (statsRes.ok) {
      const data = await statsRes.json()
      stats.value = {
        totalUsers: data.total_users || 0,
        activeSessions: data.active_sessions || 0,
        oauth2Clients: data.oauth2_clients || 0,
        failedLogins: data.failed_logins_24h || 0
      }
    } else if (statsRes.status === 401) {
      localStorage.removeItem('user')
      navigateTo('/login')
    }

    // Fetch recent logs
    const logsRes = await fetch(`${config.public.apiBase}/admin/api/audit-logs?limit=10`, {
      credentials: 'include'
    })
    if (logsRes.ok) {
      const data = await logsRes.json()
      recentLogs.value = data.logs || []
    }
  } catch (error) {
    console.error('Failed to load dashboard data:', error)
  } finally {
    loading.value = false
  }
}

const refreshLogs = () => {
  loadData()
}

onMounted(() => {
  loadData()
})

useHead({
  title: 'Dashboard - SSO Management'
})
</script>
