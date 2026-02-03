<template>
  <div class="min-h-screen bg-gradient-to-br from-primary-50 via-white to-primary-100 flex items-center justify-center p-4">
    <div class="card w-full max-w-md p-8 text-center">
      <div class="inline-flex items-center justify-center w-16 h-16 bg-primary-600 rounded-full mb-4">
        <svg class="w-8 h-8 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
        </svg>
      </div>
      
      <h1 class="text-2xl font-bold text-gray-900 mb-2">SSO System</h1>
      <p class="text-gray-600 mb-6">Single Sign-On Authentication</p>
      
      <div class="space-y-3">
        <NuxtLink to="/login" class="btn btn-primary w-full block">
          Sign In
        </NuxtLink>
        
        <p class="text-sm text-gray-600">
          Secure authentication for all your applications
        </p>
      </div>
      
      <div class="mt-8 pt-6 border-t border-gray-200">
        <p class="text-xs text-gray-500">
          Protected by enterprise-grade security
        </p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
useHead({
  title: 'SSO System - Single Sign-On Authentication'
})
</script>
