import { ref } from 'vue'
import { listProducts } from '../api/product'
import type { Product } from '../types'

export function useProducts() {
  const products = ref<Product[]>([])
  const total = ref(0)
  const loading = ref(false)

  async function load(params: { page?: number; page_size?: number; category?: string; campus?: string; keyword?: string; status?: string } = {}) {
    loading.value = true
    try {
      const res = await listProducts({ page: 1, page_size: 100, ...params })
      products.value = res.data.items
      total.value = res.data.total
    } finally {
      loading.value = false
    }
  }

  return { products, total, loading, load }
}
