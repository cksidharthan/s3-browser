<template>
  <nav v-if="showNavbar" class="bg-blue-600 text-white shadow-sm">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
      <div class="flex justify-between h-16">
        <div class="flex items-center">
          <div class="flex-shrink-0 flex items-center">
            <svg class="h-8 w-8 mr-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
            </svg>
            <h1 class="text-xl font-bold">S3 Browser</h1>
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
  </nav>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'

const router = useRouter()
const route = useRoute()

const showConnectionModal = ref(false)
const connectionInfo = ref<any>(null)
const forceUpdate = ref(0) // Trigger for manual reactivity

// Only show navbar when we have an active session
const showNavbar = computed(() => {
  // Include forceUpdate to make it reactive to manual updates
  forceUpdate.value // This makes it reactive
  // Check both the reactive value and sessionStorage directly for reliability
  const hasConnectionInfo = connectionInfo.value !== null
  const hasStoredInfo = sessionStorage.getItem('connectionInfo') !== null
  return hasConnectionInfo || hasStoredInfo
})

// Watch for route changes to update connection info
watch(() => route.path, () => {
  fetchConnectionDetails()
}, { immediate: true })

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
      connectionInfo.value = null
      forceUpdate.value++
      // Redirect to home page which will show connection form
      router.push('/')
    } else {
      console.error('Failed to logout')
    }
  } catch (err) {
    console.error('Error during logout:', err)
  }
}

// Fetch connection details
const fetchConnectionDetails = async () => {
  try {
    // First check if we have an active session
    const sessionResponse = await fetch('/api/session/status', {
      credentials: 'include'
    })
    
    if (!sessionResponse.ok) {
      return
    }
    
    const sessionData = await sessionResponse.json()
    if (!sessionData.hasSession) {
      connectionInfo.value = null
      return
    }

    // For now, we'll need to store connection details in session storage during connection
    // since the backend doesn't expose them via API
    const storedInfo = sessionStorage.getItem('connectionInfo')
    if (storedInfo) {
      connectionInfo.value = JSON.parse(storedInfo)
    } else {
      connectionInfo.value = null
    }
    
    // Trigger reactivity update
    forceUpdate.value++
  } catch (err) {
    console.error('Error fetching connection details:', err)
    connectionInfo.value = null
    forceUpdate.value++
  }
}

// Initialize on mount
onMounted(() => {
  fetchConnectionDetails()
})
</script>
