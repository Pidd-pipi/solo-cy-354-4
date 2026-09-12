import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { TradeOrder } from '../types'
import { listMyOrders } from '../api/tradeOrder'

export const useTradeStore = defineStore('trade', () => {
  const orders = ref<TradeOrder[]>([])

  async function fetch() {
    const res = await listMyOrders({ page: 1, page_size: 100 })
    orders.value = res.data.items
  }

  return { orders, fetch }
})
