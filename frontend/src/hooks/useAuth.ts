import { computed } from 'vue'
import { useAuthStore } from '../stores/authStore'

export function useAuth() {
  const authStore = useAuthStore()
  const isLoggedIn = computed(() => !!authStore.token)
  const currentUser = computed(() => authStore.user)
  const role = computed(() => authStore.user?.role || '')

  function hasRole(roles: string[]): boolean {
    return roles.includes(role.value)
  }

  return { isLoggedIn, currentUser, role, hasRole }
}
