<template>
  <div class="space-y-6">
    <div class="flex justify-between items-center">
      <div>
        <h2 class="text-xl font-bold text-gray-800">Objects</h2>
        <p class="text-sm text-gray-500 mt-1">
          {{ objects.length }} objects
        </p>
      </div>

      <div class="flex space-x-4">
        <button
          @click="refreshObjects"
          class="inline-flex items-center px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
        >
          <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
          Refresh
        </button>
        <button
          @click="showUploadModal = true"
          class="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
        >
          <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
          </svg>
          Upload
        </button>
      </div>
    </div>

    <div v-if="error" class="rounded-md bg-red-50 p-4">
      <div class="flex">
        <div class="flex-shrink-0">
          <svg class="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
            <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd" />
          </svg>
        </div>
        <div class="ml-3">
          <p class="text-sm text-red-800">{{ error }}</p>
        </div>
      </div>
    </div>

    <div v-if="loading" class="flex justify-center py-10">
      <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500"></div>
    </div>

    <div v-else-if="objects.length === 0" class="text-center py-10 bg-gray-50 rounded-lg">
      <svg class="mx-auto h-12 w-12 text-gray-400" stroke="currentColor" fill="none" viewBox="0 0 48 48">
        <path d="M28 8H12a4 4 0 00-4 4v20m32-12v8m0 0v8a4 4 0 01-4 4H12a4 4 0 01-4-4v-4m32-4l-3.172-3.172a4 4 0 00-5.656 0L28 28M8 32l9.172-9.172a4 4 0 015.656 0L28 28m0 0l4 4m4-24h8m-4-4v8m-12 4h.02" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
      </svg>
      <h3 class="mt-2 text-sm font-medium text-gray-900">No objects</h3>
      <p class="mt-1 text-sm text-gray-500">Upload objects to this bucket to get started.</p>
      <div class="mt-6">
        <button
          @click="showUploadModal = true"
          class="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
        >
          <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
          </svg>
          Upload Object
        </button>
      </div>
    </div>

    <div v-else class="bg-white shadow-md rounded-lg overflow-hidden">
      <table class="min-w-full divide-y divide-gray-200">
        <thead class="bg-gray-50">
          <tr>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Name</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Size</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Storage Class</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Last Modified</th>
            <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
          </tr>
        </thead>
        <tbody class="bg-white divide-y divide-gray-200">
          <tr v-for="object in objects" :key="object.key" class="hover:bg-gray-50">
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
              <div class="flex items-center">
                <svg class="w-5 h-5 text-gray-400 mr-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                </svg>
                {{ object.key }}
              </div>
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
              {{ formatFileSize(object.size) }}
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
              {{ object.storage_class || 'STANDARD' }}
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
              {{ object.etag }}
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
              <div class="flex justify-end space-x-2">
                <button
                  @click="viewObject(object.key)"
                  class="text-blue-600 hover:text-blue-800"
                  title="View"
                >
                  <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                  </svg>
                </button>
                <button
                  @click="downloadObject(object.key)"
                  class="text-green-600 hover:text-green-800"
                  title="Download"
                >
                  <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 10v6m0 0l-3-3m3 3l3-3M3 17V7a2 2 0 012-2h6l2 2h6a2 2 0 012 2v10a2 2 0 01-2 2H5a2 2 0 01-2-2z" />
                  </svg>
                </button>
                <button
                  @click="confirmDelete(object.key)"
                  class="text-red-600 hover:text-red-800"
                  title="Delete"
                >
                  <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                  </svg>
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Upload Modal -->
    <div v-if="showUploadModal" class="fixed inset-0 z-50 overflow-y-auto">
      <div class="flex items-center justify-center min-h-screen pt-4 px-4 pb-20 text-center">
        <div class="fixed inset-0 bg-gray-500 bg-opacity-75" @click="showUploadModal = false"></div>
        <div class="inline-block bg-white rounded-lg text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-lg sm:w-full">
          <div class="bg-white px-4 pt-5 pb-4 sm:p-6 sm:pb-4">
            <h3 class="text-lg font-medium text-gray-900 mb-4">Upload Object</h3>

            <div class="mb-4">
              <label class="block text-sm font-medium text-gray-700 mb-2">Object Key (Name)</label>
              <input
                v-model="uploadKey"
                type="text"
                placeholder="Enter object name or leave empty to use filename"
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
            </div>

            <div class="mb-4">
              <label class="block text-sm font-medium text-gray-700 mb-2">File</label>
              <input
                @change="handleFileSelect"
                type="file"
                ref="fileInput"
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
            </div>

            <div v-if="selectedFile" class="mb-4 p-3 bg-gray-50 rounded-md">
              <p class="text-sm text-gray-700"><strong>File:</strong> {{ selectedFile.name }}</p>
              <p class="text-sm text-gray-700"><strong>Size:</strong> {{ formatFileSize(selectedFile.size) }}</p>
            </div>
          </div>
          <div class="bg-gray-50 px-4 py-3 sm:px-6 sm:flex sm:flex-row-reverse">
            <button
              @click="uploadFile"
              :disabled="!selectedFile || uploading"
              class="w-full inline-flex justify-center rounded-md border border-transparent shadow-sm px-4 py-2 bg-blue-600 text-base font-medium text-white hover:bg-blue-700 disabled:bg-gray-400 sm:ml-3 sm:w-auto sm:text-sm"
            >
              {{ uploading ? 'Uploading...' : 'Upload' }}
            </button>
            <button
              @click="cancelUpload"
              :disabled="uploading"
              class="mt-3 w-full inline-flex justify-center rounded-md border border-gray-300 shadow-sm px-4 py-2 bg-white text-base font-medium text-gray-700 hover:bg-gray-50 disabled:bg-gray-100 sm:mt-0 sm:w-auto sm:text-sm"
            >
              Cancel
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Delete Confirmation Modal -->
    <div v-if="objectToDelete" class="fixed inset-0 z-50 overflow-y-auto">
      <div class="flex items-center justify-center min-h-screen pt-4 px-4 pb-20 text-center">
        <div class="fixed inset-0 bg-gray-500 bg-opacity-75" @click="objectToDelete = null"></div>
        <div class="inline-block bg-white rounded-lg text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-lg sm:w-full">
          <div class="bg-white px-4 pt-5 pb-4 sm:p-6 sm:pb-4">
            <h3 class="text-lg font-medium text-gray-900 mb-4">Delete Object</h3>
            <p class="text-sm text-gray-700">Are you sure you want to delete "{{ objectToDelete }}"?</p>
            <p class="text-sm text-red-500 mt-2">This action cannot be undone.</p>
          </div>
          <div class="bg-gray-50 px-4 py-3 sm:px-6 sm:flex sm:flex-row-reverse">
            <button
              @click="deleteObject"
              class="w-full inline-flex justify-center rounded-md border border-transparent shadow-sm px-4 py-2 bg-red-600 text-base font-medium text-white hover:bg-red-700 sm:ml-3 sm:w-auto sm:text-sm"
            >
              Delete
            </button>
            <button
              @click="objectToDelete = null"
              class="mt-3 w-full inline-flex justify-center rounded-md border border-gray-300 shadow-sm px-4 py-2 bg-white text-base font-medium text-gray-700 hover:bg-gray-50 sm:mt-0 sm:w-auto sm:text-sm"
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
import { ref, onMounted } from 'vue'

