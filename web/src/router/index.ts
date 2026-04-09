import { createRouter, createWebHistory, RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/login/index.vue'),
    meta: { requiresAuth: false },
  },
  {
    path: '/activation',
    name: 'Activation',
    component: () => import('@/views/activation/index.vue'),
    meta: { requiresAuth: false },
  },
  {
    path: '/',
    component: () => import('@/layouts/AppLayout.vue'),
    meta: { requiresAuth: true },
    children: [
      {
        path: '',
        name: 'Dashboard',
        component: () => import('@/views/DashboardView.vue'),
      },
      {
        path: '/account',
        name: 'Account',
        component: () => import('@/views/account/index.vue'),
      },
      {
        path: '/admin/accounts',
        name: 'AdminAccounts',
        component: () => import('@/views/admin/accounts/index.vue'),
        meta: { requiresAdmin: true },
      },
      {
        path: '/resources',
        name: 'Resources',
        component: () => import('@/views/resources/Index.vue'),
      },
      {
        path: '/resources/new',
        name: 'CreateResource',
        component: () => import('@/views/resources/Create.vue'),
      },
      {
        path: '/resources/:id',
        name: 'ResourceDetail',
        component: () => import('@/views/resources/Detail.vue'),
      },
      {
        path: '/tools',
        name: 'Tools',
        component: () => import('@/views/tools/Index.vue'),
      },
      {
        path: '/tools/new',
        name: 'CreateTool',
        component: () => import('@/views/tools/Create.vue'),
      },
      {
        path: '/tools/import',
        name: 'ImportTools',
        component: () => import('@/views/tools/Import.vue'),
      },
      {
        path: '/tools/:id',
        name: 'ToolDetail',
        component: () => import('@/views/tools/Detail.vue'),
      },
      {
        path: '/flows',
        name: 'Flows',
        component: () => import('@/views/flows/Index.vue'),
      },
      {
        path: '/flows/new',
        name: 'CreateFlow',
        component: () => import('@/views/flows/Create.vue'),
      },
      {
        path: '/flows/:id',
        name: 'FlowCanvas',
        component: () => import('@/views/flows/Canvas.vue'),
      },
      {
        path: '/triggers',
        name: 'Triggers',
        component: () => import('@/views/triggers/Index.vue'),
      },
      {
        path: '/triggers/new',
        name: 'CreateTrigger',
        component: () => import('@/views/triggers/Create.vue'),
      },
      {
        path: '/triggers/:id',
        name: 'TriggerDetail',
        component: () => import('@/views/triggers/Detail.vue'),
      },
      {
        path: '/runs',
        name: 'Runs',
        component: () => import('@/views/runs/Index.vue'),
      },
      {
        path: '/runs/:id',
        name: 'RunDetail',
        component: () => import('@/views/runs/Detail.vue'),
      },
    ],
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/',
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

// 路由守卫
router.beforeEach((to, _from, next) => {
  const authStore = useAuthStore()
  
  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    next('/login')
  } else if (to.meta.requiresAdmin && !authStore.isAdmin) {
    next('/')
  } else {
    next()
  }
})

export default router
