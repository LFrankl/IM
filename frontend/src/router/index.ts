import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/LoginView.vue'),
      meta: { public: true },
    },
    {
      path: '/',
      name: 'main',
      component: () => import('@/views/MainView.vue'),
      redirect: '/chat',
      children: [
        { path: 'chat', name: 'chat', component: () => import('@/views/ChatView.vue') },
        { path: 'contacts', name: 'contacts', component: () => import('@/views/ContactsView.vue') },
        { path: 'groups', name: 'groups', component: () => import('@/views/GroupsView.vue') },
        { path: 'space/:userId?', name: 'space', component: () => import('@/views/SpaceView.vue') },
      ],
    },
    { path: '/:pathMatch(.*)*', redirect: '/' },
  ],
})

// 路由守卫
router.beforeEach(async (to) => {
  const auth = useAuthStore()
  if (to.meta.public) return true

  if (!auth.token) {
    return { name: 'login' }
  }

  // 已有 token 但没有 user 信息（页面刷新场景）
  if (!auth.user) {
    try {
      await auth.fetchMe()
    } catch {
      return { name: 'login' }
    }
  }

  return true
})

export default router
