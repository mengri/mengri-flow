import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { getToken } from '@/utils/request'

const routes: RouteRecordRaw[] = [
  // --- Public routes (no auth required) ---
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/login/index.vue'),
    meta: { public: true },
  },
  {
    path: '/activation',
    name: 'Activation',
    component: () => import('@/views/activation/index.vue'),
    meta: { public: true },
  },
  // --- Authenticated routes (with layout) ---
  {
    path: '/',
    name: 'Home',
    component: () => import('@/views/home/index.vue'),
  },
  {
    path: '/account',
    name: 'Account',
    component: () => import('@/views/account/index.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/admin/accounts',
    name: 'AdminAccounts',
    component: () => import('@/views/admin/accounts/index.vue'),
    meta: { requiresAuth: true, requiresAdmin: true },
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

// Navigation guard
router.beforeEach(async (to) => {
  const requiresAuth = to.meta.requiresAuth === true
  const token = getToken()

  // 已登录用户访问登录页 → 跳转首页
  if (to.path === '/login' && token) {
    return { path: '/' }
  }

  // 需要认证但没有 token → 跳转登录页
  if (requiresAuth && !token) {
    return { path: '/login', query: { redirect: to.fullPath } }
  }

  // Admin 路由检查 — 需要在拿到 profile 后才能判断，
  // 这里先做基础的 token 检查，具体角色由页面级别判断
  // (profile 在 App.vue onMounted 中加载)

  return true
})

export default router
