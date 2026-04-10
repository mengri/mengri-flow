<template>
  <aside :class="['sidebar', { expanded: isExpanded, collapsed: isCollapsed }]">
    <!-- 侧边栏头部 -->
    <div class="sidebar-header">
      <router-link v-if="isExpanded" :to="dashboardPath()" class="sidebar-brand">
        <AppLogo size="md" />
      </router-link>

      <router-link v-else :to="dashboardPath()" class="sidebar-brand-collapsed">
        <AppLogo size="sm" />
      </router-link>
      
      <button
        v-if="isExpanded"
        class="sidebar-toggle"
        @click="toggleCollapse"
        :aria-label="'Collapse sidebar'"
      >
        <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M11 19l-7-7 7-7m8 14l-7-7 7-7"
          />
        </svg>
      </button>
    </div>
    
    <!-- 侧边栏导航 -->
    <nav class="sidebar-nav">
      <div class="nav-section" v-for="(section, index) in navigation" :key="index">
        <div v-if="section.title && isExpanded" class="section-title">
          {{ section.title }}
        </div>
        
        <ul class="nav-items">
          <li
            v-for="item in section.items"
            :key="item.path"
            class="nav-item"
            :class="{
              'nav-item-active': isActive(item),
              'nav-item-has-child': item.children,
              'nav-item-collapsed': isCollapsed,
            }"
          >
            <!-- 顶级菜单项 -->
            <component
              :is="item.children ? 'div' : 'router-link'"
              :to="item.children ? undefined : item.path"
              :class="[
                'nav-link',
                {
                  'has-children': item.children,
                  'has-action': item.action,
                  'has-badge': item.badge,
                },
              ]"
              @click="item.children ? toggleSubMenu(item) : undefined"
            >
              <!-- 图标 -->
              <span class="nav-icon-wrapper">
                <component
                  v-if="item.icon"
                  :is="item.icon"
                  class="nav-icon"
                  :class="{ 'nav-icon-active': isActive(item) }"
                />
                <span v-else class="nav-icon-placeholder"></span>
              </span>
              
              <!-- 标签文本 -->
              <transition name="slide-fade" mode="out-in">
                <span v-if="isExpanded" class="nav-label">
                  {{ item.label }}
                </span>
              </transition>
              
              <!-- 徽章 -->
              <transition name="fade">
                <span v-if="item.badge && isExpanded" class="nav-badge">
                  {{ item.badge }}
                </span>
              </transition>
              
              <!-- 子菜单指示器 -->
              <span
                v-if="item.children && isExpanded"
                class="nav-chevron"
                :class="{ 'rotate-90': isSubMenuOpen(item) }"
              >
                <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
                </svg>
              </span>
              
              <!-- 动作按钮 -->
              <button
                v-if="item.action && isExpanded"
                class="nav-action"
                @click.stop="handleAction(item.action)"
                :aria-label="`${item.label} action`"
              >
                <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
                </svg>
              </button>
            </component>
            
            <!-- 子菜单 -->
            <transition name="expand">
              <div
                v-if="item.children && isSubMenuOpen(item) && isExpanded"
                class="submenu"
              >
                <ul class="submenu-items">
                  <li
                    v-for="child in item.children"
                    :key="child.path"
                    class="submenu-item"
                    :class="{ 'submenu-item-active': $route.path === child.path }"
                  >
                    <router-link
                      :to="child.path"
                      class="submenu-link"
                      @click="closeOnMobile"
                    >
                      <span class="submenu-icon">
                        <component
                          v-if="child.icon"
                          :is="child.icon"
                          class="h-4 w-4"
                        />
                        <span v-else class="submenu-dot"></span>
                      </span>
                      <span class="submenu-label">{{ child.label }}</span>
                      <span v-if="child.badge" class="submenu-badge">
                        {{ child.badge }}
                      </span>
                    </router-link>
                  </li>
                </ul>
              </div>
            </transition>
          </li>
        </ul>
      </div>
    </nav>
    
    <!-- 侧边栏底部 -->
    <div v-if="isExpanded" class="sidebar-footer">
      <!-- 工作空间切换 -->
      <div v-if="workspaces.length > 0" class="workspace-selector">
        <button
          class="workspace-button"
          @click="toggleWorkspaceMenu"
          :aria-label="`Switch workspace: ${currentWorkspace.name}`"
          aria-haspopup="true"
          :aria-expanded="showWorkspaceMenu"
        >
          <span class="workspace-avatar">
            {{ getInitials(currentWorkspace.name) }}
          </span>
          <span class="workspace-info">
            <span class="workspace-name">{{ currentWorkspace.name }}</span>
            <span class="workspace-type">{{ currentWorkspace.status }}</span>
          </span>
          <svg class="workspace-chevron h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
          </svg>
        </button>
        
        <!-- 工作空间菜单 -->
        <transition name="slide-down">
          <div
            v-if="showWorkspaceMenu"
            class="workspace-menu"
            ref="workspaceMenu"
          >
            <div class="workspace-menu-header">
              <h4 class="workspace-menu-title">Switch Workspace</h4>
              <button
                class="workspace-menu-close"
                @click="showWorkspaceMenu = false"
                aria-label="Close workspace menu"
              >
                <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
            
            <div class="workspace-list">
              <button
                v-for="workspace in workspaces"
                :key="workspace.id"
                class="workspace-item"
                :class="{ 'workspace-item-active': workspace.id === currentWorkspace.id }"
                @click="switchWorkspace(workspace)"
              >
                <span class="workspace-item-avatar">
                  {{ getInitials(workspace.name) }}
                </span>
                <span class="workspace-item-info">
                  <span class="workspace-item-name">{{ workspace.name }}</span>
                  <span class="workspace-item-type">{{ workspace.status }}</span>
                </span>
                <span v-if="workspace.id === currentWorkspace.id" class="workspace-item-check">
                  <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                  </svg>
                </span>
              </button>
            </div>
          </div>
        </transition>
      </div>
      
      <!-- 帮助和设置 -->
      <div class="sidebar-actions">
        <button
          class="sidebar-action"
          @click="showHelp"
          aria-label="Help"
        >
          <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.228 9c.549-1.165 2.03-2 3.772-2 2.21 0 4 1.343 4 3 0 1.4-1.278 2.575-3.006 2.907-.542.104-.994.54-.994 1.093m0 3h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <span v-if="isExpanded" class="sidebar-action-label">Help</span>
        </button>
        
        <button
          class="sidebar-action"
          @click="$emit('open-settings')"
          aria-label="Settings"
        >
          <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
          </svg>
          <span v-if="isExpanded" class="sidebar-action-label">Settings</span>
        </button>
      </div>
    </div>
  </aside>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useWindowSize } from '@vueuse/core'
