<template>
  <div class="min-h-screen bg-gradient-to-br from-blue-50 via-white to-purple-50 flex items-center justify-center">
    <div class="bg-white rounded-2xl shadow-xl p-8 max-w-md w-full mx-4">
      <div v-if="loading" class="text-center">
        <div class="animate-spin rounded-full h-16 w-16 border-b-4 border-blue-600 mx-auto mb-4"></div>
        <h2 class="text-2xl font-semibold text-gray-800 mb-2">Processing...</h2>
        <p class="text-gray-600">Completing authentication</p>
      </div>

      <div v-else-if="error" class="text-center">
        <div class="w-16 h-16 bg-red-100 rounded-full flex items-center justify-center mx-auto mb-4">
          <svg class="w-8 h-8 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </div>
        <h2 class="text-2xl font-semibold text-gray-800 mb-2">Authentication Failed</h2>
        <p class="text-gray-600 mb-6">{{ error }}</p>
        <button
          @click="goHome"
          class="bg-blue-600 hover:bg-blue-700 text-white font-semibold py-3 px-6 rounded-lg transition-all duration-200 w-full"
        >
          Return to Home
        </button>
      </div>

      <div v-else class="text-center">
        <div class="w-16 h-16 bg-green-100 rounded-full flex items-center justify-center mx-auto mb-4">
          <svg class="w-8 h-8 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
          </svg>
        </div>
        <h2 class="text-2xl font-semibold text-gray-800 mb-2">Success!</h2>
        <p class="text-gray-600 mb-6">Redirecting to home page...</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'

const router = useRouter()
const route = useRoute()
const config = useRuntimeConfig()

const loading = ref(true)
const error = ref('')

onMounted(async () => {
  const code = route.query.code as string
  const state = route.query.state as string
  const errorParam = route.query.error as string

  if (errorParam) {
    error.value = route.query.error_description as string || 'Authentication failed'
    loading.value = false
    return
  }

  if (!code) {
    error.value = 'No authorization code received'
    loading.value = false
    return
  }

  try {
    await exchangeCodeForToken(code)
    // Redirect to home after successful authentication
    setTimeout(() => {
      router.push('/')
    }, 1500)
  } catch (err: any) {
    error.value = err.message || 'Failed to complete authentication'
    loading.value = false
  }
})

async function exchangeCodeForToken(code: string) {
  const codeVerifier = localStorage.getItem('code_verifier')
  if (!codeVerifier) {
    throw new Error('Code verifier not found')
  }

  const response = await fetch(`${config.public.ssoServerUrl}/oauth2/token`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/x-www-form-urlencoded'
    },
    body: new URLSearchParams({
      grant_type: 'authorization_code',
      code: code,
      redirect_uri: config.public.redirectUri,
      client_id: config.public.clientId,
      code_verifier: codeVerifier
    })
  })

  if (!response.ok) {
    const errorData = await response.json()
    throw new Error(errorData.error_description || 'Token exchange failed')
  }

  const data = await response.json()
  
  // Store tokens
  localStorage.setItem('access_token', data.access_token)
  if (data.refresh_token) {
    localStorage.setItem('refresh_token', data.refresh_token)
  }
  
  // Clean up code verifier
  localStorage.removeItem('code_verifier')
  
  loading.value = false
}

function goHome() {
  router.push('/')
}
</script>
