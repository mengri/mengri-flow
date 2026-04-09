<template>
  <nav class="topbar">
    <div class="topbar-container">
      <!-- Logo 区域 -->
      <div class="logo-area">
        <button
          v-if="showMobileMenuToggle"
          class="mobile-menu-toggle"
          @click="$emit('toggle-sidebar')"
          aria-label="Toggle menu"
          aria-expanded="false"
        >
          <svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
          </svg>
        </button>
        
        <div class="logo-wrapper">
          <router-link to="/" class="logo-link">
            <div class="logo-icon">
              <svg class="h-8 w-8" viewBox="0 0 32 32" fill="none" xmlns="http://www.w3.org/2000/svg">
                <defs>
                  <linearGradient id="logo-gradient" x1="0%" y1="0%" x2="100%" y2="100%">
                    <stop offset="0%" stop-color="var(--color-primary-500)" />
                    <stop offset="100%" stop-color="var(--color-secondary-500)" />
                  </linearGradient>
                </defs>
                <path d="M16 2L2 8L16 14L30 8L16 2Z" fill="url(#logo-gradient)" />
                <path d="M30 8L16 14V30L30 24V8Z" fill="var(--color-primary-400)" />
                <path d="M2 8L16 14V30L2 24V8Z" fill="var(--color-secondary-400)" />
              </svg>
            </div>
            <div class="logo-text">
              <span class="logo-text-primary">Mengri</span>
              <span class="logo-text-secondary">Flow</span>
            </div>
          </router-link>
        </div>
        
        <!-- 桌面端主导航 -->
        <nav v-if="!isMobile" class="main-nav">
          <ul class="nav-list">
            <li class="nav-item" v-for="(item, index) in menuItems" :key="item.path">
              <router-link
                :to="item.path"
                class="nav-link"
                :class="{ active: $route.path.startsWith(item.path) }"
                @focus="activeNavIndex = index"
              >
                <component
                  v-if="item.icon"
                  :is="item.icon"
                  class="nav-icon"
                />
                <span class="nav-label">{{ item.label }}</span>
                <span v-if="item.badge" class="nav-badge">{{ item.badge }}</span>
              </router-link>
            </li>
          </ul>
        </nav>
      </div>
      
      <!-- 右侧功能区 -->
      <div class="actions-area">
        <!-- 搜索框 - 桌面端 -->
        <div v-if="!isMobile && showSearch" class="search-container">
          <div class="search-wrapper">
            <svg class="search-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
            <input
              type="search"
              v-model="searchQuery"
              placeholder="Search..."
              class="search-input"
              @input="handleSearch"
              @focus="isSearchFocused = true"
              @blur="isSearchFocused = false"
            />
            <button
              v-if="searchQuery"
              class="clear-search"
              @click="clearSearch"
              aria-label="Clear search"
            >
              <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
          
          <!-- 搜索建议 -->
          <div
            v-if="searchSuggestions.length > 0 && isSearchFocused"
            class="search-suggestions"
          >
            <div
              v-for="suggestion in searchSuggestions"
              :key="suggestion.id"
              class="suggestion-item"
              @click="selectSuggestion(suggestion)"
            >
              <span class="suggestion-text">{{ suggestion.text }}</span>
              <span class="suggestion-type">{{ suggestion.type }}</span>
            </div>
          </div>
        </div>
        
        <!-- 移动端搜索按钮 -->
        <button
          v-if="isMobile && showSearch"
          class="search-button"
          @click="$emit('open-search')"
          aria-label="Search"
        >
          <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
        </button>
        
        <!-- 通知区域 -->
        <div class="notifications-area">
          <button
            class="notification-button"
            @click="toggleNotifications"
            :class="{ 'has-notifications': unreadCount > 0 }"
            aria-label="Notifications"
            aria-haspopup="true"
            :aria-expanded="showNotifications"
          >
            <svg class="notification-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
            </svg>
            <span v-if="unreadCount > 0" class="notification-badge">
              {{ unreadCount > 9 ? '9+' : unreadCount }}
            </span>
          </button>
          
          <!-- 通知下拉 -->
          <div
            v-if="showNotifications"
            class="notifications-dropdown"
            ref="notificationsDropdown"
          >
            <div class="notifications-header">
              <h3 class="notifications-title">Notifications</h3>
              <button
                v-if="unreadCount > 0"
                class="mark-all-read"
                @click="markAllAsRead"
              >
                Mark all as read
              </button>
            </div>
            
            <div class="notifications-list">
              <div
                v-for="notification in notifications"
                :key="notification.id"
                class="notification-item"
                :class="{ unread: !notification.read }"
                @click="handleNotificationClick(notification)"
              >
                <div class="notification-icon-wrapper">
                  <component
                    :is="getNotificationIcon(notification.type)"
                    class="h-5 w-5"
                  />
                </div>
                <div class="notification-content">
                  <p class="notification-message">{{ notification.message }}</p>
                  <span class="notification-time">{{ notification.time }}</span>
                </div>
                <button
                  v-if="!notification.read"
                  class="mark-as-read"
                  @click.stop="markAsRead(notification.id)"
                  aria-label="Mark as read"
                >
                  <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                  </svg>
                </button>
              </div>
            </div>
            
            <div class="notifications-footer">
              <router-link to="/notifications" class="view-all">
                View all notifications
              </router-link>
            </div>
          </div>
        </div>
        
        <!-- 主题切换 -->
        <button
          class="theme-toggle"
          @click="toggleTheme"
          aria-label="Toggle theme"
        >
          <svg v-if="theme === 'light'" class="theme-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z" />
          </svg>
          <svg v-else class="theme-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z" />
          </svg>
        </button>
        
        <!-- 用户菜单 -->
        <div class="user-menu-area">
          <button
            class="user-button"
            @click="toggleUserMenu"
            aria-label="User menu"
            aria-haspopup="true"
            :aria-expanded="showUserMenu"
          >
            <div class="user-avatar">
              <img
                v-if="user.avatar"
                :src="user.avatar"
                :alt="user.displayName"
                class="avatar-image"
              />
              <div v-else class="avatar-placeholder">
                {{ getUserInitials(user.displayName) }}
              </div>
            </div>
            <div class="user-info">
              <span class="user-name">{{ user.displayName }}</span>
              <span v-if="user.role" class="user-role">{{ user.role }}</span>
            </div>
            <svg class="user-chevron" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
            </svg>
          </button>
          
          <!-- 用户下拉菜单 -->
          <div
            v-if="showUserMenu"
            class="user-menu-dropdown"
            ref="userMenuDropdown"
          >
            <div class="user-menu-header">
              <div class="user-menu-avatar">
                <div class="avatar-large">
                  <img
                    v-if="user.avatar"
                    :src="user.avatar"
                    :alt="user.displayName"
                    class="avatar-image-large"
                  />
                  <div v-else class="avatar-placeholder-large">
                    {{ getUserInitials(user.displayName) }}
                  </div>
                </div>
                <div class="user-menu-info">
                  <h4 class="user-menu-name">{{ user.displayName }}</h4>
                  <p class="user-menu-email">{{ user.email }}</p>
                  <div class="user-menu-role-container">
                    <span class="user-menu-role">{{ user.role || 'User' }}</span>
                    <el-tag v-if="user.isAdmin" size="mini" type="danger">Admin</el-tag>
                  </div>
                </div>
              </div>
            </div>
            
            <div class="user-menu-list">
              <router-link
                to="/profile"
                class="user-menu-item"
                @click="showUserMenu = false"
              >
                <svg class="menu-item-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                </svg>
                <span>Profile</span>
              </router-link>
              
              <router-link
                to="/settings"
                class="user-menu-item"
                @click="showUserMenu = false"
              >
                <svg class="menu-item-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426 -1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543 .826-3.31 2.37-2.37 .996 .608 2.296 .07 2.572-1.065z" />
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                </svg>
                <span>Settings</span>
              </router-link>
              
              <router-link
                v-if="user.isAdmin"
                to="/admin"
                class="user-menu-item"
                @click="showUserMenu = false"
              >
                <svg class="menu-item-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19.428 15.428a2 2 0 00-1.022-.547l-2.387-.477a6 6 0 00-3.86.517l-.318.158a6 6 0 01-3.86.517L6.05 15.21a2 2 0 00-1.806.547M8 4h8l-1 1v5.172a2 2 0 00.586 1.414l5 5c1.26 1.26.367 3.414-1.415 3.414H4.828c-1.782 0-2.674-2.154-1.414-3.414l5-5A2 2 0 009 10.172V5L8 4z" />
                </svg>
                <span>Admin Center</span>
              </router-link>
              
              <div class="menu-divider" />
              
              <button
                class="user-menu-item logout-item"
                @click="handleLogout"
              >
                <svg class="menu-item-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
                </svg>
                <span>Logout</span>
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </nav>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, type VNode } from 'vue'
import { useWindowSize, useEventListener } from '@vueuse/core'
import { useAuthStore } from '@/stores/auth'
import { useAuth } from '@/composables/useAuth'
import {
  CheckCircleIcon,
  InformationCircleIcon,
  ExclamationCircleIcon,
  XCircleIcon,
} from '@/components/icons'

