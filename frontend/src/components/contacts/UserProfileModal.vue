<script setup lang="ts">
import { ref } from 'vue'
import Modal from '@/components/common/Modal.vue'
import Avatar from '@/components/common/Avatar.vue'
import { friendApi } from '@/api/friend'
import { useContactsStore } from '@/stores/contacts'
import type { Friendship } from '@/types/user'

const props = defineProps<{
  friendship: Friendship
}>()

const emit = defineEmits<{ close: []; deleted: [] }>()

const contacts = useContactsStore()
const editingRemark = ref(false)
const remark = ref(props.friendship.remark)
const saving = ref(false)
const confirmDelete = ref(false)

async function saveRemark() {
  saving.value = true
  try {
    await friendApi.updateRemark(props.friendship.friend_id, remark.value)
    props.friendship.remark = remark.value
    editingRemark.value = false
  } finally {
    saving.value = false
  }
}

async function deleteFriend() {
  await friendApi.deleteFriend(props.friendship.friend_id)
  contacts.removeFriend(props.friendship.friend_id)
  emit('deleted')
  emit('close')
}

const displayName = props.friendship.remark || props.friendship.friend?.nickname || ''
</script>

<template>
  <Modal :title="displayName" :width="380" @close="emit('close')">
    <div class="profile-body">
      <div class="profile-header">
        <Avatar
          :name="friendship.friend?.nickname"
          :size="64"
          :status="friendship.friend?.status"
          show-status
        />
        <div class="profile-info">
          <div class="profile-nickname">{{ friendship.friend?.nickname }}</div>
          <div class="profile-username">用户名：{{ friendship.friend?.username }}</div>
          <div v-if="friendship.friend?.bio" class="profile-bio">{{ friendship.friend.bio }}</div>
        </div>
      </div>

      <!-- 备注 -->
      <div class="section">
        <div class="section-label">备注名</div>
        <div v-if="!editingRemark" class="remark-display">
          <span>{{ remark || '未设置' }}</span>
          <button class="text-btn" @click="editingRemark = true">修改</button>
        </div>
        <div v-else class="remark-edit">
          <input v-model="remark" class="remark-input" placeholder="输入备注名" maxlength="20" />
          <div class="remark-actions">
            <button class="text-btn" @click="editingRemark = false">取消</button>
            <button class="primary-btn" :disabled="saving" @click="saveRemark">
              {{ saving ? '保存中...' : '保存' }}
            </button>
          </div>
        </div>
      </div>

      <!-- 危险操作 -->
      <div class="danger-section">
        <div v-if="!confirmDelete">
          <button class="danger-btn" @click="confirmDelete = true">删除好友</button>
        </div>
        <div v-else class="confirm-delete">
          <span>确认删除好友「{{ friendship.friend?.nickname }}」？</span>
          <div class="confirm-actions">
            <button class="text-btn" @click="confirmDelete = false">取消</button>
            <button class="danger-btn-sm" @click="deleteFriend">确认删除</button>
          </div>
        </div>
      </div>
    </div>
  </Modal>
</template>

<style scoped>
.profile-body { display: flex; flex-direction: column; gap: 20px; }

.profile-header {
  display: flex;
  align-items: flex-start;
  gap: 16px;
}

.profile-info { flex: 1; }

.profile-nickname {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 4px;
}

.profile-username {
  font-size: 12px;
  color: var(--text-secondary);
  margin-bottom: 4px;
}

.profile-bio {
  font-size: 13px;
  color: var(--text-tertiary);
}

.section { display: flex; flex-direction: column; gap: 6px; }

.section-label {
  font-size: 12px;
  color: var(--text-secondary);
  font-weight: 500;
}

.remark-display {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  color: var(--text-primary);
}

.remark-edit { display: flex; flex-direction: column; gap: 8px; }

.remark-input {
  height: 36px;
  border: 1px solid var(--border-input);
  border-radius: var(--radius-input);
  padding: 0 12px;
  font-size: 13px;
  background: var(--bg-input);
  color: var(--text-primary);
  user-select: text;
}

.remark-input:focus {
  border-color: var(--qq-blue-primary);
}

.remark-actions { display: flex; justify-content: flex-end; gap: 8px; }

.text-btn {
  font-size: 13px;
  color: var(--qq-blue-primary);
  padding: 4px 8px;
}

.text-btn:hover { text-decoration: underline; }

.primary-btn {
  height: 30px;
  padding: 0 14px;
  background: var(--qq-blue-primary);
  color: white;
  border-radius: var(--radius-btn);
  font-size: 13px;
}

.primary-btn:disabled { opacity: 0.6; cursor: not-allowed; }

.danger-section { border-top: 1px solid var(--border-light); padding-top: 16px; }

.danger-btn {
  font-size: 13px;
  color: var(--color-error);
  padding: 6px 12px;
  border: 1px solid var(--color-error);
  border-radius: var(--radius-btn);
  transition: background var(--transition-fast);
}

.danger-btn:hover { background: #fff1f0; }

.confirm-delete {
  display: flex;
  flex-direction: column;
  gap: 10px;
  font-size: 13px;
  color: var(--text-primary);
}

.confirm-actions { display: flex; gap: 8px; }

.danger-btn-sm {
  height: 30px;
  padding: 0 14px;
  background: var(--color-error);
  color: white;
  border-radius: var(--radius-btn);
  font-size: 13px;
}
</style>
