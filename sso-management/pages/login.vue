<template>
  <div class="min-h-screen bg-gradient-to-br from-primary-600 to-primary-800 flex items-center justify-center p-4">
    <div class="max-w-md w-full">
      <!-- Login Card -->
      <div class="bg-white rounded-2xl shadow-2xl p-8">
        <!-- Logo/Title -->
        <div class="text-center mb-8">
          <h1 class="text-3xl font-bold text-gray-900">SSO Admin</h1>
          <p class="text-gray-600 mt-2">Management Dashboard</p>
        </div>

        <!-- Login Form -->
        <form @submit.prevent="handleLogin">
          <!-- Email -->
          <div class="mb-4">
            <label for="email" class="block text-sm font-medium text-gray-700 mb-2">
              Email Address
            </label>
            <input
              id="email"
              v-model="email"
              type="email"
              required
              class="w-full px-4 py-3 rounded-lg border border-gray-300 focus:ring-2 focus:ring-primary-500 focus:border-transparent"
              placeholder="admin@sso-system.com"
            />
          </div>

          <!-- Password -->
          <div class="mb-6">
            <label for="password" class="block text-sm font-medium text-gray-700 mb-2">
              Password
            </label>
            <input
              id="password"
              v-model="password"
              type="password"
              required
              class="w-full px-4 py-3 rounded-lg border border-gray-300 focus:ring-2 focus:ring-primary-500 focus:border-transparent"
              placeholder="••••••••"
            />
          </div>

          <!-- Error Message -->
          <div v-if="errorMessage" class="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg">
            <p class="text-sm text-red-800">{{ errorMessage }}</p>
          </div>

          <!-- Login Button -->
          <button
            type="submit"
            :disabled="loading"
            class="w-full bg-primary-600 hover:bg-primary-700 text-white font-semibold py-3 px-4 rounded-lg 
                   transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <span v-if="loading">Logging in...</span>
            <span v-else>Login</span>
          </button>
        </form>

        <!-- Footer -->
        <div class="mt-6 text-center text-sm text-gray-600">
          <p>Default credentials:</p>
          <p class="text-xs mt-1 font-mono text-gray-500">admin@sso-system.com / Admin123!@#</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
const config = useRuntimeConfig()
const router = useRouter()

const email = ref('admin@sso-system.com')
const password = ref('Admin123!@#')
const loading = ref(false)
const errorMessage = ref('')

const handleLogin = async () => {
  loading.value = true
  errorMessage.value = ''

  try {
    const response = await fetch(`${config.public.apiBase}/auth/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include', // Important: allows cookies to be set
      body: JSON.stringify({
        email: email.value,
        password: password.value,
      }),
    })

    const data = await response.json()

    if (response.ok && data.success) {
      // Session cookie is automatically set by backend
      // Optionally store user info in localStorage
      if (data.user) {
        localStorage.setItem('user', JSON.stringify(data.user))
      }

      // Redirect to dashboard
      await router.push('/')
    } else {
      errorMessage.value = data.error || 'Login failed. Please check your credentials.'
    }
  } catch (error) {
    console.error('Login error:', error)
    errorMessage.value = 'Network error. Please check if the backend is running.'
  } finally {
    loading.value = false
  }
}

// Check if already logged in
onMounted(() => {
  const user = localStorage.getItem('user')
  if (user) {
    router.push('/')
  }
})
</script>
