<template>
  <div class="min-h-screen relative z-10 flex items-center justify-center p-6">
    <!-- Animated floating circles -->
    <div class="absolute inset-0 overflow-hidden pointer-events-none">
      <div class="absolute top-1/4 left-1/4 w-20 h-20 bg-white bg-opacity-10 rounded-full animate-pulse animation-delay-0"></div>
      <div class="absolute top-3/4 right-1/4 w-32 h-32 bg-white bg-opacity-10 rounded-full animate-bounce animation-delay-1000"></div>
      <div class="absolute bottom-1/4 left-1/3 w-16 h-16 bg-white bg-opacity-10 rounded-full animate-pulse animation-delay-2000"></div>
    </div>
    
    <!-- Main content card -->
    <Card class="max-w-2xl w-full glass-modal">
      <template #content>
        <!-- 404 Icon -->
        <div class="mb-8 text-center">
          <div class="relative inline-block">
            <i class="pi pi-search text-8xl text-indigo-600 animate-bounce"></i>
            <div class="absolute -top-2 -right-2">
              <Badge value="404" severity="danger" size="large" class="animate-pulse" />
            </div>
          </div>
        </div>
        
        <!-- Error Text -->
        <div class="mb-10 text-center">
          <h1 class="text-6xl font-black text-gradient mb-4">
            4<span class="animate-pulse">0</span>4
          </h1>
          <h2 class="text-3xl font-bold text-gray-900 mb-4">Oops! Page not found</h2>
          <p class="text-lg text-gray-600 font-medium max-w-md mx-auto leading-relaxed">
            The page you're looking for seems to have wandered off.
            <br class="hidden sm:block">Don't worry, it happens to the best of us!
          </p>
        </div>
        
        <!-- Action Buttons -->
        <div class="flex flex-col sm:flex-row gap-4 justify-center mb-8">
          <Button 
            @click="goHome" 
            class="btn-gradient"
            size="large"
          >
            <i class="pi pi-home mr-3"></i>
            <span>Take me home</span>
            <i class="pi pi-arrow-right ml-3"></i>
          </Button>
          
          <Button 
            @click="goBack" 
            class="btn-glass"
            size="large"
          >
            <i class="pi pi-arrow-left mr-3"></i>
            <span>Go back</span>
          </Button>
        </div>
        
        <!-- Help text -->
        <div class="pt-6 border-t border-gray-200">
          <InlineMessage severity="info" class="w-full">
            If you believe this is an error, please contact support or try refreshing the page.
          </InlineMessage>
        </div>
      </template>
    </Card>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import Card from 'primevue/card'
import Button from 'primevue/button'
import Badge from 'primevue/badge'
import InlineMessage from 'primevue/inlinemessage'

const router = useRouter()

const goHome = () => {
  router.push('/')
}

const goBack = () => {
  // Check if there's history to go back to
  if (window.history.length > 1) {
    router.go(-1)
  } else {
    router.push('/')
  }
}
</script>
