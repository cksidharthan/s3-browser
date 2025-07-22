<template>
  <div class="min-h-screen bg-gray-100">
    <header class="bg-white shadow">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex justify-between h-16">
          <div class="flex">
            <div class="flex-shrink-0 flex items-center">
              <span class="text-xl font-bold text-gray-800">S3 Manager</span>
            </div>
          </div>
          
          <div class="flex items-center space-x-4">
            <!-- Connection Details Button -->
            <button
              @click="showConnectionModal = true"
              class="inline-flex items-center px-3 py-2 border border-blue-400 rounded-md text-sm font-medium text-white bg-blue-500 hover:bg-blue-400 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-300 transition-colors"
            >
              <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              Connection Info
            </button>
            
            <!-- Logout Button -->
            <button
              @click="logout"
              class="inline-flex items-center px-3 py-2 border border-red-400 rounded-md text-sm font-medium text-white bg-red-500 hover:bg-red-400 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-300 transition-colors"
            >
              <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
              </svg>
              Logout
            </button>
          </div>
        </div>
      </div>
    </header>

    <main class="py-6 px-4 sm:px-6 lg:px-8 max-w-7xl mx-auto">
      <div class="mb-6 px-2">
        <h1 class="text-2xl font-bold">S3 Buckets</h1>
      </div>

      <div class="bg-white rounded-lg shadow p-6">
        <BucketList />
      </div>
    </main>
    
    <!-- Connection Details Modal -->
    <div v-if="showConnectionModal" class="fixed inset-0 z-50 overflow-y-auto">
      <div class="flex items-center justify-center min-h-screen pt-4 px-4 pb-20 text-center">
        <div class="fixed inset-0 bg-gray-500 bg-opacity-75" @click="showConnectionModal = false"></div>
        <div class="inline-block bg-white rounded-lg text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-md sm:w-full">
          <div class="bg-white px-4 pt-5 pb-4 sm:p-6">
            <div class="flex items-center mb-4">
              <svg class="w-6 h-6 text-blue-500 mr-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <h3 class="text-lg font-medium text-gray-900">Connection Details</h3>
            </div>
            
            <div v-if="connectionInfo" class="space-y-3">
              <div class="flex justify-between py-2 border-b">
                <span class="text-sm font-medium text-gray-500">Endpoint:</span>
                <span class="text-sm text-gray-900">{{ connectionInfo.endpoint }}</span>
              </div>
              <div class="flex justify-between py-2 border-b">
                <span class="text-sm font-medium text-gray-500">Region:</span>
                <span class="text-sm text-gray-900">{{ connectionInfo.region }}</span>
              </div>
              <div class="flex justify-between py-2 border-b">
                <span class="text-sm font-medium text-gray-500">Access Key:</span>
                <span class="text-sm text-gray-900">{{ connectionInfo.access_key.substring(0, 8) }}...</span>
              </div>
              <div class="flex justify-between py-2">
                <span class="text-sm font-medium text-gray-500">SSL:</span>
                <span class="text-sm text-gray-900">{{ connectionInfo.use_ssl ? 'Enabled' : 'Disabled' }}</span>
              </div>
            </div>
            
            <div v-else class="text-center py-4">
              <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500 mx-auto"></div>
              <p class="text-sm text-gray-500 mt-2">Loading connection details...</p>
            </div>
          </div>
          
          <div class="bg-gray-50 px-4 py-3 sm:px-6 sm:flex sm:flex-row-reverse">
            <button
              @click="showConnectionModal = false"
              class="w-full inline-flex justify-center rounded-md border border-gray-300 shadow-sm px-4 py-2 bg-white text-base font-medium text-gray-700 hover:bg-gray-50 sm:mt-0 sm:w-auto sm:text-sm"
            >
              Close
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import BucketList from '@/components/buckets/BucketList.vue'

const router = useRouter()
const showConnectionModal = ref(false)
const connectionInfo = ref<any>(null)

// Logout function
const logout = async () => {
  try {
    const response = await fetch('/api/logout', {
      method: 'POST',
      credentials: 'include'
    })

    if (response.ok) {
      // Clear stored connection info
      sessionStorage.removeItem('connectionInfo')
      // Force page reload to ensure clean state
      window.location.href = '/'
    } else {
      console.error('Failed to logout')
    }
  } catch (err) {
    console.error('Error during logout:', err)
  }
}

// Fetch connection details
const fetchConnectionDetails = () => {
  try {
    const storedInfo = sessionStorage.getItem('connectionInfo')
    if (storedInfo) {
      connectionInfo.value = JSON.parse(storedInfo)
    }
  } catch (err) {
    console.error('Error fetching connection details:', err)
  }
}

// Load connection details on mount
onMounted(() => {
  fetchConnectionDetails()
})
</script>
