<template>
  <div class="page">
    <h2>书籍交换</h2>
    <el-card class="publish-card">
      <template #header>发布交换书籍</template>
      <el-form inline>
        <el-form-item label="可交换">
          <el-input v-model="form.offer_book" placeholder="书名" style="width: 180px" />
        </el-form-item>
        <el-form-item label="想换">
          <el-input v-model="form.want_book" placeholder="书名" style="width: 180px" />
        </el-form-item>
        <el-form-item label="说明">
          <el-input v-model="form.description" placeholder="可选" style="width: 220px" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="submitting" @click="publish">发布</el-button>
        </el-form-item>
      </el-form>
    </el-card>
    <el-row :gutter="16">
      <el-col v-for="e in exchanges" :key="e.id" :span="8" class="col">
        <ExchangeCard :exchange="e" :mine="e.user_id === authStore.user?.id" @close="closeExchange" />
      </el-col>
    </el-row>
    <el-empty v-if="!loading && exchanges.length === 0" description="暂无交换请求" />
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import ExchangeCard from '../components/common/ExchangeCard.vue'
import { createBookExchange, listBookExchanges, closeBookExchange } from '../api/bookExchange'
import type { BookExchange } from '../types'
import { useAuthStore } from '../stores/authStore'
import { useRouter } from 'vue-router'

const exchanges = ref<BookExchange[]>([])
const loading = ref(false)
const submitting = ref(false)
const authStore = useAuthStore()
const router = useRouter()
const form = reactive({ offer_book: '', want_book: '', description: '' })

async function publish() {
  if (!authStore.token) {
    ElMessage.warning('请先登录')
    router.push('/login')
    return
  }
  if (!form.offer_book || !form.want_book) {
    ElMessage.warning('请填写书名')
    return
  }
  submitting.value = true
  try {
    await createBookExchange({ ...form })
    ElMessage.success('发布成功（若存在匹配将自动撮合）')
    form.offer_book = ''
    form.want_book = ''
    form.description = ''
    await load()
  } finally {
    submitting.value = false
  }
}

async function closeExchange(e: BookExchange) {
  await closeBookExchange(e.id)
  ElMessage.success('已关闭')
  await load()
}

async function load() {
  loading.value = true
  try {
    const res = await listBookExchanges({ page: 1, page_size: 100 })
    exchanges.value = res.data.items
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<style scoped>
.publish-card {
  margin-bottom: 16px;
}
.col {
  margin-bottom: 16px;
}
</style>