import AppLogo from '@/components/ui/AppLogo.vue'
import { useWorkspaceStore } from '@/stores/workspace'
import { useWorkspaceRoute } from '@/composables/useWorkspaceRoute'

// Props
const props = defineProps<{
  navigation: Array<{
    title?: string
    items: Array<{
      path: string
      label: string
      icon?: any
      badge?: number | string
      children?: Array<{
        path: string
        label: string
        icon?: any
        badge?: number | string
      }>
      action?: () => void
    }>
  }>
}>()

// Emits
const emit = defineEmits<{
  'open-settings': []
  'toggle': [collapsed: boolean]
  'item-click': [item: any]
  'workspace-change': [workspaceId: string]
}>()

// Route
const route = useRoute()
const router = useRouter()

// Stores
const workspaceStore = useWorkspaceStore()
const { dashboardPath } = useWorkspaceRoute()

// Reactive state
const isCollapsed = ref(false)
const openSubMenus = ref<Set<string>>(new Set())
const showWorkspaceMenu = ref(false)

// Window size
const { width } = useWindowSize()

// Computed
const isMobile = computed(() => width.value < 768)
const isExpanded = computed(() => !isCollapsed.value)

const workspaces = computed(() => workspaceStore.workspaces)

const currentWorkspace = computed(() =>
  workspaceStore.currentWorkspace || { id: '', name: 'No Workspace', type: '' }
)