interface MenuItem {
  path: string
  label: string
  icon?: VNode | string
  badge?: string | number
}

interface Props {
  showMobileMenuToggle?: boolean
  menuItems?: MenuItem[]
  showSearch?: boolean
}

// Props - use defineProps without assignment to avoid unused variable error
withDefaults(defineProps<Props>(), {
  showMobileMenuToggle: true,
  menuItems: () => [
    { path: '/dashboard', label: 'Dashboard' },
    { path: '/flows', label: 'Flows' },
    { path: '/resources', label: 'Resources' },
    { path: '/tools', label: 'Tools' },
  ],
  showSearch: true,
})

// Emits
const emit = defineEmits<{
  'toggle-sidebar': []
  'open-search': []
  'logout': []
}>()

// Store
const authStore = useAuthStore()
const { handleLogout: authLogout } = useAuth()

// Reactive state
const searchQuery = ref('')
const searchSuggestions = ref<Array<{id: string, text: string, type: string}>>([])
const isSearchFocused = ref(false)
const showNotifications = ref(false)
const showUserMenu = ref(false)
const activeNavIndex = ref(-1)
const theme = ref<'light' | 'dark'>('light')

// Window size
const { width } = useWindowSize()
const isMobile = computed(() => width.value < 768)

// Notifications (mock data)
const notifications = ref([
  { id: 1, type: 'success', message: 'New user registered', time: '2 min ago', read: false },
  { id: 2, type: 'info', message: 'System update scheduled', time: '1 hour ago', read: true },
  { id: 3, type: 'warning', message: 'Low disk space warning', time: '3 hours ago', read: false },
  { id: 4, type: 'danger', message: 'Failed login attempt detected', time: '1 day ago', read: true },
])

