<template>
  <div class="min-h-screen bg-gradient-to-br from-primary-50 via-white to-primary-100 flex items-center justify-center p-4">
    <div class="card w-full max-w-md p-8">
      <!-- Header -->
      <div class="text-center mb-8">
        <div class="inline-flex items-center justify-center w-16 h-16 bg-primary-600 rounded-full mb-4">
          <svg class="w-8 h-8 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z" />
          </svg>
        </div>
        <h1 class="text-2xl font-bold text-gray-900">Reset Password</h1>
        <p class="text-gray-600 mt-2">Enter your new password</p>
      </div>

      <!-- Success Message -->
      <div v-if="resetSuccess" class="mb-6 p-4 bg-green-50 border border-green-200 rounded-lg">
        <p class="text-sm text-green-800">
          Your password has been reset successfully!
        </p>
        <NuxtLink to="/login" class="text-sm text-primary-600 hover:text-primary-700 font-medium mt-2 inline-block">
          Click here to login →
        </NuxtLink>
      </div>

      <!-- Error Message -->
      <div v-if="errorMessage" class="mb-4 p-4 bg-red-50 border border-red-200 rounded-lg">
        <p class="text-sm text-red-800">{{ errorMessage }}</p>
      </div>

      <!-- Form -->
      <form v-if="!resetSuccess" @submit.prevent="handleSubmit" class="space-y-4">
        <div>
          <label for="password" class="block text-sm font-medium text-gray-700 mb-1">
            New Password
          </label>
          <input
            id="password"
            v-model="form.password"
            type="password"
            required
            minlength="8"
            class="input"
            placeholder="••••••••"
            :disabled="loading"
          >
          <p class="text-xs text-gray-500 mt-1">
            Minimum 8 characters, include uppercase, lowercase, number, and special character
          </p>
        </div>

        <div>
          <label for="confirm_password" class="block text-sm font-medium text-gray-700 mb-1">
            Confirm New Password
          </label>
          <input
            id="confirm_password"
            v-model="form.confirmPassword"
            type="password"
            required
            class="input"
            placeholder="••••••••"
            :disabled="loading"
          >
        </div>

        <!-- Password Strength Indicator -->
        <div v-if="form.password" class="space-y-2">
          <div class="flex items-center justify-between text-xs">
            <span class="text-gray-600">Password Strength:</span>
            <span :class="passwordStrengthColor">{{ passwordStrengthText }}</span>
          </div>
          <div class="h-2 bg-gray-200 rounded-full overflow-hidden">
            <div
              :class="passwordStrengthColor"
              :style="{ width: `${passwordStrength}%` }"
              class="h-full transition-all duration-300"
            ></div>
          </div>
        </div>

        <!-- Password Match Indicator -->
        <div v-if="form.confirmPassword" class="text-xs">
          <span v-if="passwordsMatch" class="text-green-600">✓ Passwords match</span>
          <span v-else class="text-red-600">✗ Passwords do not match</span>
        </div>

        <button
          type="submit"
          class="btn btn-primary w-full"
          :disabled="loading || !passwordsMatch || !form.password"
        >
          <span v-if="loading" class="flex items-center justify-center">
            <svg class="animate-spin -ml-1 mr-3 h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
            Resetting...
          </span>
          <span v-else>Reset Password</span>
        </button>
      </form>

      <!-- Back to Login -->
      <div class="mt-6 text-center">
        <NuxtLink to="/login" class="text-sm text-primary-600 hover:text-primary-700 font-medium">
          ← Back to login
        </NuxtLink>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const config = useRuntimeConfig()
const route = useRoute()

const form = ref({
  password: '',
  confirmPassword: ''
})

const loading = ref(false)
const errorMessage = ref('')
const resetSuccess = ref(false)

// Get token from URL
const token = computed(() => route.query.token as string)

// Password validation
const passwordsMatch = computed(() => {
  return form.value.password === form.value.confirmPassword && form.value.confirmPassword.length > 0
})

const passwordStrength = computed(() => {
  const password = form.value.password
  let strength = 0
  
  if (password.length >= 8) strength += 25
  if (password.length >= 12) strength += 15
  if (/[a-z]/.test(password)) strength += 15
  if (/[A-Z]/.test(password)) strength += 15
  if (/[0-9]/.test(password)) strength += 15
  if (/[^a-zA-Z0-9]/.test(password)) strength += 15
  
  return Math.min(strength, 100)
})

const passwordStrengthText = computed(() => {
  if (passwordStrength.value < 40) return 'Weak'
  if (passwordStrength.value < 70) return 'Medium'
  return 'Strong'
})

const passwordStrengthColor = computed(() => {
  if (passwordStrength.value < 40) return 'bg-red-500 text-red-600'
  if (passwordStrength.value < 70) return 'bg-yellow-500 text-yellow-600'
  return 'bg-green-500 text-green-600'
})

const handleSubmit = async () => {
  if (!token.value) {
    errorMessage.value = 'Invalid or missing reset token'
    return
  }

  if (!passwordsMatch.value) {
    errorMessage.value = 'Passwords do not match'
    return
  }

  loading.value = true
  errorMessage.value = ''

  try {
    const response = await fetch(`${config.public.apiBase}/password/reset`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        token: token.value,
        new_password: form.value.password
      })
    })

    const data = await response.json()

    if (!response.ok) {
      throw new Error(data.error || 'Failed to reset password')
    }

    resetSuccess.value = true
  } catch (error: any) {
    errorMessage.value = error.message || 'An error occurred'
  } finally {
    loading.value = false
  }
}

// Check if token exists on mount
onMounted(() => {
  if (!token.value) {
    errorMessage.value = 'Invalid or missing reset token. Please request a new password reset link.'
  }
})

useHead({
  title: 'Reset Password - SSO System'
})
</script>
