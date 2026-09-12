import axios, { type AxiosInstance } from 'axios'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '../stores/authStore'

const request: AxiosInstance = axios.create({
  baseURL: '/api/v1',
  timeout: 10000,
})

request.interceptors.request.use((config) => {
  const authStore = useAuthStore()
  if (authStore.token) {
    config.headers.Authorization = `Bearer ${authStore.token}`
  }
  return config
})

request.interceptors.response.use(
  (response) => {
    const body = response.data
    if (body && body.code !== 0) {
      ElMessage.error(body.message || '请求失败')
      return Promise.reject(new Error(body.message))
    }
    return body
  },
  (error) => {
    const status = error.response?.status
    const message = error.response?.data?.message
    if (status === 401) {
      const authStore = useAuthStore()
      authStore.logout()
    }
    ElMessage.error(message || '网络错误')
    return Promise.reject(error)
  },
)

export default request
