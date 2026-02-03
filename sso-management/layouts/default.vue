<template>
  <div class="min-h-screen flex bg-gray-50">
    <!-- Sidebar -->
    <aside :class="['sidebar transition-all duration-300', sidebarCollapsed && 'sidebar-collapsed']">
      <!-- Logo -->
      <div class="p-6 border-b border-gray-200">
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 bg-gradient-to-br from-indigo-600 to-purple-600 rounded-lg flex items-center justify-center">
            <Icon name="shield" class="w-6 h-6 text-white" />
          </div>
          <div v-if="!sidebarCollapsed" class="flex-1">
            <h1 class="text-xl font-bold gradient-text">SSO Admin</h1>
            <p class="text-xs text-gray-500">Management Panel</p>
          </div>
        </div>
      </div>

      <!-- Navigation -->
      <nav class="flex-1 p-4 space-y-1 overflow-y-auto">
        <NuxtLink
          v-for="item in navItems"
          :key="item.to"
          :to="item.to"
          :class="['sidebar-link group', $route.path === item.to && 'sidebar-link-active']"
        >
          <Icon :name="item.icon" class="w-5 h-5 flex-shrink-0" />
          <span v-if="!sidebarCollapsed" class="font-medium">{{ item.label }}</span>
          <Badge
            v-if="item.badge && !sidebarCollapsed"
            :variant="item.badgeVariant"
            class="ml-auto"
          >
            {{ item.badge }}
          </Badge>
        </NuxtLink>
      </nav>

      <!-- User Section -->
      <div class="p-4 border-t border-gray-200">
        <div v-if="!sidebarCollapsed" class="flex items-center gap-3 mb-3">
          <div class="w-10 h-10 bg-gradient-to-br from-indigo-500 to-purple-500 rounded-full flex items-center justify-center text-white font-semibold">
            {{ userInitials }}
          </div>
          <div class="flex-1 min-w-0">
            <p class="text-sm font-medium text-gray-900 truncate">{{ userName }}</p>
            <p class="text-xs text-gray-500 truncate">{{ userEmail }}</p>
          </div>
        </div>
        <Button
          variant="secondary"
          size="sm"
          :class="['w-full', sidebarCollapsed && 'px-2']"
          @click="handleLogout"
        >
          <Icon name="logout" class="w-4 h-4" />
          <span v-if="!sidebarCollapsed">Logout</span>
        </Button>
      </div>
    </aside>

    <!-- Main Content -->
    <div class="flex-1 flex flex-col min-w-0">
      <!-- Header -->
      <header class="bg-white border-b border-gray-200 px-6 py-4 sticky top-0 z-10">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-4">
            <button
              @click="sidebarCollapsed = !sidebarCollapsed"
              class="p-2 hover:bg-gray-100 rounded-lg transition-colors lg:hidden"
            >
              <Icon name="menu" class="w-5 h-5" />
            </button>
            <div>
              <h2 class="text-2xl font-bold text-gray-900">{{ title }}</h2>
              <p v-if="subtitle" class="text-sm text-gray-500">{{ subtitle }}</p>
            </div>
          </div>
          
          <div class="flex items-center gap-3">
            <!-- Search -->
            <div class="hidden md:flex items-center gap-2 px-4 py-2 bg-gray-100 rounded-lg">
              <Icon name="search" class="w-4 h-4 text-gray-400" />
              <input
                type="text"
                placeholder="Search..."
                class="bg-transparent border-none outline-none text-sm w-64"
              />
            </div>
            
            <!-- Notifications -->
            <button class="relative p-2 hover:bg-gray-100 rounded-lg transition-colors">
              <Icon name="bell" class="w-5 h-5 text-gray-600" />
              <span class="absolute top-1 right-1 w-2 h-2 bg-red-500 rounded-full"></span>
            </button>
          </div>
        </div>
      </header>

      <!-- Page Content -->
      <main class="flex-1 p-6 overflow-auto">
        <slot />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

interface Props {
  title?: string
  subtitle?: string
}

const props = withDefaults(defineProps<Props>(), {
  title: 'Dashboard',
  subtitle: ''
})

const sidebarCollapsed = ref(false)
const userName = ref('Admin User')
const userEmail = ref('admin@sso.com')

const userInitials = computed(() => {
  return userName.value
    .split(' ')
    .map(n => n[0])
    .join('')
    .toUpperCase()
    .slice(0, 2)
})

const navItems = [
  { to: '/', label: 'Dashboard', icon: 'chart' },
  { to: '/users', label: 'Users', icon: 'users' },
  { to: '/roles', label: 'Roles', icon: 'shield' },
  { to: '/permissions', label: 'Permissions', icon: 'lock' },
  { to: '/oauth2-clients', label: 'OAuth2 Clients', icon: 'key' },
  { to: '/audit-logs', label: 'Audit Logs', icon: 'clipboard' },
  { to: '/settings', label: 'Settings', icon: 'cog' },
]

const handleLogout = async () => {
  try {
    // Call backend to invalidate session cookie
    await useAdminFetch('/auth/logout', { method: 'POST' })
  } catch (error) {
    console.error('Logout failed:', error)
  } finally {
    // Clear local state and redirect
    localStorage.removeItem('user')
    navigateTo('/login')
  }
}
</script>
