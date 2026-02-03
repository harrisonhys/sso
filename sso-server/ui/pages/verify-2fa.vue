<template>
  <div class="min-h-screen bg-gradient-to-br from-primary-50 via-white to-primary-100 flex items-center justify-center p-4">
    <div class="card w-full max-w-md p-8">
      <!-- Header -->
      <div class="text-center mb-8">
        <div class="inline-flex items-center justify-center w-16 h-16 bg-primary-600 rounded-full mb-4">
          <svg class="w-8 h-8 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
          </svg>
        </div>
        <h1 class="text-2xl font-bold text-gray-900">Two-Factor Authentication</h1>
        <p class="text-gray-600 mt-2">Enter the 6-digit code from your authenticator app</p>
      </div>

      <!-- Error Message -->
      <div v-if="errorMessage" class="mb-4 p-4 bg-red-50 border border-red-200 rounded-lg">
        <p class="text-sm text-red-800">{{ errorMessage }}</p>
      </div>

      <!-- 2FA Form -->
      <form @submit.prevent="handleVerify" class="space-y-6">
        <div>
          <label for="code" class="block text-sm font-medium text-gray-700 mb-2 text-center">
            Verification Code
          </label>
          <input
            id="code"
            v-model="code"
            type="text"
            required
            maxlength="6"
            pattern="[0-9]{6}"
            class="input text-center text-2xl font-mono tracking-widest"
            placeholder="000000"
            :disabled="loading"
            autofocus
          >
          <p class="text-xs text-gray-500 mt-2 text-center">
            Enter the 6-digit code from your authenticator app
          </p>
        </div>

        <button
          type="submit"
          class="btn btn-primary w-full"
          :disabled="loading || code.length !== 6"
        >
          <span v-if="loading" class="flex items-center justify-center">
            <svg class="animate-spin -ml-1 mr-3 h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
            Verifying...
          </span>
          <span v-else>Verify</span>
        </button>

        <div class="text-center">
          <button
            type="button"
            @click="useBackupCode = !useBackupCode"
            class="text-sm text-primary-600 hover:text-primary-700 font-medium"
          >
            {{ useBackupCode ? 'Use authenticator code' : 'Use backup code instead' }}
          </button>
        </div>

        <div class="text-center">
          <NuxtLink to="/login" class="text-sm text-gray-600 hover:text-gray-800">
            ← Back to login
          </NuxtLink>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
const config = useRuntimeConfig()
const router = useRouter()

const code = ref('')
const loading = ref(false)
const errorMessage = ref('')
const useBackupCode = ref(false)

// Get temp token from session storage
const tempToken = ref('')

onMounted(() => {
  const token = sessionStorage.getItem('temp_token')
  if (!token) {
    // No temp token, redirect to login
    router.push('/login')
    return
  }
  tempToken.value = token
})

// Handle verification
const handleVerify = async () => {
  loading.value = true
  errorMessage.value = ''

  try {
    const response = await fetch(`${config.public.apiBase}/auth/verify-2fa`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        temp_token: tempToken.value,
        code: code.value
      }),
      credentials: 'include'
    })

    const data = await response.json()

    if (!response.ok) {
      throw new Error(data.error || 'Verification failed')
    }

    if (data.success) {
      // Clear temp token
      sessionStorage.removeItem('temp_token')
      sessionStorage.removeItem('user_email')

      // Store session token if provided
      if (data.session_token) {
        sessionStorage.setItem('session_token', data.session_token)
      }

      // Redirect to callback URL or default
      const redirectUrl = new URLSearchParams(window.location.search).get('redirect_uri')
      if (redirectUrl) {
        window.location.href = redirectUrl
      } else {
        window.location.href = '/'
      }
    }
  } catch (error: any) {
    errorMessage.value = error.message || 'An error occurred during verification'
    code.value = '' // Clear code on error
  } finally {
    loading.value = false
  }
}

// Auto-submit when 6 digits entered
watch(code, (newValue) => {
  if (newValue.length === 6 && !loading.value) {
    handleVerify()
  }
})

useHead({
  title: '2FA Verification - SSO System'
})
</script>
