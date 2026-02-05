<template>
  <div class="min-h-screen bg-gradient-to-br from-blue-50 via-white to-purple-50 flex items-center justify-center p-4">
    <div class="bg-white rounded-2xl shadow-xl p-8 max-w-md w-full">
      <div class="text-center mb-8">
        <div class="w-16 h-16 bg-blue-100 rounded-full flex items-center justify-center mx-auto mb-4">
          <svg class="w-8 h-8 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
          </svg>
        </div>
        <h1 class="text-2xl font-bold text-gray-900 mb-2">Authorization Request</h1>
        <p class="text-gray-600">{{ clientName }} wants to access your account</p>
      </div>

      <div class="mb-6">
        <h2 class="text-sm font-semibold text-gray-700 mb-3">This application will be able to:</h2>
        <ul class="space-y-2">
          <li v-for="scope in scopes" :key="scope" class="flex items-start">
            <svg class="w-5 h-5 text-green-500 mr-2 mt-0.5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
            </svg>
            <span class="text-gray-700">{{ getScopeDescription(scope) }}</span>
          </li>
        </ul>
      </div>

      <form @submit.prevent="handleConsent" class="space-y-4">
        <input type="hidden" name="client_id" :value="clientId" />
        <input type="hidden" name="redirect_uri" :value="redirectUri" />
        <input type="hidden" name="scope" :value="scopeString" />
        <input type="hidden" name="state" :value="state" />
        <input type="hidden" name="code_challenge" :value="codeChallenge" />
        <input type="hidden" name="code_challenge_method" :value="codeChallengeMethod" />

        <div class="flex flex-col sm:flex-row gap-3">
          <button
            type="button"
            @click="handleDeny"
            class="flex-1 bg-gray-200 hover:bg-gray-300 text-gray-800 font-semibold py-3 px-6 rounded-lg transition-colors duration-200"
          >
            Deny
          </button>
          <button
            type="submit"
            class="flex-1 bg-blue-600 hover:bg-blue-700 text-white font-semibold py-3 px-6 rounded-lg transition-colors duration-200"
          >
            Allow
          </button>
        </div>
      </form>

      <p class="text-xs text-gray-500 text-center mt-6">
        By clicking "Allow", you authorize {{ clientName }} to access your information according to their privacy policy.
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()
const config = useRuntimeConfig()

const clientId = ref(route.query.client_id as string || '')
const redirectUri = ref(route.query.redirect_uri as string || '')
const scopeString = ref(route.query.scope as string || '')
const state = ref(route.query.state as string || '')
const codeChallenge = ref(route.query.code_challenge as string || '')
const codeChallengeMethod = ref(route.query.code_challenge_method as string || '')

const clientName = computed(() => {
  // You can fetch client name from API or use a mapping
  const clientNames: Record<string, string> = {
    'demo-client': 'Demo Client Application',
    'default': 'Application'
  }
  return clientNames[clientId.value] || clientNames['default']
})

const scopes = computed(() => {
  return scopeString.value.split(' ').filter(s => s.length > 0)
})

function getScopeDescription(scope: string): string {
  const descriptions: Record<string, string> = {
    'openid': 'Verify your identity',
    'profile': 'Access your basic profile information',
    'email': 'Access your email address',
    'offline_access': 'Access your data while you are offline'
  }
  return descriptions[scope] || `Access ${scope}`
}

async function handleConsent() {
  try {
    const formData = new URLSearchParams({
      client_id: clientId.value,
      redirect_uri: redirectUri.value,
      scope: scopeString.value,
      state: state.value,
      code_challenge: codeChallenge.value,
      code_challenge_method: codeChallengeMethod.value,
      approve: 'true'
    })

    const response = await fetch(`${config.public.apiBase}/oauth2/authorize/consent`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
      },
      body: formData.toString(),
      credentials: 'include', // Important for cookies
      redirect: 'manual' // Don't follow redirects automatically
    })

    // Check if it's a redirect (status 3xx)
    if (response.type === 'opaqueredirect' || response.status >= 300 && response.status < 400) {
      const redirectUrl = response.headers.get('Location')
      if (redirectUrl) {
        window.location.href = redirectUrl
      }
    } else if (response.ok) {
      // If response is OK, try to get redirect from response
      const data = await response.text()
      // Server might return redirect URL in response
      window.location.href = redirectUri.value
    } else {
      console.error('Consent failed:', response.statusText)
      alert('Failed to process consent. Please try again.')
    }
  } catch (error) {
    console.error('Error submitting consent:', error)
    alert('An error occurred. Please try again.')
  }
}

function handleDeny() {
  // Redirect back to client with error
  let redirectUrl = redirectUri.value + '?error=access_denied'
  if (state.value) {
    redirectUrl += '&state=' + state.value
  }
  window.location.href = redirectUrl
}

onMounted(() => {
  // Validate required parameters
  if (!clientId.value || !redirectUri.value) {
    alert('Invalid authorization request')
    router.push('/login')
  }
})
</script>
