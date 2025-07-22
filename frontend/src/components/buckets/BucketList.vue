<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between mb-6">
      <h2 class="text-2xl font-bold text-gray-800">Your Buckets</h2>
      <div class="flex space-x-3">
        <button
          @click="refreshBuckets"
          class="px-4 py-2 text-gray-600 bg-gray-50 border border-gray-200 rounded-md hover:bg-gray-100 hover:border-gray-300 focus:outline-none focus:ring-2 focus:ring-gray-500 focus:ring-offset-2 transition-colors"
        >
          Refresh
        </button>
        <button
          @click="showCreateModal = true"
          class="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700"
        >
          <svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
          </svg>
          Create Bucket
        </button>
      </div>
    </div>

    <!-- Error Message Toast -->
    <div v-if="error" class="fixed top-4 right-4 left-4 max-w-md mx-auto z-[9999] p-4 bg-red-50 border border-red-200 rounded-md shadow-lg">
      <div class="flex items-center">
        <svg class="w-5 h-5 text-red-400 mr-2 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <p class="text-sm text-red-600 flex-grow">{{ error }}</p>
        <button 
          @click="error = null" 
          class="ml-2 text-red-400 hover:text-red-600 flex-shrink-0"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>
    </div>

    <div v-if="loading" class="flex justify-center py-10">
      <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500"></div>
    </div>

    <div v-else-if="buckets.length === 0" class="text-center py-10 bg-gray-50 rounded-lg">
      <svg class="mx-auto text-gray-400 h-12 w-12" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
      </svg>
      <h3 class="mt-2 text-sm font-medium text-gray-900">No buckets</h3>
      <p class="mt-1 text-sm text-gray-500">Get started by creating a new bucket.</p>
      <div class="mt-6">
        <button
          @click="showCreateModal = true"
          class="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700"
        >
          <svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
          </svg>
          New Bucket
        </button>
      </div>
    </div>

    <div v-else class="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
      <div
        v-for="bucket in buckets"
        :key="bucket.name"
        class="bg-white rounded-lg shadow-md overflow-hidden hover:shadow-lg transition-shadow duration-300"
      >
        <div class="p-5">
          <div class="flex items-center">
            <svg class="text-blue-500 h-8 w-8 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
            </svg>
            <h3 class="text-lg font-medium text-gray-900 truncate" :title="bucket.name">
              {{ bucket.name }}
            </h3>
          </div>

          <div class="mt-3 text-sm text-gray-500">
            Created: {{ formatDate(bucket.creation_date) }}
          </div>

          <div class="mt-4 flex space-x-3">
            <button
              @click="selectBucket(bucket.name)"
              class="flex-1 inline-flex justify-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700"
            >
              Open
            </button>
            <button
              @click="confirmDelete(bucket.name)"
              class="px-3 inline-flex items-center border border-transparent text-sm font-medium rounded-md text-white bg-red-600 hover:bg-red-700"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
              </svg>
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Create Bucket Modal -->
    <div v-if="showCreateModal" class="fixed inset-0 z-50 overflow-y-auto" aria-labelledby="modal-title" role="dialog" aria-modal="true">
      <div class="flex items-end justify-center min-h-screen pt-4 px-4 pb-20 text-center sm:block sm:p-0">
        <!-- Background overlay -->
        <div class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" @click="showCreateModal = false"></div>

        <!-- This element centers the modal contents -->
        <span class="hidden sm:inline-block sm:align-middle sm:h-screen" aria-hidden="true">&#8203;</span>

        <!-- Modal panel -->
        <div
          class="inline-block align-bottom bg-white rounded-lg text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-lg sm:w-full"
          @click.stop
        >
          <div class="bg-white px-4 pt-5 pb-4 sm:p-6 sm:pb-4">
            <div class="sm:flex sm:items-start">
              <div class="mt-3 text-center sm:mt-0 sm:text-left w-full">
                <h3 class="text-lg leading-6 font-medium text-gray-900" id="modal-title">
                  Create New Bucket
                </h3>
                <div class="mt-4">
                  <form @submit.prevent="createBucket">
                    <div class="mt-4">
                      <label for="bucketName" class="block text-sm font-medium text-gray-700">Bucket Name</label>
                      <input
                        id="bucketName"
                        v-model="newBucketName"
                        placeholder="my-unique-bucket-name"
                        required
                        class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
                      />
                      <p class="mt-1 text-xs text-gray-500">
                        Bucket names must be unique across all AWS S3 or compatible services.
                      </p>
                    </div>
                  </form>
                </div>
              </div>
            </div>
          </div>
          <div class="bg-gray-50 px-4 py-3 sm:px-6 sm:flex sm:flex-row-reverse">
            <button
              type="button"
              class="w-full inline-flex justify-center rounded-md border border-transparent shadow-sm px-4 py-2 bg-blue-600 text-base font-medium text-white hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 sm:ml-3 sm:w-auto sm:text-sm"
              :disabled="!isValidBucketName"
              @click="createBucket"
            >
              Create
            </button>
            <button
              type="button"
              class="mt-3 w-full inline-flex justify-center rounded-md border border-gray-300 shadow-sm px-4 py-2 bg-white text-base font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 sm:mt-0 sm:ml-3 sm:w-auto sm:text-sm"
              @click="showCreateModal = false"
            >
              Cancel
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Delete Confirmation Modal -->
    <div v-if="bucketToDelete" class="fixed inset-0 z-50 overflow-y-auto" aria-labelledby="modal-title" role="dialog" aria-modal="true">
      <div class="flex items-end justify-center min-h-screen pt-4 px-4 pb-20 text-center sm:block sm:p-0">
        <!-- Background overlay -->
        <div class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" @click="bucketToDelete = null"></div>

        <!-- This element centers the modal contents -->
        <span class="hidden sm:inline-block sm:align-middle sm:h-screen" aria-hidden="true">&#8203;</span>

        <!-- Modal panel -->
        <div
          class="inline-block align-bottom bg-white rounded-lg text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-lg sm:w-full"
          @click.stop
        >
          <div class="bg-white px-4 pt-5 pb-4 sm:p-6 sm:pb-4">
            <div class="sm:flex sm:items-start">
              <div class="mt-3 text-center sm:mt-0 sm:ml-4 sm:text-left">
                <h3 class="text-lg leading-6 font-medium text-gray-900" id="modal-title">
                  Delete Bucket
                </h3>
                <div class="mt-2">
                  <p class="text-sm text-gray-500">
                    Are you sure you want to delete the bucket "{{ bucketToDelete }}"?
                  </p>
                  <p class="text-sm text-red-500 mt-2">This action cannot be undone.</p>
                </div>
              </div>
            </div>
          </div>
          <div class="bg-gray-50 px-4 py-3 sm:px-6 sm:flex sm:flex-row-reverse">
            <button
              type="button"
              class="w-full inline-flex justify-center rounded-md border border-transparent shadow-sm px-4 py-2 bg-red-600 text-base font-medium text-white hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 sm:ml-3 sm:w-auto sm:text-sm"
              @click="deleteBucket"
            >
              Delete
            </button>
            <button
              type="button"
              class="mt-3 w-full inline-flex justify-center rounded-md border border-gray-300 shadow-sm px-4 py-2 bg-white text-base font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 sm:mt-0 sm:ml-3 sm:w-auto sm:text-sm"
              @click="bucketToDelete = null"
            >
              Cancel
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useRouter } from 'vue-router';
// Removed Icon component dependency - using inline SVGs instead

