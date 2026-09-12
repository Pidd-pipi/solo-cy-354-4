import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { User } from '../types'
import { getProfile } from '../api/user'

const TOKEN_KEY = 'lpcampusmarket_token'
const USER_KEY = 'lpcampusmarket_user'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string>(localStorage.getItem(TOKEN_KEY) || '')
  const user = ref<User | null>(JSON.parse(localStorage.getItem(USER_KEY) || 'null'))

  function setSession(newToken: string, newUser: User) {
    token.value = newToken
    user.value = newUser
    localStorage.setItem(TOKEN_KEY, newToken)
    localStorage.setItem(USER_KEY, JSON.stringify(newUser))
  }

  function logout() {
    token.value = ''
    user.value = null
    localStorage.removeItem(TOKEN_KEY)
    localStorage.removeItem(USER_KEY)
  }

  function restore() {
    if (token.value && !user.value) {
      getProfile().then((res) => {
        user.value = res.data
        localStorage.setItem(USER_KEY, JSON.stringify(res.data))
      })
    }
  }

  function isAdmin(): boolean {
    return user.value?.role === 'admin'
  }

  return { token, user, setSession, logout, restore, isAdmin }
})
