<template>
  <el-card class="product-card" shadow="hover">
    <div class="product-head">
      <h3>{{ product.title }}</h3>
      <el-tag :type="productStatusType(product.status) as any" size="small">{{ productStatusLabel(product.status) }}</el-tag>
    </div>
    <p class="product-desc">{{ product.description || '暂无描述' }}</p>
    <div class="product-price">¥{{ product.price.toFixed(2) }}</div>
    <div class="product-meta">
      <span>{{ categoryLabel(product.category) }}</span>
      <span>{{ product.campus }}</span>
      <span>{{ product.condition }}</span>
    </div>
    <div class="product-actions">
      <el-button size="small" @click="$emit('detail', product)">详情</el-button>
      <el-button v-if="!hideBuy" size="small" type="primary" :disabled="product.status !== 'on_sale'" @click="$emit('buy', product)">购买</el-button>
      <el-button v-if="showChat" size="small" @click="$emit('chat', product)">私信</el-button>
    </div>
  </el-card>
</template>

<script setup lang="ts">
import type { Product } from '../../types'
import { categoryLabel, productStatusLabel, productStatusType } from '../../constants/product'

withDefaults(defineProps<{ product: Product; hideBuy?: boolean; showChat?: boolean }>(), { hideBuy: false, showChat: false })
defineEmits<{ (e: 'detail', p: Product): void; (e: 'buy', p: Product): void; (e: 'chat', p: Product): void }>()
</script>

<style scoped>
.product-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
}
.product-head h3 {
  margin: 0;
  font-size: 16px;
  color: #303133;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}
.product-desc {
  color: #909399;
  font-size: 13px;
  height: 36px;
  overflow: hidden;
  margin: 8px 0;
}
.product-price {
  color: #f56c6c;
  font-size: 20px;
  font-weight: 700;
}
.product-meta {
  display: flex;
  gap: 12px;
  font-size: 12px;
  color: #606266;
  margin: 8px 0;
}
</style>
