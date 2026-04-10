<script setup lang="ts">
import { computed, ref, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useWindowSize } from '@vueuse/core'
import { useI18n } from 'vue-i18n'
import { useWorkspaceStore } from '@/stores/workspace'
import { useWorkspaceRoute } from '@/composables/useWorkspaceRoute'

// 组件导入
import MengriSidebar from '@/components/ui/MengriSidebar.vue'

// 图标导入
import {
  HomeIcon,
  ArrowsRightLeftIcon,
  QueueListIcon,
  PlayIcon,
  PuzzleIcon,
  WrenchScrewdriverIcon,
} from '@/components/icons'

// 窗口尺寸
const { width } = useWindowSize()
const isMobile = computed(() => width.value < 768)

const { t } = useI18n()

const route = useRoute()
const router = useRouter()
const workspaceStore = useWorkspaceStore()
const {
  dashboardPath,
  flowsPath,
  triggersPath,
  resourcesPath,
  toolsPath,
  runsPath,
} = useWorkspaceRoute()

// 响应式状态
const showSidebar = ref(true)

// 面包屑导航 - 所有页面都显示面包屑
const breadcrumbs = computed(() => {
  const workspaceName = workspaceStore.currentWorkspace?.name || ''
  const crumbs: Array<{ path: string; label: string; isLast?: boolean }> = []

  const pathArray = route.path.split('/').filter(Boolean)

  // === 非工作区路由 ===
  const isWorkspaceRoute = pathArray[0] === 'workspace'

  if (!isWorkspaceRoute) {
    // 设置相关路由
    if (route.path.startsWith('/admin')) {
      crumbs.push(
        { path: '/admin/accounts', label: t('common.settings') },
        { path: '/admin/accounts', label: t('nav.account'), isLast: route.path === '/admin/accounts' },
      )
      return crumbs
    }
    if (route.path.startsWith('/workspaces')) {
      crumbs.push(
        { path: '/workspaces', label: t('common.settings') },
        { path: '/workspaces', label: t('nav.manageWorkspaces'), isLast: true },
      )
      return crumbs
    }
    if (route.path.startsWith('/account')) {
      crumbs.push(
        { path: '/account', label: 'Profile', isLast: true },
      )
      return crumbs
    }
    return crumbs
  }

  // === 工作区内路由 ===
  const workspaceId = route.params.workspaceId as string
  const wsBasePath = workspaceId ? `/workspace/${workspaceId}` : dashboardPath()

  // 如果在仪表板（概览）页面
  if (pathArray.length <= 2 || (pathArray.length === 3 && pathArray[2] === '')) {
    crumbs.push(
      { path: wsBasePath, label: workspaceName },
      { path: route.path, label: t('nav.dashboard'), isLast: true },
    )
    return crumbs
  }

  // 其他工作区页面
  crumbs.push({ path: wsBasePath, label: workspaceName })

  const segmentLabels: Record<string, string> = {
    flows: t('nav.flows'),
    triggers: t('nav.triggers'),
    resources: t('nav.resources'),
    tools: t('nav.tools'),
    runs: t('nav.runs'),
    new: t('common.create'),
    import: t('common.import'),
  }

  let currentPath = wsBasePath
  for (let i = 2; i < pathArray.length; i++) {
    currentPath += `/${pathArray[i]}`
    const isLast = i === pathArray.length - 1
    const label = (route.matched[i]?.meta?.title as string)
      || segmentLabels[pathArray[i]]
      || pathArray[i].charAt(0).toUpperCase() + pathArray[i].slice(1)

    crumbs.push({
      path: isLast ? route.path : currentPath,
      label: label || pathArray[i],
      isLast,
    })
  }

  return crumbs
})

// 页面标题（由各 view 自行管理，layout 不再全局显示）
const showFooter = computed(() => route.meta?.showFooter !== false)

// 侧边栏导航配置
const navigation = computed(() => [
  {
    title: t('nav.navigation'),
    items: [
      { path: dashboardPath(), label: t('nav.dashboard'), icon: HomeIcon },
      { path: resourcesPath(), label: t('nav.resources'), icon: PuzzleIcon },
      { path: flowsPath(), label: t('nav.flows'), icon: ArrowsRightLeftIcon },
    ],
  },
  {
    title: t('nav.integration'),
    items: [
      { path: triggersPath(), label: t('nav.triggers'), icon: PlayIcon },
      { path: toolsPath(), label: t('nav.tools'), icon: WrenchScrewdriverIcon },
      { path: runsPath(), label: t('nav.runList'), icon: QueueListIcon },
    ],
  },
])

