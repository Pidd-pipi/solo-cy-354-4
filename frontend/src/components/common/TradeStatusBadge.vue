<template>
  <el-tag :type="tradeStatusType(status) as any" size="small">{{ tradeStatusLabel(status) }}</el-tag>
</template>

<script setup lang="ts">
import { tradeStatusLabel, tradeStatusType } from '../../constants/trade'

defineProps<{ status: string }>()
</script>
