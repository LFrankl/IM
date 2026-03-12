<script setup lang="ts">
import { computed } from 'vue'
import type { UserStatus } from '@/types/user'

const props = withDefaults(defineProps<{
  src?: string
  name?: string
  size?: number
  status?: UserStatus | null
  showStatus?: boolean
}>(), {
  size: 40,
  showStatus: false,
})

const initial = computed(() => {
  if (!props.name) return '?'
  return props.name.charAt(0).toUpperCase()
})

// 根据名字生成固定颜色
const bgColor = computed(() => {
  if (!props.name) return '#1677FF'
  const colors = ['#1677FF', '#52C41A', '#FA8C16', '#F5222D', '#722ED1', '#13C2C2', '#EB2F96']
  let hash = 0
  for (const c of props.name) hash = (hash * 31 + c.charCodeAt(0)) & 0xffff
  return colors[hash % colors.length]
})

const statusClass = computed(() => {
  switch (props.status) {
    case 'online': return 'status-online'
    case 'busy': return 'status-busy'
    default: return 'status-offline'
  }
})
</script>

<template>
  <div class="avatar-wrap" :style="{ width: `${size}px`, height: `${size}px` }">
    <img v-if="src" :src="src" class="avatar-img" :alt="name" />
    <div v-else class="avatar-initial" :style="{ background: bgColor, fontSize: `${Math.round(size * 0.4)}px` }">
      {{ initial }}
    </div>
    <span v-if="showStatus" class="status-dot" :class="statusClass" />
  </div>
</template>

<style scoped>
.avatar-wrap {
  position: relative;
  flex-shrink: 0;
}

.avatar-img,
.avatar-initial {
  width: 100%;
  height: 100%;
  border-radius: var(--radius-avatar);
  object-fit: cover;
}

.avatar-initial {
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-weight: 600;
}

.status-dot {
  position: absolute;
  bottom: -2px;
  right: -2px;
  width: 10px;
  height: 10px;
  border-radius: 50%;
  border: 2px solid var(--bg-surface);
}

.status-online  { background: var(--color-success); }
.status-busy    { background: var(--color-warning); }
.status-offline { background: var(--color-offline); }
</style>