// Methods
const toggleCollapse = () => {
  isCollapsed.value = !isCollapsed.value
  emit('toggle', isCollapsed.value)
}

const isActive = (item: any) => {
  if (route.path === item.path) return true
  if (item.children) {
    return item.children.some((child: any) => route.path === child.path)
  }
  return false
}

const isSubMenuOpen = (item: any) => {
  return openSubMenus.value.has(item.path)
}

const toggleSubMenu = (item: any) => {
  if (openSubMenus.value.has(item.path)) {
    openSubMenus.value.delete(item.path)
  } else {
    openSubMenus.value.add(item.path)
  }
}

const closeOnMobile = () => {
  if (isMobile.value) {
    emit('toggle', true)
  }
}

const handleAction = (action: () => void) => {
  action()
}

const toggleWorkspaceMenu = () => {
  showWorkspaceMenu.value = !showWorkspaceMenu.value
}

const switchWorkspace = (workspace: any) => {
  workspaceStore.setCurrentWorkspace(workspace.id)
  // 导航到新空间
  router.push(`/workspace/${workspace.id}`)
  showWorkspaceMenu.value = false
}

const addWorkspace = () => {
  showWorkspaceMenu.value = false
}

const manageWorkspaces = () => {
  showWorkspaceMenu.value = false
}

const showHelp = () => {
  console.log('Show help')
}

const getInitials = (name: string) => {
  if (!name) return ''
  return name
    .split(' ')
    .map(part => part[0])
    .join('')
    .toUpperCase()
    .substring(0, 2)
}

// Event handlers
const handleClickOutside = (event: MouseEvent) => {
  const workspaceMenu = document.querySelector('.workspace-menu')
  
  if (showWorkspaceMenu.value && workspaceMenu && 
      !workspaceMenu.contains(event.target as Node)) {
    showWorkspaceMenu.value = false
  }
}

