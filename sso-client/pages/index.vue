<template>
  <div class="min-h-screen bg-gradient-to-br from-blue-50 via-white to-purple-50">
    <div class="container mx-auto px-4 py-16">
      <!-- Header -->
      <div class="text-center mb-12">
        <h1 class="text-5xl font-bold text-gray-900 mb-4">
          SSO Client Demo
        </h1>
        <p class="text-xl text-gray-600">
          Single Sign-On Integration Example
        </p>
      </div>

      <!-- Main Content -->
      <div class="max-w-4xl mx-auto">
        <!-- Status Card -->
        <div class="bg-white rounded-2xl shadow-xl p-8 mb-8">
          <div class="flex items-center justify-between mb-6">
            <h2 class="text-2xl font-semibold text-gray-800">
              Authentication Status
            </h2>
            <div v-if="isAuthenticated" class="flex items-center space-x-2">
              <div class="w-3 h-3 bg-green-500 rounded-full animate-pulse"></div>
              <span class="text-green-600 font-medium">Authenticated</span>
            </div>
            <div v-else class="flex items-center space-x-2">
              <div class="w-3 h-3 bg-gray-400 rounded-full"></div>
              <span class="text-gray-500 font-medium">Not Authenticated</span>
            </div>
          </div>

          <!-- User Info (if authenticated) -->
          <div v-if="isAuthenticated && user" class="bg-gradient-to-r from-blue-50 to-purple-50 rounded-xl p-6 mb-6">
            <h3 class="text-lg font-semibold text-gray-800 mb-4">User Information</h3>
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <p class="text-sm text-gray-600">Name</p>
                <p class="text-lg font-medium text-gray-900">{{ user.name }}</p>
              </div>
              <div>
                <p class="text-sm text-gray-600">Email</p>
                <p class="text-lg font-medium text-gray-900">{{ user.email }}</p>
              </div>
              <div>
                <p class="text-sm text-gray-600">User ID</p>
                <p class="text-sm font-mono text-gray-700">{{ user.id }}</p>
              </div>
              <div>
                <p class="text-sm text-gray-600">Email Verified</p>
                <p class="text-lg">
                  <span v-if="user.email_verified" class="text-green-600">✓ Verified</span>
                  <span v-else class="text-yellow-600">⚠ Not Verified</span>
                </p>
              </div>
            </div>
          </div>

          <!-- Actions -->
          <div class="flex flex-col sm:flex-row gap-4">
            <button
              v-if="!isAuthenticated"
              @click="login"
              class="flex-1 bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700 text-white font-semibold py-3 px-6 rounded-lg transition-all duration-200 transform hover:scale-105 shadow-lg"
            >
              Sign in with SSO
            </button>
            <template v-else>
              <button
                @click="refreshToken"
                class="flex-1 bg-blue-600 hover:bg-blue-700 text-white font-semibold py-3 px-6 rounded-lg transition-all duration-200 shadow-md"
              >
                Refresh Token
              </button>
              <button
                @click="logout"
                class="flex-1 bg-red-600 hover:bg-red-700 text-white font-semibold py-3 px-6 rounded-lg transition-all duration-200 shadow-md"
              >
                Sign Out
              </button>
            </template>
          </div>
        </div>

        <!-- Access Token Card (if authenticated) -->
        <div v-if="isAuthenticated && accessToken" class="bg-white rounded-2xl shadow-xl p-8 mb-8">
          <h3 class="text-xl font-semibold text-gray-800 mb-4">Access Token</h3>
          <div class="bg-gray-50 rounded-lg p-4 font-mono text-sm text-gray-700 break-all">
            {{ accessToken }}
          </div>
          <button
            @click="copyToken"
            class="mt-4 bg-gray-200 hover:bg-gray-300 text-gray-800 font-medium py-2 px-4 rounded-lg transition-colors duration-200"
          >
            {{ copied ? '✓ Copied!' : 'Copy Token' }}
          </button>
        </div>

        <!-- Features Grid -->
        <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
          <div class="bg-white rounded-xl shadow-lg p-6 hover:shadow-xl transition-shadow duration-200">
            <div class="w-12 h-12 bg-blue-100 rounded-lg flex items-center justify-center mb-4">
              <svg class="w-6 h-6 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
              </svg>
            </div>
            <h3 class="text-lg font-semibold text-gray-800 mb-2">Secure Authentication</h3>
            <p class="text-gray-600">OAuth2 with PKCE for enhanced security</p>
          </div>

          <div class="bg-white rounded-xl shadow-lg p-6 hover:shadow-xl transition-shadow duration-200">
            <div class="w-12 h-12 bg-purple-100 rounded-lg flex items-center justify-center mb-4">
              <svg class="w-6 h-6 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
              </svg>
            </div>
            <h3 class="text-lg font-semibold text-gray-800 mb-2">2FA Support</h3>
            <p class="text-gray-600">Two-factor authentication with TOTP</p>
          </div>

          <div class="bg-white rounded-xl shadow-lg p-6 hover:shadow-xl transition-shadow duration-200">
            <div class="w-12 h-12 bg-green-100 rounded-lg flex items-center justify-center mb-4">
              <svg class="w-6 h-6 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
              </svg>
            </div>
            <h3 class="text-lg font-semibold text-gray-800 mb-2">Fast & Reliable</h3>
            <p class="text-gray-600">Built with modern web technologies</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'

