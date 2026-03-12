<script setup lang="ts">
import { useRoute } from 'vue-router'
import { computed } from 'vue'

const route = useRoute()

const title = computed(() => {
  switch (route.name) {
    case 'chat': return '消息'
    case 'contacts': return '联系人'
    case 'groups': return '群组'
    default: return ''
  }
})
</script>

<template>
  <aside class="list-panel">
    <div class="list-header">
      <h2 class="list-title">{{ title }}</h2>
      <div class="list-actions">
        <!-- 操作按钮区（加好友/建群等），Phase 2 实现 -->
      </div>
    </div>

    <div class="search-wrap">
      <div class="search-bar">
        <span class="search-icon">🔍</span>
        <input placeholder="搜索" />
      </div>
    </div>

    <!-- 列表内容由各子视图通过 slot 注入，Phase 2 实现 -->
    <div class="list-content">
      <slot>
        <div class="empty-hint">暂无内容</div>
      </slot>
    </div>
  </aside>
</template>

<style scoped>
.list-panel {
  width: var(--list-width);
  flex-shrink: 0;
  height: 100vh;
  background: var(--bg-list);
  border-right: 1px solid var(--border-light);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.list-header {
  height: 52px;
  padding: 0 16px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-shrink: 0;
}

.list-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
}

.search-wrap {
  padding: 0 10px 8px;
  flex-shrink: 0;
}

.search-bar {
  height: 32px;
  background: #E8E8E8;
  border-radius: var(--radius-search);
  display: flex;
  align-items: center;
  padding: 0 10px;
  gap: 6px;
  transition: background var(--transition-fast), box-shadow var(--transition-fast);
}

.search-bar:focus-within {
  background: white;
  box-shadow: 0 0 0 1px var(--qq-blue-primary);
}

.search-icon {
  font-size: 13px;
  opacity: 0.5;
}

.search-bar input {
  flex: 1;
  font-size: 13px;
  color: var(--text-primary);
  background: transparent;
  user-select: text;
}

.search-bar input::placeholder {
  color: var(--text-tertiary);
}

.list-content {
  flex: 1;
  overflow-y: auto;
}

.empty-hint {
  padding: 40px 0;
  text-align: center;
  font-size: 13px;
  color: var(--text-tertiary);
}
</style>
