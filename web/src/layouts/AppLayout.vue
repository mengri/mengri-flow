<script setup lang="ts">
import { computed, ref, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useWindowSize } from '@vueuse/core'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import { useAuth } from '@/composables/useAuth'
import { useWorkspaceRoute } from '@/composables/useWorkspaceRoute'

// 组件导入
import MengriSidebar from '@/components/ui/MengriSidebar.vue'

// 图标导入
import {
  HomeIcon,
  ArrowsRightLeftIcon,
  ChartBarIcon,
  PuzzleIcon,
  CogIcon,
  UsersIcon,
  UserCircleIcon,
  FolderIcon,
} from '@/components/icons'

// 窗口尺寸
const { width } = useWindowSize()
const isMobile = computed(() => width.value < 768)

const { t } = useI18n()

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const { handleLogout: authLogout } = useAuth()
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

// 面包屑导航
const breadcrumbs = computed(() => {
  const pathArray = route.path.split('/').filter(Boolean)
  // 跳过 workspaceId 段（/workspace/:workspaceId/... 中的第2段）
  const crumbs: Array<{ path: string; label: string }> = [{ path: dashboardPath(), label: t('nav.dashboard') }]

  // pathArray 格式: ['workspace', '<wsId>', 'flows', '123', ...]
  // 跳过前两段（workspace + workspaceId）
  let currentPath = `/workspace/${route.params.workspaceId || ''}`
  for (let i = 2; i < pathArray.length; i++) {
    currentPath += `/${pathArray[i]}`
    const routeName = route.matched[i]?.meta?.title ||
                     pathArray[i].charAt(0).toUpperCase() + pathArray[i].slice(1)

    crumbs.push({
      path: currentPath,
      label: (routeName as string) || pathArray[i],
    })
  }

  return crumbs
})

// 页面标题
const pageTitle = computed(() => route.meta?.title || getRouteTitle(route.path))
const pageSubtitle = computed(() => route.meta?.subtitle || '')
const showFooter = computed(() => route.meta?.showFooter !== false)

// 侧边栏导航配置
const navigation = computed(() => {
  const sections = [
    {
      title: t('nav.workspace'),
      items: [
        { path: dashboardPath(), label: t('nav.dashboard'), icon: HomeIcon },
        { path: flowsPath(), label: t('nav.flows'), icon: ArrowsRightLeftIcon },
        { path: triggersPath(), label: t('nav.triggers'), icon: CogIcon },
        { path: resourcesPath(), label: t('nav.resources'), icon: PuzzleIcon },
        { path: toolsPath(), label: t('nav.tools'), icon: CogIcon },
      ]
    },
    {
      title: t('common.settings'),
      items: [
        { path: '/workspaces', label: t('nav.manageWorkspaces'), icon: FolderIcon },
      ]
    },
    {
      title: t('nav.runs'),
      items: [
        { path: runsPath(), label: t('nav.runList'), icon: ChartBarIcon },
      ]
    }
  ]

  if (authStore.isAdmin) {
    sections.push({
      title: t('common.settings'),
      items: [
        { path: '/admin/accounts', label: t('nav.account'), icon: UsersIcon },
        { path: '/account', label: t('nav.accountSettings'), icon: UserCircleIcon },
      ]
    })
  }

  return sections
})

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

const handleLogout = () => {
  authLogout()
}

const openSettings = () => {
  console.log('Open settings dialog')
}

const switchWorkspace = (workspaceId: string) => {
  router.push(`/workspace/${workspaceId}`)
}

const getRouteTitle = (path: string): string => {
  const titleMap: Record<string, string> = {
    '/dashboard': t('nav.dashboard'),
    '/workflows': t('nav.flows'),
    '/templates': t('nav.flows'),
    '/analytics': t('nav.runs'),
    '/account': t('nav.accountSettings'),
    '/admin/users': t('nav.account'),
    '/admin/settings': t('common.settings'),
  }

  return titleMap[path] || path.split('/').pop()?.replace(/-/g, ' ') || 'Page'
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
  <div class="min-h-screen flex flex-col bg-gray-50">
    <div class="flex flex-1 overflow-hidden">
      <!-- 侧边栏导航 -->
      <transition name="slide-left">
        <MengriSidebar
          v-if="showSidebar || !isMobile"
          :navigation="navigation"
          :class="{ 'mobile-sidebar': isMobile }"
          @toggle="onSidebarToggle"
          @open-settings="openSettings"
          @workspace-change="switchWorkspace"
        />
      </transition>
      
      <!-- 主要内容区 -->
      <main class="flex-1 overflow-auto p-4 md:p-6">
        <!-- 面包屑导航 -->
        <div v-if="breadcrumbs.length > 1" class="mb-6">
          <nav class="flex items-center space-x-2 text-sm">
            <router-link
              v-for="crumb in breadcrumbs"
              :key="crumb.path"
              :to="crumb.path"
              class="text-gray-500 hover:text-gray-700 transition-colors"
              :class="{ 'text-gray-900 font-medium': $route.path === crumb.path }"
            >
              {{ crumb.label }}
              <span v-if="crumb.path !== breadcrumbs[breadcrumbs.length - 1].path" class="ml-2">/</span>
            </router-link>
          </nav>
        </div>
        
        <!-- 页面标题区 -->
        <div v-if="pageTitle" class="mb-6">
          <h1 class="text-2xl md:text-3xl font-bold text-gray-900 mb-2">{{ pageTitle }}</h1>
          <p v-if="pageSubtitle" class="text-gray-600 text-base md:text-lg">{{ pageSubtitle }}</p>
        </div>
        
        <!-- 内容容器 -->
        <div class="content-container">
          <RouterView />
        </div>
        
        <!-- 页脚 -->
        <footer v-if="showFooter" class="mt-12 pt-6 border-t border-gray-200">
          <div class="flex flex-col md:flex-row justify-between items-center">
            <div class="text-sm text-gray-600 mb-4 md:mb-0">
              <span class="font-semibold text-gray-900">Mengri Flow</span> © {{ currentYear }}. All rights reserved.
            </div>
            <div class="flex items-center space-x-6">
              <a href="/privacy" class="text-sm text-gray-600 hover:text-gray-900">{{ t('common.settings') }}</a>
              <a href="/terms" class="text-sm text-gray-600 hover:text-gray-900">{{ t('common.help') }}</a>
              <a href="/contact" class="text-sm text-gray-600 hover:text-gray-900">Contact</a>
            </div>
          </div>
        </footer>
      </main>
    </div>
    
  </div>
</template>

<style scoped>
.content-container {
  @apply max-w-full mx-auto;
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