// Computed
const user = computed(() => ({
  displayName: authStore.displayName || 'User',
  avatar: null, // 这里可以以后集成真实头像
  email: authStore.profile?.email || 'user@example.com',
  role: authStore.isAdmin ? 'Administrator' : 'User',
  isAdmin: authStore.isAdmin,
}))

const unreadCount = computed(() =>
  notifications.value.filter(n => !n.read).length
)

// Methods
const toggleTheme = () => {
  theme.value = theme.value === 'light' ? 'dark' : 'light'
  document.documentElement.setAttribute('data-theme', theme.value)
}

const toggleNotifications = () => {
  showNotifications.value = !showNotifications.value
  showUserMenu.value = false
}

const toggleUserMenu = () => {
  showUserMenu.value = !showUserMenu.value
  showNotifications.value = false
}

const handleSearch = () => {
  // 这里可以集成真实搜索逻辑
  console.log('Search:', searchQuery.value)
}

const clearSearch = () => {
  searchQuery.value = ''
  searchSuggestions.value = []
}

const getNotificationIcon = (type: string) => {
  const icons: Record<string, any> = {
    success: CheckCircleIcon,
    info: InformationCircleIcon,
    warning: ExclamationCircleIcon,
    danger: XCircleIcon,
  }
  return icons[type] || InformationCircleIcon
}

const markAsRead = (id: number) => {
  const notification = notifications.value.find(n => n.id === id)
  if (notification) {
    notification.read = true
  }
}

const markAllAsRead = () => {
  notifications.value.forEach(n => (n.read = true))
}

const handleNotificationClick = (notification: any) => {
  markAsRead(notification.id)
  // 这里可以处理通知点击逻辑
  showNotifications.value = false
}

const selectSuggestion = (suggestion: any) => {
  searchQuery.value = suggestion.text
  isSearchFocused.value = false
  // 执行搜索
}

const handleLogout = () => {
  showUserMenu.value = false
  authLogout()
  emit('logout')
}

