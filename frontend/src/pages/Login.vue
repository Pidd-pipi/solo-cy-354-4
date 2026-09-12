<template>
  <div class="auth-page">
    <el-card class="auth-card">
      <h2 class="auth-title">登录</h2>
      <el-form label-width="70px">
        <el-form-item label="手机号">
          <el-input v-model="form.phone" placeholder="11位手机号" maxlength="11" />
        </el-form-item>
        <el-form-item label="密码">
          <el-input v-model="form.password" type="password" show-password placeholder="密码" />
        </el-form-item>
      </el-form>
      <el-button type="primary" class="submit-btn" :loading="loading" @click="submit">登录</el-button>
      <div class="auth-switch">还没有账号？<el-link type="primary" @click="$router.push('/register')">去注册</el-link></div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { login } from '../api/user'
import { useAuthStore } from '../stores/authStore'

const form = reactive({ phone: '', password: '' })
const loading = ref(false)
const authStore = useAuthStore()
const router = useRouter()

async function submit() {
  if (!form.phone || !form.password) {
    ElMessage.warning('请输入手机号和密码')
    return
  }
  loading.value = true
  try {
    const res = await login(form)
    authStore.setSession(res.data.token, res.data.user)
    ElMessage.success('登录成功')
    router.push('/products')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.auth-page {
  display: flex;
  justify-content: center;
  padding-top: 80px;
}
.auth-card {
  width: 420px;
}
.auth-title {
  text-align: center;
  margin-top: 0;
}
.submit-btn {
  width: 100%;
}
.auth-switch {
  margin-top: 12px;
  text-align: center;
  font-size: 13px;
  color: #909399;
}
</style>
