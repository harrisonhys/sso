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
        <h1 class="text-2xl font-bold text-gray-900">Forgot Password?</h1>
        <p class="text-gray-600 mt-2">
          {{ emailSent ? 'Check your email' : 'Enter your email to reset your password' }}
        </p>
      </div>

      <!-- Success Message -->
      <div v-if="emailSent" class="mb-6 p-4 bg-green-50 border border-green-200 rounded-lg">
        <p class="text-sm text-green-800">
          We've sent a password reset link to <strong>{{ form.email }}</strong>.
          Please check your email and follow the instructions.
        </p>
      </div>

      <!-- Error Message -->
      <div v-if="errorMessage" class="mb-4 p-4 bg-red-50 border border-red-200 rounded-lg">
        <p class="text-sm text-red-800">{{ errorMessage }}</p>
      </div>

      <!-- Form -->
      <form v-if="!emailSent" @submit.prevent="handleSubmit" class="space-y-4">
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
            Sending...
          </span>
          <span v-else>Send Reset Link</span>
        </button>
      </form>

      <!-- Back to Login -->
      <div class="mt-6 text-center">
        <NuxtLink to="/login" class="text-sm text-primary-600 hover:text-primary-700 font-medium">
          ← Back to login
        </NuxtLink>
      </div>

      <!-- Resend Link -->
      <div v-if="emailSent" class="mt-4 text-center">
        <button
          @click="resendEmail"
          class="text-sm text-gray-600 hover:text-gray-800"
          :disabled="loading"
        >
          Didn't receive the email? Click to resend
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const config = useRuntimeConfig()

const form = ref({
  email: ''
})

const loading = ref(false)
const errorMessage = ref('')
const emailSent = ref(false)

const handleSubmit = async () => {
  loading.value = true
  errorMessage.value = ''

  try {
    const response = await fetch(`${config.public.apiBase}/password/forgot`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        email: form.value.email
      })
    })

    const data = await response.json()

    if (!response.ok) {
      throw new Error(data.error || 'Failed to send reset email')
    }

    emailSent.value = true
  } catch (error: any) {
    errorMessage.value = error.message || 'An error occurred'
  } finally {
    loading.value = false
  }
}

const resendEmail = () => {
  emailSent.value = false
  handleSubmit()
}

useHead({
  title: 'Forgot Password - SSO System'
})
</script>