const getUserInitials = (name: string) => {
  return name
    .split(' ')
    .map(part => part[0])
    .join('')
    .toUpperCase()
    .substring(0, 2)
}

// Event handlers
const handleClickOutside = (event: MouseEvent) => {
  const notificationsDropdown = document.querySelector('.notifications-dropdown')
  const userMenuDropdown = document.querySelector('.user-menu-dropdown')
  
  if (showNotifications.value && notificationsDropdown && 
      !notificationsDropdown.contains(event.target as Node)) {
    showNotifications.value = false
  }
  
  if (showUserMenu.value && userMenuDropdown && 
      !userMenuDropdown.contains(event.target as Node)) {
    showUserMenu.value = false
  }
}

// Lifecycle
onMounted(() => {
  useEventListener(document, 'click', handleClickOutside)
  // 检查系统主题偏好
  if (window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches) {
    theme.value = 'dark'
    document.documentElement.setAttribute('data-theme', 'dark')
  }
})

onUnmounted(() => {
  // Clean up event listeners
})
</script>

<style scoped>
.topbar {
  @apply bg-white border-b border-gray-200 shadow-sm sticky top-0 z-40 w-full;
}

.topbar-container {
  @apply max-w-full mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between;
}

/* Logo 区域样式 */
.logo-area {
  @apply flex items-center gap-4;
}

.mobile-menu-toggle {
  @apply lg:hidden p-2 rounded-md text-gray-500 hover:text-gray-700 hover:bg-gray-100;
}

.logo-wrapper {
  @apply flex items-center;
}

.logo-link {
  @apply flex items-center gap-3 no-underline hover:no-underline;
}

.logo-icon {
  @apply transition-transform duration-300 hover:scale-110;
}

.logo-text {
  @apply hidden md:flex flex-col leading-tight;
}

.logo-text-primary {
  @apply text-gray-900 font-bold text-xl;
}

.logo-text-secondary {
  @apply text-primary-600 font-semibold text-lg;
}

/* 主导航样式 */
.main-nav {
  @apply hidden md:block ml-6;
}

.nav-list {
  @apply flex space-x-1;
}

.nav-item {
  @apply relative;
}

.nav-link {
  @apply flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-medium no-underline transition-colors duration-200;
}

.nav-link:hover {
  @apply bg-gray-100 text-gray-900;
}

.nav-link.active {
  @apply bg-primary-50 text-primary-700;
}

.nav-icon {
  @apply h-4 w-4;
}

.nav-badge {
  @apply ml-1 px-1.5 py-0.5 text-xs font-medium rounded-full bg-primary-100 text-primary-800;
}

/* 右侧功能区样式 */
.actions-area {
  @apply flex items-center gap-2 sm:gap-4;
}

/* 搜索框样式 */
.search-container {
  @apply relative w-64 hidden md:block;
}

.search-wrapper {
  @apply relative;
}

.search-icon {
  @apply absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400;
}

.search-input {
  @apply w-full pl-10 pr-10 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent;
}

.clear-search {
  @apply absolute right-3 top-1/2 transform -translate-y-1/2 p-1 text-gray-400 hover:text-gray-600;
}

.search-suggestions {
  @apply absolute top-full mt-1 w-full bg-white border border-gray-200 rounded-lg shadow-lg z-50 overflow-hidden;
}

.suggestion-item {
  @apply px-4 py-3 text-sm hover:bg-gray-50 cursor-pointer flex items-center justify-between;
}

.suggestion-text {
  @apply text-gray-900;
}

.suggestion-type {
  @apply text-xs text-gray-500 bg-gray-100 px-2 py-0.5 rounded;
}

.search-button {
  @apply md:hidden p-2 rounded-full text-gray-500 hover:text-gray-700 hover:bg-gray-100;
}

/* 通知区域样式 */
.notifications-area {
  @apply relative;
}

.notification-button {
  @apply relative p-2 rounded-full text-gray-500 hover:text-gray-700 hover:bg-gray-100;
}

.notification-button.has-notifications {
  @apply text-primary-600;
}

.notification-icon {
  @apply h-5 w-5;
}

.notification-badge {
  @apply absolute -top-1 -right-1 bg-error-500 text-white text-xs rounded-full w-5 h-5 flex items-center justify-center;
}

