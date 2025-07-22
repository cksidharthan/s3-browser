<script setup lang="ts">
</script>

<template>
  <div class="min-h-screen bg-gray-50">
    <!-- Loading State -->
    <div v-if="loading" class="flex items-center justify-center min-h-screen">
      <div class="animate-spin rounded-full h-16 w-16 border-b-2 border-blue-600"></div>
    </div>

    <!-- Connection Form - Show when no session -->
    <div v-else-if="!hasSession" class="min-h-screen flex items-center justify-center py-12 px-4 sm:px-6 lg:px-8">
      <div class="max-w-md w-full space-y-8">
        <div>
          <h2 class="mt-6 text-center text-3xl font-extrabold text-gray-900">
            Connect to S3
          </h2>
          <p class="mt-2 text-center text-sm text-gray-600">
            Enter your S3 credentials to get started
          </p>
        </div>
        <form class="mt-8 space-y-6" @submit.prevent="handleConnect">
          <div class="rounded-md shadow-sm -space-y-px">
            <div>
              <label for="endpoint" class="sr-only">Endpoint</label>
              <input
                id="endpoint"
                v-model="form.endpoint"
                name="endpoint"
                type="text"
                required
                class="relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-t-md focus:outline-none focus:ring-blue-500 focus:border-blue-500 focus:z-10 sm:text-sm"
                placeholder="Endpoint (e.g., s3.amazonaws.com)"
              />
            </div>
            <div>
              <label for="region" class="sr-only">Region</label>
              <input
                id="region"
                v-model="form.region"
                name="region"
                type="text"
                required
                class="relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 focus:outline-none focus:ring-blue-500 focus:border-blue-500 focus:z-10 sm:text-sm"
                placeholder="Region (e.g., us-east-1)"
              />
            </div>
            <div>
              <label for="access-key" class="sr-only">Access Key</label>
              <input
                id="access-key"
                v-model="form.access_key"
                name="access_key"
                type="text"
                required
                class="relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 focus:outline-none focus:ring-blue-500 focus:border-blue-500 focus:z-10 sm:text-sm"
                placeholder="Access Key"
              />
            </div>
            <div>
              <label for="secret-key" class="sr-only">Secret Key</label>
              <input
                id="secret-key"
                v-model="form.secret_key"
                name="secret_key"
                type="password"
                required
                class="relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-b-md focus:outline-none focus:ring-blue-500 focus:border-blue-500 focus:z-10 sm:text-sm"
                placeholder="Secret Key"
              />
            </div>
          </div>

          <div class="flex items-center">
            <input
              id="use-ssl"
              v-model="form.use_ssl"
              name="use_ssl"
              type="checkbox"
              class="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
            />
            <label for="use-ssl" class="ml-2 block text-sm text-gray-900">
              Use SSL
            </label>
          </div>

          <div v-if="error" class="text-red-600 text-sm text-center">
            {{ error }}
          </div>

          <div>
            <button
              type="submit"
              :disabled="connecting"
              class="group relative w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <span v-if="connecting">Testing Connection...</span>
              <span v-else>Test Connection and Continue</span>
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- Buckets View - Show when session exists -->
    <div v-else>
      <BucketsView />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import BucketsView from './BucketsView.vue'

const loading = ref(true)
const hasSession = ref(false)
const connecting = ref(false)
const error = ref('')

const form = ref({
  endpoint: '',
  access_key: '',
  secret_key: '',
  region: '',
  use_ssl: true
})

const checkSession = async () => {
  try {
    const response = await fetch('/api/session/status')
    const data = await response.json()
    hasSession.value = data.has_session
  } catch (err) {
    console.error('Error checking session:', err)
    hasSession.value = false
  } finally {
    loading.value = false
  }
}

const handleConnect = async () => {
  connecting.value = true
  error.value = ''
  
  try {
    const response = await fetch('/api/connect', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(form.value)
    })
    
    const data = await response.json()
    
    if (data.success) {
      hasSession.value = true
      
      // Store connection info for navbar (excluding secret key for security)
      const connectionInfo = {
        endpoint: form.value.endpoint,
        region: form.value.region,
        access_key: form.value.access_key,
        use_ssl: form.value.use_ssl
      }
      sessionStorage.setItem('connectionInfo', JSON.stringify(connectionInfo))
      
      // Clear form data for security
      form.value = {
        endpoint: '',
        access_key: '',
        secret_key: '',
        region: '',
        use_ssl: true
      }
    } else {
      error.value = data.message || 'Connection failed'
    }
  } catch (err) {
    error.value = 'Network error occurred'
    console.error('Connection error:', err)
  } finally {
    connecting.value = false
  }
}

onMounted(() => {
  checkSession()
})
</script>
