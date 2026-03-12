<script setup lang="ts">
import { ref, computed } from 'vue'
import Modal from '@/components/common/Modal.vue'
import { useContactsStore } from '@/stores/contacts'
import { groupApi } from '@/api/group'
import type { ApiResponse } from '@/api/client'
import type { Group } from '@/types/group'

const emit = defineEmits<{ close: []; created: [group: Group] }>()

const contacts = useContactsStore()
const groupName = ref('')
const selectedIds = ref<Set<number>>(new Set())
const creating = ref(false)
const error = ref('')

const friends = computed(() => contacts.friendships)

function toggleSelect(id: number) {
  if (selectedIds.value.has(id)) {
    selectedIds.value.delete(id)
  } else {
    selectedIds.value.add(id)
  }
  selectedIds.value = new Set(selectedIds.value)
}

async function create() {
  if (!groupName.value.trim()) {
    error.value = '请输入群名'
    return
  }
  creating.value = true
  error.value = ''
  try {
    const res = await groupApi.create(groupName.value.trim(), [...selectedIds.value])
    const body = res.data as unknown as ApiResponse<Group>
    if (body.data) {
      emit('created', body.data)
      emit('close')
    }
  } catch {
    error.value = '创建失败，请重试'
  } finally {
    creating.value = false
  }
}
</script>

<template>
  <Modal title="创建群组" @close="emit('close')">
    <div class="create-group-form">
      <div class="form-field">
        <label>群名称</label>
        <input v-model="groupName" placeholder="输入群名称" maxlength="30" />
      </div>

      <div class="form-field">
        <label>选择成员（可选）</label>
        <div class="friend-select-list">
          <div
            v-for="f in friends"
            :key="f.friend_id"
            class="friend-select-item"
            :class="{ selected: selectedIds.has(f.friend_id) }"
            @click="toggleSelect(f.friend_id)"
          >
            <div class="check-box">
              <span v-if="selectedIds.has(f.friend_id)">✓</span>
            </div>
            <span class="friend-name">{{ f.remark || f.friend?.nickname }}</span>
          </div>
          <div v-if="friends.length === 0" class="no-friends">暂无好友可添加</div>
        </div>
      </div>

      <div v-if="error" class="error-msg">{{ error }}</div>

      <div class="form-actions">
        <button class="cancel-btn" @click="emit('close')">取消</button>
        <button class="create-btn" :disabled="creating" @click="create">
          {{ creating ? '创建中…' : '创建' }}
        </button>
      </div>
    </div>
  </Modal>
</template>

<style scoped>
.create-group-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
  min-width: 320px;
}

.form-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

label {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-secondary);
}

input {
  height: 36px;
  padding: 0 10px;
  border: 1px solid var(--border-input);
  border-radius: var(--radius-input);
  font-size: 14px;
  color: var(--text-primary);
  outline: none;
  user-select: text;
}

input:focus {
  border-color: var(--qq-blue-primary);
}

.friend-select-list {
  max-height: 200px;
  overflow-y: auto;
  border: 1px solid var(--border-light);
  border-radius: 6px;
}

.friend-select-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 12px;
  cursor: pointer;
  transition: background 0.12s;
}

.friend-select-item:hover { background: var(--bg-hover); }
.friend-select-item.selected { background: var(--qq-blue-light); }

.check-box {
  width: 18px;
  height: 18px;
  border: 1.5px solid var(--border-normal);
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  color: var(--qq-blue-primary);
  flex-shrink: 0;
}

.friend-select-item.selected .check-box {
  background: var(--qq-blue-primary);
  border-color: var(--qq-blue-primary);
  color: white;
}

.friend-name {
  font-size: 14px;
  color: var(--text-primary);
}

.no-friends {
  padding: 20px;
  text-align: center;
  font-size: 13px;
  color: var(--text-tertiary);
}

.error-msg {
  font-size: 13px;
  color: var(--color-error);
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

.cancel-btn {
  height: 34px;
  padding: 0 16px;
  border: 1px solid var(--border-normal);
  border-radius: var(--radius-btn);
  font-size: 13px;
  color: var(--text-secondary);
}

.cancel-btn:hover { background: var(--bg-hover); }

.create-btn {
  height: 34px;
  padding: 0 20px;
  background: var(--qq-blue-primary);
  color: white;
  border-radius: var(--radius-btn);
  font-size: 13px;
}

.create-btn:disabled { opacity: 0.6; cursor: not-allowed; }
.create-btn:not(:disabled):hover { opacity: 0.88; }
</style>
