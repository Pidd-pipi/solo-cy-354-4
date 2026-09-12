<template>
  <div class="page">
    <h2>个人中心</h2>
    <template v-if="authStore.user">
      <el-card class="profile-card">
        <div class="profile-head">
          <el-avatar :size="64" :src="authStore.user.avatar || ''">{{ authStore.user.nickname.slice(0, 1) }}</el-avatar>
          <div class="profile-info">
            <h3>{{ authStore.user.nickname }}</h3>
            <p>{{ authStore.user.phone }} · {{ roleLabel(authStore.user.role) }} · {{ authStore.user.campus }}</p>
          </div>
          <div class="credit-box">
            <div class="credit-label">信誉分</div>
            <div class="credit-value">{{ authStore.user.credit_score }}</div>
            <div class="credit-level">{{ creditLevel(authStore.user.credit_score) }}</div>
          </div>
        </div>
      </el-card>
      <el-card class="section">
        <template #header>⭐ 收到的评价</template>
        <el-table :data="reviews">
          <el-table-column prop="id" label="ID" width="80" />
          <el-table-column label="评价">
            <template #default="{ row }">
              <el-tag size="small">{{ ratingLabel(row.rating) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="content" label="内容" />
          <el-table-column label="时间">
            <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
          </el-table-column>
        </el-table>
      </el-card>
    </template>
    <el-empty v-else description="请先登录">
      <el-button type="primary" @click="$router.push('/login')">去登录</el-button>
    </el-empty>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useAuthStore } from '../stores/authStore'
import { roleLabel } from '../constants/user'
import { ratingLabel } from '../constants/trade'
import { listMyReviews } from '../api/review'
import { formatDateTime } from '../utils/dateFormat'
import type { Review } from '../types'

const authStore = useAuthStore()
const reviews = ref<Review[]>([])

function creditLevel(score: number): string {
  if (score >= 200) return '极佳'
  if (score >= 150) return '优秀'
  if (score >= 100) return '良好'
  if (score >= 60) return '一般'
  return '待提升'
}

onMounted(async () => {
  if (!authStore.token) return
  const res = await listMyReviews()
  reviews.value = res.data
})
</script>

<style scoped>
.profile-card {
  margin-bottom: 16px;
}
.profile-head {
  display: flex;
  align-items: center;
  gap: 16px;
}
.profile-info h3 {
  margin: 0;
}
.profile-info p {
  margin: 4px 0 0;
  color: #909399;
  font-size: 13px;
}
.credit-box {
  margin-left: auto;
  text-align: center;
}
.credit-label {
  color: #909399;
  font-size: 12px;
}
.credit-value {
  font-size: 28px;
  font-weight: 700;
  color: #e6a23c;
}
.credit-level {
  font-size: 12px;
  color: #909399;
}
.section {
  margin-bottom: 16px;
}
</style>