const config = useRuntimeConfig()

const isAuthenticated = ref(false)
const user = ref<any>(null)
const accessToken = ref('')
const copied = ref(false)

console.log('SSO Server URL:', config.public.ssoServerUrl)

onMounted(() => {
  // Check if user is authenticated
  const token = localStorage.getItem('access_token')
  if (token) {
    accessToken.value = token
    isAuthenticated.value = true
    fetchUserInfo()
  }
})

function generateCodeVerifier(): string {
  const array = new Uint8Array(32)
  crypto.getRandomValues(array)
  return btoa(String.fromCharCode(...array))
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=/g, '')
}

async function generateCodeChallenge(verifier: string): Promise<string> {
  const encoder = new TextEncoder()
  const data = encoder.encode(verifier)
  const hash = await crypto.subtle.digest('SHA-256', data)
  return btoa(String.fromCharCode(...new Uint8Array(hash)))
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=/g, '')
}

async function login() {
  // Generate PKCE parameters
  const codeVerifier = generateCodeVerifier()
  const codeChallenge = await generateCodeChallenge(codeVerifier)
  
  // Store code verifier for later use
  localStorage.setItem('code_verifier', codeVerifier)
  
  // Build authorization URL
  const params = new URLSearchParams({
    client_id: config.public.clientId,
    redirect_uri: config.public.redirectUri,
    response_type: 'code',
    scope: 'openid profile email',
    code_challenge: codeChallenge,
    code_challenge_method: 'S256',
    state: Math.random().toString(36).substring(7)
  })
  
  const authUrl = `${config.public.ssoServerUrl}/oauth2/authorize?${params.toString()}`
  window.location.href = authUrl
}

async function fetchUserInfo() {
  try {
    const response = await fetch(`${config.public.ssoServerUrl}/oauth2/userinfo`, {
      headers: {
        'Authorization': `Bearer ${accessToken.value}`
      }
    })
    
    if (response.ok) {
      user.value = await response.json()
    } else {
      // Token might be expired
      logout()
    }
  } catch (error) {
    console.error('Failed to fetch user info:', error)
  }
}

async function refreshToken() {
  const refreshTokenValue = localStorage.getItem('refresh_token')
  if (!refreshTokenValue) {
    logout()
    return
  }

  try {
    const response = await fetch(`${config.public.ssoServerUrl}/oauth2/token`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded'
      },
      body: new URLSearchParams({
        grant_type: 'refresh_token',
        refresh_token: refreshTokenValue,
        client_id: config.public.clientId
      })
    })

    if (response.ok) {
      const data = await response.json()
      accessToken.value = data.access_token
      localStorage.setItem('access_token', data.access_token)
      if (data.refresh_token) {
        localStorage.setItem('refresh_token', data.refresh_token)
      }
      await fetchUserInfo()
    } else {
      logout()
    }
  } catch (error) {
    console.error('Failed to refresh token:', error)
    logout()
  }
}

function logout() {
  localStorage.removeItem('access_token')
  localStorage.removeItem('refresh_token')
  localStorage.removeItem('code_verifier')
  isAuthenticated.value = false
  user.value = null
  accessToken.value = ''
}

function copyToken() {
  navigator.clipboard.writeText(accessToken.value)
  copied.value = true
  setTimeout(() => {
    copied.value = false
  }, 2000)
}
</script>