// Define types locally
interface S3Bucket {
  name: string;
  creation_date: string;
}

const router = useRouter();
const buckets = ref<S3Bucket[]>([]);
const loading = ref(false);
const error = ref<string | null>(null);
const showCreateModal = ref(false);
const bucketToDelete = ref<string | null>(null);
const newBucketName = ref('');

// Initialize data
onMounted(async () => {
  await refreshBuckets();
});

const refreshBuckets = async () => {
  loading.value = true;
  error.value = null;

  try {
    const response = await fetch('/api/buckets', {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      }
    });

    if (!response.ok) {
      throw new Error(`Error fetching buckets: ${response.status} ${response.statusText}`);
    }

    buckets.value = await response.json();
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to load buckets';
    console.error('Error fetching buckets:', err);
  } finally {
    loading.value = false;
  }
};

const isValidBucketName = computed(() => {
  return newBucketName.value.length >= 3 &&
         newBucketName.value.length <= 63 &&
         /^[a-z0-9.-]+$/.test(newBucketName.value);
});

const createBucket = async () => {
  if (!isValidBucketName.value) return;

  loading.value = true;
  error.value = null;
  
  try {
    const response = await fetch(`/api/buckets/${encodeURIComponent(newBucketName.value)}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
      }
    });

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(errorText || `Failed to create bucket: ${response.status}`);
    }

    // Refresh the buckets list
    await refreshBuckets();
    showCreateModal.value = false;
    newBucketName.value = '';
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to create bucket';
    console.error('Error creating bucket:', err);
    // Auto-dismiss error after 8 seconds
    setTimeout(() => {
      error.value = null;
    }, 8000);
  } finally {
    loading.value = false;
  }
};

const selectBucket = (name: string) => {
  // Navigate to the objects view with bucket name
  router.push(`/objects/${encodeURIComponent(name)}`);
};

const confirmDelete = (bucketName: string) => {
  bucketToDelete.value = bucketName;
};

const deleteBucket = async () => {
  if (!bucketToDelete.value) return;

  loading.value = true;
  error.value = null;
  
  try {
    const response = await fetch(`/api/buckets/${encodeURIComponent(bucketToDelete.value)}`, {
      method: 'DELETE',
      headers: {
        'Content-Type': 'application/json',
      }
    });

    if (!response.ok) {
      let errorMessage = `Failed to delete bucket: ${response.status}`;
      try {
        const errorText = await response.text();
        // Try to parse JSON error response
        try {
          const errorData = JSON.parse(errorText);
          if (errorData.error) {
            errorMessage = errorData.error;
          } else if (errorData.message) {
            errorMessage = errorData.message;
          }
        } catch {
          // If not JSON, use the raw text
          if (errorText) {
            errorMessage = errorText;
          }
        }
      } catch {
        // Use default message if can't read response
      }
      throw new Error(errorMessage);
    }

    // Refresh the buckets list
    await refreshBuckets();
    bucketToDelete.value = null;
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to delete bucket';
    console.error('Error deleting bucket:', err);
    // Close the delete modal so error message is visible
    bucketToDelete.value = null;
    // Auto-dismiss error after 8 seconds
    setTimeout(() => {
      error.value = null;
    }, 8000);
  } finally {
    loading.value = false;
  }
};

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleString();
};
</script>
