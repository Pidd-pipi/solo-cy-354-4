<template>
  <div class="page">
    <h2>发布商品</h2>
    <el-card style="max-width: 640px">
      <ProductForm ref="formRef" />
      <el-button type="primary" :loading="submitting" @click="submit">发布</el-button>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import ProductForm from '../components/common/ProductForm.vue'
import { createProduct } from '../api/product'
import { useAuthStore } from '../stores/authStore'
import { useRouter } from 'vue-router'

const formRef = ref<InstanceType<typeof ProductForm>>()
const submitting = ref(false)
const authStore = useAuthStore()
const router = useRouter()

async function submit() {
  if (!authStore.token) {
    ElMessage.warning('请先登录')
    router.push('/login')
    return
  }
  const form = formRef.value?.form
  if (!form || !form.title || !form.category || !form.campus || form.price <= 0) {
    ElMessage.warning('请填写完整信息')
    return
  }
  submitting.value = true
  try {
    await createProduct({ ...form })
    ElMessage.success('发布成功')
    router.push('/products')
  } finally {
    submitting.value = false
  }
}
</script>
