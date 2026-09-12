<template>
  <div class="auth-page">
    <el-card class="auth-card">
      <h2 class="auth-title">注册</h2>
      <el-form label-width="70px">
        <el-form-item label="手机号">
          <el-input v-model="form.phone" placeholder="11位手机号" maxlength="11" />
        </el-form-item>
        <el-form-item label="昵称">
          <el-input v-model="form.nickname" placeholder="昵称" />
        </el-form-item>
        <el-form-item label="校区">
          <el-input v-model="form.campus" placeholder="如：东校区" />
        </el-form-item>
        <el-form-item label="密码">
          <el-input v-model="form.password" type="password" show-password placeholder="至少6位" />
        </el-form-item>
      </el-form>
      <el-button type="primary" class="submit-btn" :loading="loading" @click="submit">注册</el-button>
      <div class="auth-switch">已有账号？<el-link type="primary" @click="$router.push('/login')">去登录</el-link></div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { register } from '../api/user'

const form = reactive({ phone: '', nickname: '', campus: '', password: '' })
const loading = ref(false)
const router = useRouter()

async function submit() {
  if (!form.phone || !form.nickname || !form.campus || !form.password) {
    ElMessage.warning('请填写完整信息')
    return
  }
  loading.value = true
  try {
    await register(form)
    ElMessage.success('注册成功，请登录')
    router.push('/login')
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
