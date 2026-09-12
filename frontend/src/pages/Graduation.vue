<template>
  <div class="page">
    <h2>🎓 毕业季专场</h2>
    <el-alert title="毕业季专场：学长学姐闲置好物集中放送" type="warning" :closable="false" class="banner" />
    <el-row :gutter="16">
      <el-col v-for="p in products" :key="p.id" :span="6" class="col">
        <ProductCard :product="p" @detail="showDetail" @buy="buy" />
      </el-col>
    </el-row>
    <el-empty v-if="!loading && products.length === 0" description="专场暂无商品" />
    <el-dialog v-model="detailVisible" :title="current?.title" width="520px">
      <el-descriptions :column="2" border v-if="current">
        <el-descriptions-item label="分类">{{ categoryLabel(current.category) }}</el-descriptions-item>
        <el-descriptions-item label="成色">{{ current.condition }}</el-descriptions-item>
        <el-descriptions-item label="校区">{{ current.campus }}</el-descriptions-item>
        <el-descriptions-item label="价格">¥{{ current.price.toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item label="描述" :span="2">{{ current.description }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import ProductCard from '../components/common/ProductCard.vue'
import { listGraduation } from '../api/product'
import { createTradeOrder } from '../api/tradeOrder'
import { categoryLabel } from '../constants/product'
import type { Product } from '../types'
import { useAuthStore } from '../stores/authStore'
import { useRouter } from 'vue-router'

const products = ref<Product[]>([])
const loading = ref(false)
const detailVisible = ref(false)
const current = ref<Product | null>(null)
const authStore = useAuthStore()
const router = useRouter()

function showDetail(p: Product) {
  current.value = p
  detailVisible.value = true
}

async function buy(p: Product) {
  if (!authStore.token) {
    ElMessage.warning('请先登录')
    router.push('/login')
    return
  }
  await createTradeOrder(p.id)
  ElMessage.success('已下单')
}

onMounted(async () => {
  loading.value = true
  try {
    const res = await listGraduation()
    products.value = res.data.items
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.banner {
  margin-bottom: 16px;
}
.col {
  margin-bottom: 16px;
}
</style>
