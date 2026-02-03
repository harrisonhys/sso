<template>
  <div class="min-h-screen bg-gradient-to-br from-green-50 via-white to-green-100 flex items-center justify-center p-4">
    <div class="card w-full max-w-md p-8 text-center">
      <!-- Success Icon -->
      <div class="inline-flex items-center justify-center w-20 h-20 bg-green-600 rounded-full mb-6">
        <svg class="w-10 h-10 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
        </svg>
      </div>
      
      <h1 class="text-3xl font-bold text-gray-900 mb-2">Login Successful!</h1>
      <p class="text-gray-600 mb-8">You have successfully authenticated with SSO Server</p>
      
      <!-- User Info -->
      <div class="bg-gray-50 rounded-lg p-4 mb-6 text-left">
        <h2 class="text-sm font-medium text-gray-700 mb-2">Session Information</h2>
        <div class="space-y-1 text-sm text-gray-600">
          <p><span class="font-medium">Email:</span> {{ userEmail || 'Not available' }}</p>
          <p><span class="font-medium">Session:</span> Active</p>
          <p><span class="font-medium">Time:</span> {{ currentTime }}</p>
        </div>
      </div>
      
      <!-- Actions -->
      <div class="space-y-3">
        <p class="text-sm text-gray-600 mb-4">
          This is the SSO authentication server. In a production environment, you would be redirected to your application.
        </p>
        
        <button
          @click="logout"
          class="btn btn-secondary w-full"
        >
          Logout
        </button>
        
        <NuxtLink to="/login" class="block text-sm text-primary-600 hover:text-primary-700">
          Back to Login
        </NuxtLink>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const config = useRuntimeConfig()
const router = useRouter()

// Get user email from session storage
const userEmail = ref('')
const currentTime = ref('')

onMounted(() => {
  userEmail.value = sessionStorage.getItem('user_email') || ''
  currentTime.value = new Date().toLocaleString()
  
  // Check if user has session token
  const sessionToken = sessionStorage.getItem('session_token')
  if (!sessionToken) {
    // No session, redirect to login
    router.push('/login')
  }
})

const logout = async () => {
  try {
    await fetch(`${config.public.apiBase}/auth/logout`, {
      method: 'POST',
      credentials: 'include'
    })
  } catch (error) {
    console.error('Logout error:', error)
  } finally {
    // Clear session storage
    sessionStorage.clear()
    // Redirect to login
    router.push('/login')
  }
}

useHead({
  title: 'Login Successful - SSO System'
})
</script>