.notifications-dropdown {
  @apply absolute right-0 mt-2 w-80 bg-white border border-gray-200 rounded-lg shadow-lg z-50 overflow-hidden;
  top: 100%;
}

.notifications-header {
  @apply px-4 py-3 border-b border-gray-200 flex items-center justify-between;
}

.notifications-title {
  @apply text-sm font-semibold text-gray-900;
}

.mark-all-read {
  @apply text-xs text-primary-600 hover:text-primary-800;
}

.notifications-list {
  @apply max-h-96 overflow-y-auto;
}

.notification-item {
  @apply px-4 py-3 hover:bg-gray-50 cursor-pointer flex items-start gap-3;
}

.notification-item.unread {
  @apply bg-primary-50;
}

.notification-icon-wrapper {
  @apply mt-0.5;
}

.notification-content {
  @apply flex-1 min-w-0;
}

.notification-message {
  @apply text-sm text-gray-900;
}

.notification-time {
  @apply text-xs text-gray-500 mt-1 block;
}

.mark-as-read {
  @apply p-1 text-gray-400 hover:text-primary-600;
}

.notifications-footer {
  @apply px-4 py-3 border-t border-gray-200 text-center;
}

.view-all {
  @apply text-sm text-primary-600 hover:text-primary-800 no-underline;
}

/* 主题切换样式 */
.theme-toggle {
  @apply p-2 rounded-full text-gray-500 hover:text-gray-700 hover:bg-gray-100;
}

.theme-icon {
  @apply h-5 w-5;
}

/* 用户菜单样式 */
.user-menu-area {
  @apply relative;
}

.user-button {
  @apply flex items-center gap-3 p-2 rounded-full hover:bg-gray-100;
}

.user-avatar {
  @apply h-8 w-8 rounded-full bg-primary-100 text-primary-600 flex items-center justify-center font-medium text-sm;
}

.avatar-image {
  @apply h-full w-full rounded-full object-cover;
}

.avatar-placeholder {
  @apply h-full w-full rounded-full flex items-center justify-center bg-gradient-to-br from-primary-500 to-secondary-500 text-white font-semibold;
}

.user-info {
  @apply hidden md:block text-left;
}

.user-name {
  @apply text-sm font-medium text-gray-900 block;
}

.user-role {
  @apply text-xs text-gray-500 block;
}

.user-chevron {
  @apply hidden md:block h-5 w-5 text-gray-400;
}

.user-menu-dropdown {
  @apply absolute right-0 mt-2 w-64 bg-white border border-gray-200 rounded-lg shadow-lg z-50 overflow-hidden;
  top: 100%;
}

.user-menu-header {
  @apply px-4 py-3 border-b border-gray-200;
}

.user-menu-avatar {
  @apply flex items-center gap-3;
}

.avatar-large {
  @apply h-12 w-12 rounded-full bg-primary-100 text-primary-600 flex items-center justify-center font-medium text-base;
}

.avatar-image-large {
  @apply h-full w-full rounded-full object-cover;
}

.avatar-placeholder-large {
  @apply h-full w-full rounded-full flex items-center justify-center bg-gradient-to-br from-primary-500 to-secondary-500 text-white font-semibold;
}

.user-menu-info {
  @apply flex-1 min-w-0;
}

.user-menu-name {
  @apply text-sm font-semibold text-gray-900 truncate;
}

.user-menu-email {
  @apply text-xs text-gray-500 truncate mt-0.5;
}

.user-menu-role-container {
  @apply flex items-center gap-2 mt-1;
}

.user-menu-role {
  @apply text-xs text-gray-600;
}

.user-menu-list {
  @apply py-2;
}

.user-menu-item {
  @apply flex items-center gap-3 px-4 py-2.5 text-sm text-gray-700 hover:bg-gray-50 hover:text-gray-900 no-underline;
  @apply w-full text-left;
}

.menu-item-icon {
  @apply h-5 w-5 text-gray-400;
}

.menu-divider {
  @apply border-t border-gray-200 my-2;
}

.logout-item {
  @apply text-error-600 hover:text-error-800;
}

/* 响应式调整 */
@media (max-width: 767px) {
  .topbar-container {
    @apply px-3;
  }
  
  .actions-area {
    @apply gap-1;
  }
  
  .user-info {
    @apply hidden;
  }
  
  .user-chevron {
    @apply hidden;
  }
}
</style>