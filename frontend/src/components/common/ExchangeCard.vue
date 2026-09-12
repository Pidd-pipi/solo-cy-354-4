<template>
  <el-card class="exchange-card" shadow="hover">
    <div class="exchange-head">
      <h3>{{ exchange.offer_book }} ⇄ {{ exchange.want_book }}</h3>
      <el-tag size="small" :type="exchange.status === 'open' ? 'success' : 'info'">{{ exchange.status === 'open' ? '开放中' : exchange.status === 'matched' ? '已匹配' : '已关闭' }}</el-tag>
    </div>
    <p class="exchange-desc">{{ exchange.description || '暂无说明' }}</p>
    <div class="exchange-actions">
      <el-button v-if="mine && exchange.status === 'open'" size="small" type="danger" @click="$emit('close', exchange)">关闭交换</el-button>
    </div>
  </el-card>
</template>

<script setup lang="ts">
import type { BookExchange } from '../../types'

defineProps<{ exchange: BookExchange; mine: boolean }>()
defineEmits<{ (e: 'close', x: BookExchange): void }>()
</script>

<style scoped>
.exchange-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.exchange-head h3 {
  margin: 0;
  font-size: 15px;
}
.exchange-desc {
  color: #909399;
  font-size: 13px;
  margin: 8px 0;
}
</style>
