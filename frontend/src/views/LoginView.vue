<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const auth = useAuthStore()

const isRegister = ref(false)
const loading = ref(false)
const error = ref('')

const form = reactive({
  username: '',
  password: '',
  nickname: '',
})

function switchMode(toRegister: boolean) {
  isRegister.value = toRegister
  error.value = ''
  form.nickname = ''
  form.password = ''
}

async function submit() {
  error.value = ''
  loading.value = true
  try {
    if (isRegister.value) {
      await auth.register(form.username, form.password, form.nickname)
    } else {
      await auth.login(form.username, form.password)
    }
    router.push('/')
  } catch (e: unknown) {
    const msg = (e as { response?: { data?: { message?: string } } })?.response?.data?.message
    error.value = msg || '操作失败，请重试'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login-page">
    <div class="login-card">
      <!-- Logo 区域 -->
      <div class="logo-area">
        <div class="logo-circle">Q</div>
        <div class="logo-title">WebQQ</div>
      </div>

      <!-- 表单 -->
      <form class="login-form" @submit.prevent="submit">
        <div class="form-item">
          <input
            v-model="form.username"
            type="text"
            placeholder="QQ号/用户名"
            required
            maxlength="20"
          />
        </div>

        <div v-if="isRegister" class="form-item">
          <input
            v-model="form.nickname"
            type="text"
            placeholder="昵称"
            required
            maxlength="20"
          />
        </div>

        <div class="form-item">
          <input
            v-model="form.password"
            type="password"
            placeholder="密码"
            required
            minlength="6"
          />
        </div>

        <div v-if="error" class="error-msg">{{ error }}</div>

        <button type="submit" class="submit-btn" :disabled="loading">
          {{ loading ? '请稍候...' : isRegister ? '注册' : '登录' }}
        </button>
      </form>

      <!-- 切换登录/注册 -->
      <div class="switch-mode">
        <span v-if="!isRegister">
          没有账号？
          <a @click="switchMode(true)">立即注册</a>
        </span>
        <span v-else>
          已有账号？
          <a @click="switchMode(false)">立即登录</a>
        </span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.login-page {
  height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #1677FF 0%, #0958D9 100%);
}

.login-card {
  width: 360px;
  background: white;
  border-radius: 16px;
  padding: 40px 36px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.2);
}

.logo-area {
  text-align: center;
  margin-bottom: 32px;
}

.logo-circle {
  width: 64px;
  height: 64px;
  border-radius: 50%;
  background: #1677FF;
  color: white;
  font-size: 32px;
  font-weight: 700;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 12px;
}

.logo-title {
  font-size: 22px;
  font-weight: 600;
  color: var(--text-primary);
}

.login-form {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.form-item input {
  width: 100%;
  height: 44px;
  border: 1px solid var(--border-input);
  border-radius: var(--radius-input);
  padding: 0 14px;
  font-size: 14px;
  background: var(--bg-input);
  color: var(--text-primary);
  transition: border-color var(--transition-fast);
  user-select: text;
}

.form-item input:focus {
  border-color: var(--qq-blue-primary);
  box-shadow: 0 0 0 2px var(--qq-blue-light);
}

.error-msg {
  font-size: 13px;
  color: var(--color-error);
  text-align: center;
}

.submit-btn {
  height: 44px;
  background: var(--qq-blue-primary);
  color: white;
  font-size: 15px;
  font-weight: 500;
  border-radius: var(--radius-btn);
  transition: background var(--transition-fast);
  margin-top: 4px;
}

.submit-btn:hover:not(:disabled) {
  background: var(--qq-blue-hover);
}

.submit-btn:active:not(:disabled) {
  background: var(--qq-blue-pressed);
}

.submit-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.switch-mode {
  margin-top: 20px;
  text-align: center;
  font-size: 13px;
  color: var(--text-secondary);
}

.switch-mode a {
  color: var(--qq-blue-primary);
  cursor: pointer;
}

.switch-mode a:hover {
  text-decoration: underline;
}
</style>
