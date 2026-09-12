import type { Router } from 'vue-router'
import { useAuthStore } from '../stores/authStore'

export function setupGuards(router: Router) {
  router.beforeEach((to) => {
    const authStore = useAuthStore()
    if (to.meta.requiresAuth && !authStore.token) {
      return { path: '/login', query: { redirect: to.fullPath } }
    }
    return true
  })
}