// 当前年份（用于页脚）
const currentYear = computed(() => new Date().getFullYear())

// 方法
const toggleSidebar = () => {
  showSidebar.value = !showSidebar.value
}

const onSidebarToggle = (collapsed: boolean) => {
  if (isMobile.value) {
    showSidebar.value = !collapsed
  }
}

const switchWorkspace = (workspaceId: string) => {
  router.push(`/workspace/${workspaceId}`)
}

// 生命周期
onMounted(() => {
  if (isMobile.value) {
    showSidebar.value = false
  }
})

onUnmounted(() => {
  // 清理逻辑
})
</script>

<template>
  <div class="h-screen flex flex-col bg-gray-50 overflow-hidden">
    <div class="flex flex-1 overflow-hidden">
      <!-- 侧边栏导航 -->
      <transition name="slide-left">
        <MengriSidebar
          v-if="showSidebar || !isMobile"
          :navigation="navigation"
          :class="{ 'mobile-sidebar': isMobile }"
          @toggle="onSidebarToggle"
        />
      </transition>
      
      <!-- 右侧区域 -->
      <div class="flex-1 flex flex-col overflow-hidden" :class="{ 'main-with-sidebar': showSidebar || !isMobile }">
        <!-- 面包屑导航 - 固定顶部不滚动 -->
        <div v-if="breadcrumbs.length > 0" class="breadcrumb-bar">
          <nav class="flex items-center space-x-1.5 text-sm">
            <template v-for="(crumb, index) in breadcrumbs" :key="crumb.path">
              <span v-if="index > 0" class="text-gray-300 select-none">/</span>
              <router-link
                v-if="!crumb.isLast"
                :to="crumb.path"
                class="text-gray-500 hover:text-gray-700 transition-colors"
              >
                {{ crumb.label }}
              </router-link>
              <span v-else class="text-gray-900 font-medium">{{ crumb.label }}</span>
            </template>
          </nav>
        </div>

        <!-- 滚动包裹层：负责滚动，无 padding/margin -->
        <div class="main-scroll-area">
          <!-- 内容容器：padding 在这里 -->
          <div class="content-container p-4 md:p-6">
            <RouterView />
          </div>
        </div>

        <!-- 页脚 - 固定底部不参与滚动 -->
        <footer v-if="showFooter" class="footer-bar">
          <div class="text-sm text-gray-600">
            <span class="font-semibold text-gray-900">Mengri Flow</span> © {{ currentYear }}. All rights reserved.
          </div>
        </footer>
      </div>
    </div>
    
  </div>
</template>

<style scoped>
/* 滚动包裹层：始终显示滚动条轨道，避免内容宽度抖动 */
.main-scroll-area {
  flex: 1;
  overflow-y: scroll;
  scrollbar-gutter: stable;
}

.content-container {
  @apply max-w-7xl mx-auto;
}

/* 面包屑导航栏 - 白色底色 */
.breadcrumb-bar {
  @apply px-4 md:px-6 py-3 bg-white border-b border-gray-200 flex-shrink-0;
}

/* 页脚 - 固定底部不滚动 */
.footer-bar {
  @apply px-4 md:px-6 py-3 bg-white border-t border-gray-200 flex-shrink-0;
  text-align: right;
}

/* 有侧边栏时内容区左边距 */
.main-with-sidebar {
  margin-left: 16rem;
}

/* 移动端侧边栏样式 */
.mobile-sidebar {
  position: fixed;
  top: 0;
  left: 0;
  height: 100vh;
  z-index: 40;
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 8px 10px -6px rgba(0, 0, 0, 0.1);
}

/* 侧边栏过渡动画 */
.slide-left-enter-active,
.slide-left-leave-active {
  transition: transform 0.3s ease;
}

.slide-left-enter-from {
  transform: translateX(-100%);
}

.slide-left-leave-to {
  transform: translateX(-100%);
}

/* 响应式调整 */
@media (max-width: 768px) {
  .content-container {
    @apply px-0;
  }
}

@media (min-width: 768px) {
  .content-container {
    @apply max-w-7xl;
  }
}

/* 打印样式优化 */
@media print {
  .mobile-sidebar {
    display: none !important;
  }
  
  .content-container {
    margin: 0;
    padding: 0;
    max-width: none;
  }
}
</style>
