<template>
  <aside class="sidebar">
    <!-- 头部：空间切换 -->
    <div class="sidebar-header">
      <button class="ws-switch-btn" @click="showWorkspacePanel = true">
        <span class="ws-switch-avatar">{{ getInitials(currentWorkspace.name) }}</span>
        <span class="ws-switch-info">
          <span class="ws-switch-name">{{ currentWorkspace.name }}</span>
          <span class="ws-switch-hint">{{ t('nav.switchWorkspace') }}</span>
        </span>
        <svg class="ws-switch-arrow h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.25 4.5l7.5 7.5-7.5 7.5" />
        </svg>
      </button>
    </div>

    <!-- 导航菜单 -->
    <nav class="sidebar-nav">
      <div class="nav-section" v-for="(section, index) in navigation" :key="index">
        <div v-if="section.title" class="section-title">
          {{ section.title }}
        </div>
        <ul class="nav-items">
          <li
            v-for="item in section.items"
            :key="item.path"
            class="nav-item"
            :class="{ 'nav-item-active': isActive(item) }"
          >
            <router-link
              :to="item.path"
              class="nav-link"
              @click="closeOnMobile"
            >
              <span class="nav-icon-wrapper">
                <component
                  v-if="item.icon"
                  :is="item.icon"
                  class="nav-icon"
                  :class="{ 'nav-icon-active': isActive(item) }"
                />
                <span v-else class="nav-icon-placeholder"></span>
              </span>
              <span class="nav-label">{{ item.label }}</span>
              <span v-if="item.badge" class="nav-badge">{{ item.badge }}</span>
            </router-link>
          </li>
        </ul>
      </div>
    </nav>

    <!-- 分割线 -->
    <div class="sidebar-divider"></div>

    <!-- 底部：个人资料 + 设置 -->
    <div class="sidebar-footer">
      <!-- 个人资料 -->
      <router-link to="/account" class="profile-link" @click="closeOnMobile">
        <span class="profile-avatar">
          <img
            v-if="user.avatar"
            :src="user.avatar"
            :alt="user.displayName"
            class="avatar-img"
          />
          <span v-else class="avatar-placeholder">{{ getUserInitials(user.displayName) }}</span>
        </span>
        <span class="profile-info">
          <span class="profile-name">{{ user.displayName }}</span>
          <span v-if="user.role" class="profile-role">{{ user.role }}</span>
        </span>
      </router-link>

      <!-- 设置按钮（沉底） -->
      <div class="settings-area">
        <button
          ref="settingsBtnRef"
          class="settings-btn"
          @click="openSettingsMenu"
          :class="{ 'settings-btn-open': showSettingsMenu }"
        >
          <svg class="settings-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
              d="M9.594 3.94c.09-.542.56-.94 1.11-.94h2.593c.55 0 1.02.398 1.11.94l.213 1.281c.063.374.313.686.645.87.074.04.147.083.22.127.324.196.72.257 1.075.124l1.217-.456a1.125 1.125 0 011.37.49l1.296 2.247a1.125 1.125 0 01-.26 1.431l-1.003.827c-.293.24-.438.613-.431.992a6.759 6.759 0 010 .255c-.007.378.138.75.43.99l1.005.828c.424.35.534.954.26 1.43l-1.298 2.247a1.125 1.125 0 01-1.369.491l-1.217-.456c-.355-.133-.75-.072-1.076.124a6.57 6.57 0 01-.22.128c-.331.183-.581.495-.644.869l-.213 1.28c-.09.543-.56.941-1.11.941h-2.594c-.55 0-1.02-.398-1.11-.94l-.213-1.281c-.062-.374-.312-.686-.644-.87a6.52 6.52 0 01-.22-.127c-.325-.196-.72-.257-1.076-.124l-1.217.456a1.125 1.125 0 01-1.369-.49l-1.297-2.247a1.125 1.125 0 01.26-1.431l1.004-.827c.292-.24.437-.613.43-.992a6.932 6.932 0 010-.255c.007-.378-.138-.75-.43-.99l-1.004-.828a1.125 1.125 0 01-.26-1.43l1.297-2.247a1.125 1.125 0 011.37-.491l1.216.456c.356.133.751.072 1.076-.124.072-.044.146-.087.22-.128.332-.183.582-.495.644-.869l.214-1.281z" />
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
              d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
          </svg>
          <span class="settings-label">{{ t('common.settings') }}</span>
          <svg
            class="settings-chevron h-4 w-4"
            fill="none" viewBox="0 0 24 24" stroke="currentColor"
          >
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.25 4.5l7.5 7.5-7.5 7.5" />
          </svg>
        </button>
      </div>
    </div>

    <!-- 设置面板：Teleport 到 body，完全脱离 sidebar 布局 -->
    <Teleport to="body">
      <transition name="fade">
        <div
          v-if="showSettingsMenu"
          class="settings-panel"
          :style="settingsPanelStyle"
        >
          <div class="settings-panel-header">
            <span class="settings-panel-title">{{ t('common.settings') }}</span>
          </div>
          <div class="settings-panel-list">
            <router-link
              to="/workspaces"
              class="settings-panel-item"
              @click="showSettingsMenu = false; closeOnMobile()"
            >
              <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.25 21h19.5m-18-18v18m10.5-18v18m6-13.5V21M6.75 6.75h.75m-.75 3h.75m-.75 3h.75m3-6h.75m-.75 3h.75m-.75 3h.75M6.75 21v-3.375c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125V21M3 3h12m-.75 4.5H21m-3.75 3h.008v.008h-.008v-.008zm0 3h.008v.008h-.008v-.008zm0 3h.008v.008h-.008v-.008z" />
              </svg>
              <span>{{ t('nav.manageWorkspaces') }}</span>
            </router-link>
            <router-link
              v-if="authStore.isAdmin"
              to="/admin/accounts"
              class="settings-panel-item"
              @click="showSettingsMenu = false; closeOnMobile()"
            >
              <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19.128a9.38 9.38 0 002.625.372 9.337 9.337 0 004.121-.952 4.125 4.125 0 00-7.533-2.493M15 19.128v-.003c0-1.113-.285-2.16-.786-3.07M15 19.128v.106A12.318 12.318 0 018.624 21c-2.331 0-4.512-.645-6.374-1.766l-.001-.109a6.375 6.375 0 0111.964-3.07M12 6.375a3.375 3.375 0 11-6.75 0 3.375 3.375 0 016.75 0zm8.25 2.25a2.625 2.625 0 11-5.25 0 2.625 2.625 0 015.25 0z" />
              </svg>
              <span>{{ t('nav.account') }}</span>
            </router-link>
          </div>
        </div>
      </transition>
    </Teleport>

    <!-- 紧贴侧边栏右侧弹出：空间列表 -->
    <transition name="slide-panel">
      <div v-if="showWorkspacePanel" class="ws-panel">
        <div class="ws-panel-header">
          <h3 class="ws-panel-title">{{ t('nav.workspace') }}</h3>
          <button class="ws-panel-close" @click="showWorkspacePanel = false">
            <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
        <div class="ws-panel-list">
          <button
            v-for="workspace in workspaces"
            :key="workspace.id"
            class="ws-panel-item"
            :class="{ 'ws-panel-item-active': workspace.id === currentWorkspace.id }"
            @click="switchWorkspace(workspace)"
          >
            <span class="ws-panel-item-avatar">{{ getInitials(workspace.name) }}</span>
            <div class="ws-panel-item-info">
              <span class="ws-panel-item-name">{{ workspace.name }}</span>
              <span class="ws-panel-item-desc">{{ workspace.description || '' }}</span>
            </div>
            <svg v-if="workspace.id === currentWorkspace.id" class="ws-panel-item-check h-5 w-5 text-primary-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12.75L11.25 15 15 9.75M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </button>
        </div>
        <div class="ws-panel-footer">
          <button class="ws-panel-manage" @click="goToWorkspaces">
            {{ t('nav.manageWorkspaces') }}
          </button>
        </div>
      </div>
    </transition>
  </aside>
