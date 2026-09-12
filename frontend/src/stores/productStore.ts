import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Product } from '../types'
import { listProducts } from '../api/product'

export const useProductStore = defineStore('product', () => {
  const products = ref<Product[]>([])
  const total = ref(0)

  async function fetch(params: { page?: number; page_size?: number; category?: string; campus?: string; keyword?: string } = {}) {
    const res = await listProducts({ page: 1, page_size: 100, ...params })
    products.value = res.data.items
    total.value = res.data.total
  }

  return { products, total, fetch }
})
