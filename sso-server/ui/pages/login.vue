<template>
  <div class="min-h-screen bg-gradient-to-br from-primary-50 via-white to-primary-100 flex items-center justify-center p-4">
    <div class="card w-full max-w-md p-8">
      <!-- Logo/Header -->
      <div class="text-center mb-8">
        <div class="inline-flex items-center justify-center w-16 h-16 bg-primary-600 rounded-full mb-4">
          <svg class="w-8 h-8 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
          </svg>
        </div>
        <h1 class="text-2xl font-bold text-gray-900">Welcome Back</h1>
        <p class="text-gray-600 mt-2">Sign in to your account</p>
      </div>

      <!-- Error Message -->
      <div v-if="errorMessage" class="mb-4 p-4 bg-red-50 border border-red-200 rounded-lg">
        <p class="text-sm text-red-800">{{ errorMessage }}</p>
      </div>

      <!-- Login Form -->
      <form @submit.prevent="handleLogin" class="space-y-4">
        <div>
          <label for="email" class="block text-sm font-medium text-gray-700 mb-1">
            Email Address
          </label>
          <input
            id="email"
            v-model="form.email"
            type="email"
            required
            autocomplete="email"
            class="input"
            placeholder="you@example.com"
            :disabled="loading"
          >
        </div>

        <div>
          <label for="password" class="block text-sm font-medium text-gray-700 mb-1">
            Password
          </label>
          <input
            id="password"
            v-model="form.password"
            type="password"
            required
            autocomplete="current-password"
            class="input"
            placeholder="••••••••"
            :disabled="loading"
          >
        </div>

        <div class="flex items-center justify-between">
          <label class="flex items-center">
            <input
              v-model="form.remember"
              type="checkbox"
              class="rounded border-gray-300 text-primary-600 focus:ring-primary-500"
            >
            <span class="ml-2 text-sm text-gray-600">Remember me</span>
          </label>

          <NuxtLink to="/forgot-password" class="text-sm text-primary-600 hover:text-primary-700 font-medium">
            Forgot password?
          </NuxtLink>
        </div>

        <button
          type="submit"
          class="btn btn-primary w-full"
          :disabled="loading"
        >
          <span v-if="loading" class="flex items-center justify-center">
            <svg class="animate-spin -ml-1 mr-3 h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
            Signing in...
          </span>
          <span v-else>Sign In</span>
        </button>
      </form>

      <!-- Footer -->
      <div class="mt-6 text-center text-sm text-gray-600">
        <p>Secured by SSO System</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const config = useRuntimeConfig()
const router = useRouter()

// Form state
const form = ref({
  email: '',
  password: '',
  remember: false
})

const loading = ref(false)
const errorMessage = ref('')

// Handle login
const handleLogin = async () => {
  loading.value = true
  errorMessage.value = ''

  try {
    const response = await fetch(`${config.public.apiBase}/auth/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        email: form.value.email,
        password: form.value.password
      }),
      credentials: 'include' // Important for session cookies
    })

    const data = await response.json()

    if (!response.ok) {
      throw new Error(data.error || 'Login failed')
    }

    // Check if 2FA is required
    if (data.requires_two_factor) {
      // Store temp token and redirect to 2FA page
      sessionStorage.setItem('temp_token', data.temp_token)
      sessionStorage.setItem('user_email', form.value.email)
      await router.push('/verify-2fa')
      return
    }

    // Login successful
    if (data.success) {
      // Store session token if provided
      if (data.session_token) {
        sessionStorage.setItem('session_token', data.session_token)
      }
      
      // Store user email for display on success page
      sessionStorage.setItem('user_email', form.value.email)

      // Redirect to callback URL or default
      const returnUrl = new URLSearchParams(window.location.search).get('return_url')
      if (returnUrl) {
        // Redirect back to the OAuth2 flow
        window.location.href = returnUrl
      } else {
        // Default redirect to success page
        await router.push('/success')
      }
    }
  } catch (error: any) {
    errorMessage.value = error.message || 'An error occurred during login'
  } finally {
    loading.value = false
  }
}

// Check if user is already logged in on page mount
onMounted(() => {
  const sessionToken = sessionStorage.getItem('session_token')
  if (sessionToken) {
    // User already logged in, redirect to success page
    router.push('/success')
  }
})

// Set page title
useHead({
  title: 'Login - SSO System'
})
</script>