</template>

<script setup lang="ts">
import { ref, computed, reactive, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useWindowSize } from '@vueuse/core'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import { useWorkspaceStore } from '@/stores/workspace'
import { useWorkspaceRoute } from '@/composables/useWorkspaceRoute'

const props = defineProps<{
  navigation: Array<{
    title?: string
    items: Array<{
      path: string
      label: string
      icon?: any
      badge?: number | string
    }>
  }>
}>()

const emit = defineEmits<{
  'toggle': [collapsed: boolean]
}>()

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const { width } = useWindowSize()
const authStore = useAuthStore()
const workspaceStore = useWorkspaceStore()
const { dashboardPath } = useWorkspaceRoute()

const isMobile = computed(() => width.value < 768)
const showWorkspacePanel = ref(false)
const showSettingsMenu = ref(false)
const settingsBtnRef = ref<HTMLElement | null>(null)

const settingsPanelStyle = reactive({
  left: '256px',
  top: 'auto',
  bottom: '40px',
})

const openSettingsMenu = () => {
  showSettingsMenu.value = !showSettingsMenu.value
  if (!showSettingsMenu.value) return
  const btn = settingsBtnRef.value
  if (!btn) return
  const sidebar = btn.closest('.sidebar') as HTMLElement
  const rect = btn.getBoundingClientRect()
  const sidebarRect = sidebar?.getBoundingClientRect()
  // 面板紧贴侧边栏右边缘，不包含按钮内边距
  settingsPanelStyle.left = `${sidebarRect ? sidebarRect.right : rect.right}px`
  settingsPanelStyle.top = 'auto'
  settingsPanelStyle.bottom = `${window.innerHeight - rect.bottom}px`
}

