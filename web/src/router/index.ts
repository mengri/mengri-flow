import { createRouter, createWebHistory, RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useWorkspaceStore } from '@/stores/workspace'

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
    path: '/select-workspace',
    name: 'SelectWorkspace',
    component: () => import('@/views/workspaces/Select.vue'),
    meta: { requiresAuth: true, skipWorkspaceCheck: true },
  },
  // --- 非空间内路由（直接在 AppLayout 下） ---
  {
    path: '/',
    component: () => import('@/layouts/AppLayout.vue'),
    meta: { requiresAuth: true },
    children: [
      {
        path: '/account',
        name: 'Account',
        component: () => import('@/views/account/index.vue'),
      },
      {
        path: '/workspaces',
        name: 'Workspaces',
        component: () => import('@/views/workspaces/Index.vue'),
      },
      {
        path: '/admin/accounts',
        name: 'AdminAccounts',
        component: () => import('@/views/admin/accounts/index.vue'),
        meta: { requiresAdmin: true },
      },
      {
        path: '',
        redirect: () => {
          const workspaceStore = useWorkspaceStore()
          const wsId = workspaceStore.currentWorkspaceId
          return wsId ? `/workspace/${wsId}` : '/select-workspace'
        },
      },
    ],
  },
  // --- 空间内路由（带 workspaceId 参数） ---
  {
    path: '/workspace/:workspaceId',
    component: () => import('@/layouts/AppLayout.vue'),
    meta: { requiresAuth: true },
    children: [
      {
        path: '',
        name: 'Dashboard',
        component: () => import('@/views/DashboardView.vue'),
        meta: { title: '概览' },
      },
      {
        path: 'resources',
        name: 'Resources',
        component: () => import('@/views/resources/Index.vue'),
        meta: { title: '资源' },
      },
      {
        path: 'resources/new',
        name: 'CreateResource',
        component: () => import('@/views/resources/Create.vue'),
        meta: { title: '新建资源' },
      },
      {
        path: 'resources/:id',
        name: 'ResourceDetail',
        component: () => import('@/views/resources/Detail.vue'),
        meta: { title: '资源详情' },
      },
      {
        path: 'tools',
        name: 'Tools',
        component: () => import('@/views/tools/Index.vue'),
        meta: { title: '工具' },
      },
      {
        path: 'tools/new',
        name: 'CreateTool',
        component: () => import('@/views/tools/Create.vue'),
        meta: { title: '新建工具' },
      },
      {
        path: 'tools/import',
        name: 'ImportTools',
        component: () => import('@/views/tools/Import.vue'),
        meta: { title: '导入工具' },
      },
      {
        path: 'tools/:id',
        name: 'ToolDetail',
        component: () => import('@/views/tools/Detail.vue'),
        meta: { title: '工具详情' },
      },
      {
        path: 'flows',
        name: 'Flows',
        component: () => import('@/views/flows/Index.vue'),
        meta: { title: '流程' },
      },
      {
        path: 'flows/new',
        name: 'CreateFlow',
        component: () => import('@/views/flows/Create.vue'),
        meta: { title: '新建流程' },
      },
      {
        path: 'flows/:id',
        name: 'FlowCanvas',
        component: () => import('@/views/flows/Canvas.vue'),
        meta: { title: '流程详情' },
      },
      {
        path: 'triggers',
        name: 'Triggers',
        component: () => import('@/views/triggers/Index.vue'),
        meta: { title: '触发器' },
      },
      {
        path: 'triggers/new',
        name: 'CreateTrigger',
        component: () => import('@/views/triggers/Create.vue'),
        meta: { title: '新建触发器' },
      },
      {
        path: 'triggers/:id',
        name: 'TriggerDetail',
        component: () => import('@/views/triggers/Detail.vue'),
        meta: { title: '触发器详情' },
      },
      {
        path: 'runs',
        name: 'Runs',
        component: () => import('@/views/runs/Index.vue'),
        meta: { title: '运行记录' },
      },
      {
        path: 'runs/:id',
        name: 'RunDetail',
        component: () => import('@/views/runs/Detail.vue'),
        meta: { title: '运行详情' },
      },
    ],
  },
  // --- Catch-all ---
  {
    path: '/:pathMatch(.*)*',
    redirect: '/select-workspace',
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

// 路由守卫
router.beforeEach((to, _from, next) => {
  const authStore = useAuthStore()

  // 已登录用户访问登录/激活页，重定向到空间首页
  if ((to.path === '/login' || to.path === '/activation') && authStore.isAuthenticated) {
    const workspaceStore = useWorkspaceStore()
    const wsId = workspaceStore.currentWorkspaceId
    next(wsId ? `/workspace/${wsId}` : '/select-workspace')
    return
  }

  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    next({ path: '/login', query: { redirect: to.fullPath } })
    return
  }

  if (to.meta.requiresAdmin && !authStore.isAdmin) {
    const workspaceStore = useWorkspaceStore()
    const wsId = workspaceStore.currentWorkspaceId || to.params.workspaceId
    next(wsId ? `/workspace/${wsId}` : '/')
    return
  }

  // 已登录但未选择工作空间 → 拦截到选择页（/select-workspace 自身除外）
  // 仅在 workspace 列表已加载过一次后才拦截，避免与 App.vue 初始化竞态
  if (authStore.isAuthenticated && !to.meta.skipWorkspaceCheck) {
    const workspaceStore = useWorkspaceStore()
    if (workspaceStore.loaded && (workspaceStore.workspaces.length === 0 || !workspaceStore.hasCurrentWorkspace)) {
      next({ path: '/select-workspace', query: { redirect: to.fullPath } })
      return
    }
  }

  next()
})

export default router