// Lifecycle
onMounted(() => {
  // 移动端默认折叠
  if (isMobile.value) {
    isCollapsed.value = true
  }
  
  // 监听点击外部事件
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

<style scoped>
.sidebar {
  @apply flex flex-col bg-white border-r border-gray-200 transition-all duration-300 ease-in-out;
  width: var(--sidebar-width, 16rem);
  min-height: calc(100vh - 4rem);
}

.sidebar.collapsed {
  width: 4rem;
}

.sidebar.expanded {
  width: 16rem;
}

/* 侧边栏头部 */
.sidebar-header {
  @apply flex items-center justify-between h-16 px-4 border-b border-gray-200;
}

.sidebar-brand {
  @apply flex items-center gap-2.5 truncate no-underline hover:no-underline;
}

.sidebar-brand-collapsed {
  @apply flex items-center justify-center w-full no-underline hover:no-underline;
}

.sidebar-toggle {
  @apply p-1 rounded text-gray-500 hover:text-gray-700 hover:bg-gray-100;
}

/* 侧边栏导航 */
.sidebar-nav {
  @apply flex-1 py-4 overflow-y-auto;
}

.nav-section {
  @apply mb-6;
}

.section-title {
  @apply px-4 mb-2 text-xs font-semibold text-gray-500 uppercase tracking-wide;
}

.nav-items {
  @apply space-y-1;
}

.nav-item {
  @apply relative;
}

.nav-link {
  @apply flex items-center px-4 py-2.5 text-sm font-medium rounded-r-lg transition-colors duration-200;
  @apply text-gray-700 hover:text-gray-900 hover:bg-gray-100;
}

.nav-item-active > .nav-link {
  @apply bg-primary-50 text-primary-700 border-l-2 border-primary-500;
}

.nav-item-active > .nav-link:hover {
  @apply bg-primary-100;
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
  @apply ml-auto mr-2 px-2 py-0.5 text-xs font-semibold rounded-full;
}

.nav-item-active > .nav-link .nav-badge {
  @apply bg-primary-100 text-primary-800;
}

.nav-item:not(.nav-item-active) > .nav-link .nav-badge {
  @apply bg-gray-100 text-gray-800;
}

.nav-chevron {
  @apply ml-1 transition-transform duration-200;
}

.nav-chevron.rotate-90 {
  transform: rotate(90deg);
}

.nav-action {
  @apply ml-1 p-1 rounded text-gray-400 hover:text-gray-600 hover:bg-gray-200;
}

/* 子菜单样式 */
.submenu {
  @apply mt-0.5 overflow-hidden;
}

.submenu-items {
  @apply ml-9 space-y-0.5 border-l border-gray-200;
}

.submenu-item {
  @apply relative;
}

.submenu-link {
  @apply flex items-center py-1.5 pl-4 pr-3 text-sm rounded-lg transition-colors duration-200;
  @apply text-gray-600 hover:text-gray-900 hover:bg-gray-100;
}

.submenu-item-active .submenu-link {
  @apply text-primary-600 bg-primary-50 font-medium;
}

.submenu-icon {
  @apply mr-2 flex-shrink-0;
}

.submenu-dot {
  @apply h-1.5 w-1.5 rounded-full bg-current;
}

.submenu-label {
  @apply truncate;
}

.submenu-badge {
  @apply ml-auto px-1.5 py-0.5 text-xs font-medium rounded-full bg-gray-100 text-gray-800;
}

.submenu-item-active .submenu-badge {
  @apply bg-primary-100 text-primary-800;
}

/* 侧边栏底部 */
.sidebar-footer {
  @apply border-t border-gray-200 p-4;
}

.workspace-selector {
  @apply relative mb-4;
}

.workspace-button {
  @apply flex items-center w-full p-2 rounded-lg hover:bg-gray-100;
}

.workspace-avatar {
  @apply flex-shrink-0 h-8 w-8 rounded-full bg-gradient-to-br from-primary-500 to-secondary-500 flex items-center justify-center text-white font-semibold text-sm;
}

.workspace-info {
  @apply ml-3 text-left flex-1 min-w-0;
}

.workspace-name {
  @apply block text-sm font-medium text-gray-900 truncate;
}

.workspace-type {
  @apply block text-xs text-gray-500;
}

.workspace-chevron {
  @apply flex-shrink-0 ml-1 text-gray-400;
}

.workspace-menu {
  @apply absolute bottom-full left-0 right-0 mb-2 bg-white border border-gray-200 rounded-lg shadow-lg overflow-hidden z-10;
}

.workspace-menu-header {
  @apply px-4 py-3 border-b border-gray-200 flex items-center justify-between;
}

.workspace-menu-title {
  @apply text-sm font-semibold text-gray-900;
}

.workspace-menu-close {
  @apply p-1 text-gray-400 hover:text-gray-600;
}

.workspace-list {
  @apply max-h-48 overflow-y-auto py-1;
}

.workspace-item {
  @apply flex items-center w-full px-4 py-2.5 text-sm hover:bg-gray-50;
}

.workspace-item-active {
  @apply bg-primary-50;
}

.workspace-item-avatar {
  @apply flex-shrink-0 h-6 w-6 rounded-full bg-gray-200 flex items-center justify-center text-gray-700 font-medium text-xs;
}

.workspace-item-active .workspace-item-avatar {
  @apply bg-primary-100 text-primary-700;
}

.workspace-item-info {
  @apply ml-3 flex-1 min-w-0;
}

.workspace-item-name {
  @apply block text-gray-900 truncate;
}

.workspace-item-type {
  @apply block text-xs text-gray-500;
}

.workspace-item-check {
  @apply flex-shrink-0 ml-2 text-primary-600;
}

.workspace-menu-footer {
  @apply px-4 py-3 border-t border-gray-200 space-y-2;
}

.workspace-add-button {
  @apply flex items-center justify-center w-full gap-2 px-3 py-2 text-sm text-primary-600 hover:text-primary-800 hover:bg-primary-50 rounded;
}

.workspace-manage-button {
  @apply w-full text-sm text-gray-600 hover:text-gray-800 hover:bg-gray-100 rounded px-3 py-2;
}

/* 侧边栏动作按钮 */
.sidebar-actions {
  @apply flex space-x-2;
}

.sidebar-action {
  @apply flex items-center justify-center gap-2 flex-1 px-3 py-2 text-sm text-gray-600 hover:text-gray-800 hover:bg-gray-100 rounded;
}

.sidebar-action-label {
  @apply truncate;
}

/* 折叠状态下的样式 */
.sidebar.collapsed .nav-label,
.sidebar.collapsed .nav-badge,
.sidebar.collapsed .nav-chevron,
.sidebar.collapsed .nav-action,
.sidebar.collapsed .section-title,
.sidebar.collapsed .sidebar-actions,
.sidebar.collapsed .sidebar-footer {
  @apply hidden !important;
}

.sidebar.collapsed .sidebar-brand {
  @apply hidden !important;
}

.sidebar.collapsed .sidebar-toggle {
  @apply hidden !important;
}

.sidebar.collapsed .sidebar-brand-collapsed {
  @apply flex !important;
}

.sidebar.collapsed .sidebar-header {
  @apply justify-center;
}

.sidebar.collapsed .nav-link {
  @apply justify-center px-0;
}

.sidebar.collapsed .nav-link .nav-icon-wrapper {
  @apply ml-0;
}

.sidebar.collapsed .nav-item-active > .nav-link {
  @apply border-l-0 bg-primary-50;
}

/* 工具提示 - 折叠状态下的悬停提示 */
.sidebar.collapsed .nav-link::after {
  content: attr(data-tooltip);
  @apply absolute left-full ml-2 px-2 py-1 text-xs font-medium bg-gray-900 text-white rounded whitespace-nowrap opacity-0 transition-opacity duration-200 pointer-events-none;
  top: 50%;
  transform: translateY(-50%);
}

.sidebar.collapsed .nav-link:hover::after {
  @apply opacity-100;
}

/* 动画效果 */
.slide-fade-enter-active {
  transition: all 0.3s ease;
}

.slide-fade-leave-active {
  transition: all 0.2s cubic-bezier(1, 0.5, 0.8, 1);
}

.slide-fade-enter-from,
.slide-fade-leave-to {
  transform: translateX(-10px);
  opacity: 0;
}

.expand-enter-active {
  transition: height 0.3s ease;
  overflow: hidden;
}

.expand-leave-active {
  transition: height 0.2s cubic-bezier(0.65, 0.05, 0.36, 1);
  overflow: hidden;
}

.expand-enter-from,
.expand-leave-to {
  height: 0 !important;
}

.slide-down-enter-active {
  transition: all 0.2s ease;
}

.slide-down-leave-active {
  transition: all 0.2s cubic-bezier(0.65, 0.05, 0.36, 1);
}

.slide-down-enter-from,
.slide-down-leave-to {
  transform: translateY(10px);
  opacity: 0;
}

/* 响应式调整 */
@media (max-width: 767px) {
  .sidebar {
    position: fixed;
    top: 4rem;
    left: 0;
    height: calc(100vh - 4rem);
    z-index: 30;
    box-shadow: var(--shadow-lg);
  }
  
  .sidebar.collapsed {
    transform: translateX(-100%);
    width: 16rem;
  }
  
  .sidebar.expanded {
    transform: translateX(0);
  }
  
  .sidebar-toggle {
    @apply hidden;
  }
  
  .workspace-selector {
    @apply hidden;
  }
}
</style>