const workspaces = computed(() => workspaceStore.workspaces)
const currentWorkspace = computed(() =>
  workspaceStore.currentWorkspace || { id: '', name: 'No Workspace' }
)

const user = computed(() => ({
  displayName: authStore.displayName || 'User',
  avatar: null,
  role: authStore.isAdmin ? 'Administrator' : 'User',
}))

const isActive = (item: any) => {
  if (route.path === item.path) return true
  return route.path.startsWith(item.path + '/')
}

const closeOnMobile = () => {
  if (isMobile.value) {
    emit('toggle', true)
  }
}

const switchWorkspace = (workspace: any) => {
  workspaceStore.setCurrentWorkspace(workspace.id)
  router.push(`/workspace/${workspace.id}`)
  showWorkspacePanel.value = false
}

const goToWorkspaces = () => {
  showWorkspacePanel.value = false
  router.push('/workspaces')
}

const getInitials = (name: string) => {
  if (!name) return ''
  return name.split(' ').map(p => p[0]).join('').toUpperCase().substring(0, 2)
}

const getUserInitials = (name: string) => {
  if (!name) return ''
  return name.split(' ').map(p => p[0]).join('').toUpperCase().substring(0, 2)
}

const handleClickOutside = (e: MouseEvent) => {
  const target = e.target as HTMLElement
  // 关闭设置子菜单
  if (showSettingsMenu.value && !target.closest('.settings-area') && !target.closest('.settings-panel')) {
    showSettingsMenu.value = false
  }
  // 关闭空间面板（点击侧边栏外部）
  if (showWorkspacePanel.value && !target.closest('.sidebar') && !target.closest('.ws-panel')) {
    showWorkspacePanel.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

<style scoped>
.sidebar {
  @apply flex flex-col bg-white border-r border-gray-200;
  width: 16rem;
  height: 100vh;
  position: fixed;
  top: 0;
  left: 0;
  z-index: 30;
  overflow-y: auto;
}

/* === 头部：空间切换 === */
.sidebar-header {
  @apply flex items-center px-4 h-16 border-b border-gray-200;
}

.ws-switch-btn {
  @apply flex items-center w-full px-3 py-2 rounded-lg text-left transition-colors;
  @apply hover:bg-gray-100;
}

.ws-switch-avatar {
  @apply flex-shrink-0 h-9 w-9 rounded-lg bg-gradient-to-br from-primary-500 to-secondary-500;
  @apply flex items-center justify-center text-white font-semibold text-sm;
}

.ws-switch-info {
  @apply ml-3 flex-1 min-w-0;
}

.ws-switch-name {
  @apply block text-sm font-semibold text-gray-900 truncate;
}

.ws-switch-hint {
  @apply block text-xs text-gray-500 truncate;
}

.ws-switch-arrow {
  @apply flex-shrink-0 ml-1 text-gray-400;
}

/* === 导航菜单 === */
.sidebar-nav {
  @apply flex-1 py-3 overflow-y-auto;
}

.nav-section {
  @apply mb-4;
}

.section-title {
  @apply px-5 mb-1.5 text-xs font-medium text-gray-400 uppercase tracking-wide;
}

.nav-items {
  @apply space-y-0.5;
}

.nav-item {
  @apply relative;
}

.nav-link {
  @apply flex items-center px-5 py-2 text-sm font-medium transition-colors duration-150;
  @apply text-gray-600 hover:text-gray-900 hover:bg-gray-50 rounded-r-lg;
}

.nav-item-active > .nav-link {
  @apply bg-primary-50 text-primary-700;
}

.nav-icon-wrapper {
  @apply flex-shrink-0;
}

.nav-icon {
  @apply h-5 w-5 transition-colors;
}

.nav-icon-active {
  @apply text-primary-600;
}

.nav-icon-placeholder {
  @apply h-5 w-5;
}

.nav-label {
  @apply ml-3 truncate;
}

.nav-badge {
  @apply ml-auto mr-3 px-2 py-0.5 text-xs font-semibold rounded-full;
  @apply bg-gray-100 text-gray-800;
}

.nav-item-active > .nav-link .nav-badge {
  @apply bg-primary-100 text-primary-800;
}

/* === 分割线 === */
.sidebar-divider {
  @apply mx-4 border-t border-gray-100;
}

/* === 底部 === */
.sidebar-footer {
  @apply p-2 mt-auto;
}

/* 个人资料 */
.profile-link {
  @apply flex items-center px-3 py-2 rounded-lg text-left no-underline transition-colors;
  @apply text-gray-700 hover:bg-gray-50 hover:no-underline;
}

.profile-avatar {
  @apply flex-shrink-0 h-8 w-8 rounded-full bg-gradient-to-br from-primary-400 to-secondary-400;
  @apply flex items-center justify-center text-white font-medium text-xs;
  @apply overflow-hidden;
}

.avatar-img {
  @apply h-full w-full object-cover;
}

.avatar-placeholder {
  @apply text-white;
}

.profile-info {
  @apply ml-3 flex-1 min-w-0;
}

.profile-name {
  @apply block text-sm font-medium text-gray-900 truncate;
}

.profile-role {
  @apply block text-xs text-gray-500 truncate;
}

/* 设置按钮 - 不用 relative，让面板定位到 sidebar */
.settings-area {
  @apply mt-0.5;
}

.settings-btn {
  @apply flex items-center w-full px-3 py-2 rounded-lg text-sm transition-colors;
  @apply text-gray-500 hover:text-gray-700 hover:bg-gray-50;
}

.settings-btn-open {
  @apply text-gray-700 bg-gray-50;
}

.settings-icon {
  @apply h-5 w-5;
}

.settings-label {
  @apply ml-3 flex-1 text-left;
}

.settings-chevron {
  @apply text-gray-400;
}

/* 设置面板样式移到非 scoped style（Teleport 到 body） */





/* === 紧贴侧边栏右侧弹出：空间列表 === */
.ws-panel {
  @apply absolute bg-white shadow-xl z-50 flex flex-col;
  left: 100%; /* 紧贴侧边栏右边缘 */
  top: 0;
  width: 18rem;
  max-height: 100vh;
  border-top-right-radius: 0.5rem;
  border-bottom-right-radius: 0.5rem;
}

.ws-panel-header {
  @apply flex items-center justify-between px-6 py-5 border-b border-gray-100;
}

.ws-panel-title {
  @apply text-lg font-semibold text-gray-900;
}

.ws-panel-close {
  @apply p-1 rounded-md text-gray-400 hover:text-gray-600 hover:bg-gray-100;
}

.ws-panel-list {
  @apply flex-1 overflow-y-auto py-2;
}

.ws-panel-item {
  @apply flex items-center w-full px-6 py-3 text-left transition-colors;
  @apply hover:bg-gray-50;
}

.ws-panel-item-active {
  @apply bg-primary-50;
}

.ws-panel-item-avatar {
  @apply flex-shrink-0 h-10 w-10 rounded-lg bg-gradient-to-br from-primary-500 to-secondary-500;
  @apply flex items-center justify-center text-white font-semibold text-sm;
}

.ws-panel-item-active .ws-panel-item-avatar {
  @apply from-primary-600 to-secondary-600;
}

.ws-panel-item-info {
  @apply ml-3 flex-1 min-w-0;
}

.ws-panel-item-name {
  @apply block text-sm font-medium text-gray-900 truncate;
}

.ws-panel-item-desc {
  @apply block text-xs text-gray-500 truncate mt-0.5;
}

.ws-panel-item-check {
  @apply flex-shrink-0;
}

.ws-panel-footer {
  @apply px-6 py-4 border-t border-gray-100;
}

.ws-panel-manage {
  @apply w-full px-4 py-2.5 text-sm font-medium text-gray-600 rounded-lg;
  @apply hover:bg-gray-100 hover:text-gray-900 transition-colors;
}

/* === 动画 === */
.slide-panel-enter-active,
.slide-panel-leave-active {
  transition: opacity 0.15s ease;
}

.slide-panel-enter-from,
.slide-panel-leave-to {
  opacity: 0;
}

.slide-panel-enter-to,
.slide-panel-leave-from {
  opacity: 1;
}

/* === 响应式 === */
@media (max-width: 767px) {
  .sidebar {
    position: fixed;
    top: 0;
    left: 0;
    height: 100vh;
    z-index: 40;
    box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1);
  }
}
</style>

<!-- 非 scoped 样式：Teleport 到 body 的元素无法被 scoped 选中 -->
<style>
.settings-panel {
  position: fixed;
  z-index: 999;
  background: white;
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 8px 10px -6px rgba(0, 0, 0, 0.1);
  border-radius: 0.5rem;
  overflow: hidden;
  width: 13.5rem;
}

.settings-panel-header {
  padding: 0.625rem 1rem;
  border-bottom: 1px solid #f3f4f6;
}

.settings-panel-title {
  font-size: 0.75rem;
  font-weight: 500;
  color: #9ca3af;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.settings-panel-list {
  padding: 0.25rem 0;
}

.settings-panel-item {
  display: flex;
  align-items: center;
  gap: 0.625rem;
  padding: 0.625rem 1rem;
  font-size: 0.875rem;
  color: #4b5563;
  text-decoration: none;
  transition: all 0.15s ease;
}

.settings-panel-item:hover {
  background-color: #f9fafb;
  color: #111827;
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.15s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
