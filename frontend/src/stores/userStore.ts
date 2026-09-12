import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { User } from '../types'
import { useAuthStore } from './authStore'
import { getProfile } from '../api/user'

export const useUserStore = defineStore('user', () => {
  const profile = ref<User | null>(null)

  async function refresh() {
    const authStore = useAuthStore()
    if (!authStore.token) return
    const res = await getProfile()
    profile.value = res.data
    authStore.user = res.data
  }

  return { profile, refresh }
})