// Define props
interface Props {
  bucket: string
}

const props = defineProps<Props>()

// Define S3Object interface
interface S3Object {
  key: string
  size: number
  etag: string
  storage_class: string
}

// Reactive data
const objects = ref<S3Object[]>([])
const loading = ref(false)
const error = ref('')
const showUploadModal = ref(false)
const objectToDelete = ref<string | null>(null)
// Upload-related reactive data
const uploadKey = ref('')
const selectedFile = ref<File | null>(null)
const uploading = ref(false)
const fileInput = ref<HTMLInputElement | null>(null)

// Methods
const refreshObjects = async () => {
  if (!props.bucket) return

  loading.value = true
  error.value = ''

  try {
    const response = await fetch(`/api/objects?bucket=${encodeURIComponent(props.bucket)}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      }
    })

    if (!response.ok) {
      const errorText = await response.text()
      throw new Error(errorText || `Failed to fetch objects: ${response.status}`)
    }

    const data = await response.json()
    objects.value = Array.isArray(data) ? data : []
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to load objects'
    console.error('Error fetching objects:', err)
  } finally {
    loading.value = false
  }
}

const deleteObject = async () => {
  if (!objectToDelete.value) return

  try {
    const response = await fetch(`/api/objects/${encodeURIComponent(objectToDelete.value)}?bucket=${encodeURIComponent(props.bucket)}`, {
      method: 'DELETE',
      headers: {
        'Content-Type': 'application/json',
      }
    })

    if (!response.ok) {
      const errorText = await response.text()
      throw new Error(errorText || `Failed to delete object: ${response.status}`)
    }

    // Refresh objects list
    await refreshObjects()
    objectToDelete.value = null
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to delete object'
    console.error('Error deleting object:', err)
  }
}

const formatFileSize = (bytes: number): string => {
  if (bytes === 0) return '0 Bytes'

  const k = 1024
  const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))

  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const viewObject = (key: string) => {
  const fileExtension = key.split('.').pop()?.toLowerCase();
  const mediaExtensions = ['jpg', 'jpeg', 'png', 'gif', 'bmp', 'svg', 'mp4', 'webm', 'mp3', 'wav', 'ogg', 'pdf'];

  // Use the objects endpoint for viewing files
  const viewUrl = `/api/objects/${encodeURIComponent(key)}?bucket=${encodeURIComponent(props.bucket)}`;

  if (fileExtension && mediaExtensions.includes(fileExtension)) {
    // For media files, create a modal with appropriate embedded content
    const modal = document.createElement('div');
    modal.style.position = 'fixed';
    modal.style.top = '0';
    modal.style.left = '0';
    modal.style.width = '100%';
    modal.style.height = '100%';
    modal.style.backgroundColor = 'rgba(0,0,0,0.8)';
    modal.style.zIndex = '9999';
    modal.style.display = 'flex';
    modal.style.justifyContent = 'center';
    modal.style.alignItems = 'center';

    // Create close button
    const closeBtn = document.createElement('button');
    closeBtn.textContent = 'Close';
    closeBtn.style.position = 'absolute';
    closeBtn.style.top = '20px';
    closeBtn.style.right = '20px';
    closeBtn.style.padding = '10px';
    closeBtn.style.background = '#fff';
    closeBtn.style.border = 'none';
    closeBtn.style.borderRadius = '5px';
    closeBtn.style.cursor = 'pointer';
    closeBtn.onclick = () => document.body.removeChild(modal);

    let content: HTMLElement;

    // Create appropriate element based on file type
    if (['jpg', 'jpeg', 'png', 'gif', 'bmp', 'svg'].includes(fileExtension)) {
      const imgElement = document.createElement('img');
      imgElement.src = viewUrl;
      imgElement.style.maxWidth = '90%';
      imgElement.style.maxHeight = '90%';
      imgElement.style.objectFit = 'contain';
      content = imgElement;
    } else if (['mp4', 'webm'].includes(fileExtension)) {
      const videoElement = document.createElement('video');
      videoElement.src = viewUrl;
      videoElement.controls = true;
      videoElement.autoplay = true;
      videoElement.style.maxWidth = '90%';
      videoElement.style.maxHeight = '90%';
      content = videoElement;
    } else if (['mp3', 'wav', 'ogg'].includes(fileExtension)) {
      const audioElement = document.createElement('audio');
      audioElement.src = viewUrl;
      audioElement.controls = true;
      audioElement.autoplay = true;
      audioElement.style.width = '80%';
      content = audioElement;
    } else if (fileExtension === 'pdf') {
      const iframeElement = document.createElement('iframe');
      iframeElement.src = viewUrl;
      iframeElement.style.width = '90%';
      iframeElement.style.height = '90%';
      iframeElement.style.border = 'none';
      content = iframeElement;
    } else {
      // Default case for other media types - use iframe
      const iframeElement = document.createElement('iframe');
      iframeElement.src = viewUrl;
      iframeElement.style.width = '90%';
      iframeElement.style.height = '90%';
      iframeElement.style.border = 'none';
      content = iframeElement;
    }

    // Add click handler to close when clicking outside the content
    modal.onclick = (event) => {
      if (event.target === modal) {
        document.body.removeChild(modal);
      }
    };

    modal.appendChild(content);
    modal.appendChild(closeBtn);
    document.body.appendChild(modal);
  } else {
    // For non-media files (like JSON, text, etc.), open in a new tab
    window.open(viewUrl, '_blank');
  }
}

const downloadObject = (key: string) => {
  // Direct download by creating a link to the object endpoint
  const link = document.createElement('a')
  link.href = `/api/objects/${encodeURIComponent(key)}?bucket=${encodeURIComponent(props.bucket)}`
  link.download = key.split('/').pop() || key
  link.target = '_blank'
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
}

const confirmDelete = (key: string) => {
  objectToDelete.value = key
}

// Upload functions
const handleFileSelect = (event: Event) => {
  const target = event.target as HTMLInputElement
  if (target.files && target.files[0]) {
    selectedFile.value = target.files[0]
    // Auto-fill upload key with filename if empty
    if (!uploadKey.value) {
      uploadKey.value = target.files[0].name
    }
  }
}

const uploadFile = async () => {
  if (!selectedFile.value || !props.bucket) return

  uploading.value = true
  error.value = ''

  try {
    const formData = new FormData()
    formData.append('file', selectedFile.value)

    // Use uploadKey if provided, otherwise use filename
    const key = uploadKey.value || selectedFile.value.name

    const response = await fetch(`/api/objects/${encodeURIComponent(key)}?bucket=${encodeURIComponent(props.bucket)}`, {
      method: 'POST',
      body: formData
    })

    if (!response.ok) {
      const errorText = await response.text()
      throw new Error(errorText || `Failed to upload object: ${response.status}`)
    }

    // Reset upload form and refresh objects
    cancelUpload()
    await refreshObjects()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to upload object'
    console.error('Error uploading object:', err)
  } finally {
    uploading.value = false
  }
}

const cancelUpload = () => {
  showUploadModal.value = false
  uploadKey.value = ''
  selectedFile.value = null
  uploading.value = false
  if (fileInput.value) {
    fileInput.value.value = ''
  }
}

// Load objects on mount
onMounted(() => {
  refreshObjects()
})
</script>
