<template>
  <div class="msg-row" :class="{ mine: mine }">
    <div class="bubble" :class="{ mine: mine }">{{ message.content }}</div>
    <div class="time">{{ formatDateTime(message.created_at) }}</div>
  </div>
</template>

<script setup lang="ts">
import type { Message } from '../../types'
import { formatDateTime } from '../../utils/dateFormat'

defineProps<{ message: Message; mine: boolean }>()
</script>

<style scoped>
.msg-row {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  margin: 8px 0;
}
.msg-row.mine {
  align-items: flex-end;
}
.bubble {
  max-width: 70%;
  padding: 8px 12px;
  border-radius: 8px;
  background: #f0f2f5;
  color: #303133;
  font-size: 14px;
  word-break: break-word;
}
.bubble.mine {
  background: #409eff;
  color: #fff;
}
.time {
  font-size: 11px;
  color: #c0c4cc;
  margin-top: 2px;
}
</style>
