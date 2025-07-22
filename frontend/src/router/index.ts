import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '@/views/HomeView.vue'
import ObjectsView from '@/views/ObjectsView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'Home',
      component: HomeView,
    },
    {
      path: '/objects/:bucket',
      name: 'Objects',
      component: ObjectsView,
      props: true,
    },
  ],
})

export default router
