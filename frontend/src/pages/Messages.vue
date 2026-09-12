<template>
  <div class="page">
    <h2>价格协商私信</h2>
    <el-row :gutter="16">
      <el-col :span="8">
        <el-card shadow="never">
          <div v-for="c in conversations" :key="c.id" class="conv-item" :class="{ active: c.id === activeId }" @click="open(c.id)">
            会话 #{{ c.id }}（商品 {{ c.product_id }}）
          </div>
          <el-empty v-if="conversations.length === 0" description="暂无会话" :image-size="60" />
        </el-card>
      </el-col>
      <el-col :span="16">
        <el-card shadow="never">
          <div class="chat-box">
            <MessageBubble v-for="m in messages" :key="m.id" :message="m" :mine="m.sender_id === authStore.user?.id" />
            <el-empty v-if="messages.length === 0" description="选择左侧会话开始聊天" :image-size="60" />
          </div>
          <div class="chat-input">
            <el-input v-model="content" placeholder="输入消息，可协商价格与约定交易" @keyup.enter="send" />
            <el-button type="primary" @click="send">发送</el-button>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import MessageBubble from '../components/common/MessageBubble.vue'
import { useConversations } from '../hooks/useConversations'
import { useAuthStore } from '../stores/authStore'

const { conversations, messages, activeId, loadConversations, open, send } = useConversations()
const content = ref('')
const authStore = useAuthStore()

async function sendMsg() {
  if (!content.value.trim()) return
  await send(content.value.trim())
  content.value = ''
}

onMounted(() => {
  loadConversations()
})
</script>

<style scoped>
.conv-item {
  padding: 10px 12px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 13px;
  color: #606266;
}
.conv-item:hover,
.conv-item.active {
  background: #ecf5ff;
  color: #409eff;
}
.chat-box {
  height: 420px;
  overflow-y: auto;
  padding: 8px;
}
.chat-input {
  display: flex;
  gap: 8px;
  margin-top: 8px;
}
</style>
