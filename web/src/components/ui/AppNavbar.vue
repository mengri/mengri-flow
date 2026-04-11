<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useWindowSize } from '@vueuse/core'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import { useWorkspaceStore } from '@/stores/workspace'
import { useWorkspaceRoute } from '@/composables/useWorkspaceRoute'
import { useAuth } from '@/composables/useAuth'
import AppLogo from '@/components/ui/AppLogo.vue'
import LanguageSwitcher from '@/components/LanguageSwitcher.vue'
import type { Workspace } from '@/types/workspace'
import {
  HomeIcon,
  PuzzleIcon,
  ArrowsRightLeftIcon,
  PlayIcon,
  WrenchScrewdriverIcon,
  QueueListIcon,
} from '@/components/icons'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const { width } = useWindowSize()
const isMobile = computed(() => width.value < 768)

const authStore = useAuthStore()
const workspaceStore = useWorkspaceStore()
const { handleLogout: authLogout } = useAuth()
const {
  dashboardPath,
  flowsPath,
  triggersPath,
  resourcesPath,
  toolsPath,
  runsPath,
} = useWorkspaceRoute()

// --- Workspace panel (triggered by hamburger) ---
const showWsPanel = ref(false)

const workspaces = computed(() => workspaceStore.workspaces)
const currentWorkspace = computed(() => workspaceStore.currentWorkspace)

function switchWorkspace(ws: Workspace) {
  workspaceStore.setCurrentWorkspace(ws.id)
  router.push(`/workspace/${ws.id}`)
  showWsPanel.value = false
}

function goToWorkspaces() {
  showWsPanel.value = false
  router.push('/workspaces')
}

// --- Breadcrumb ---
const breadcrumbs = computed(() => {
  const ws = currentWorkspace.value
  const crumbs: Array<{ path: string; label: string; isLast?: boolean }> = []
  const pathArray = route.path.split('/').filter(Boolean)
  const isWorkspaceRoute = pathArray[0] === 'workspace'

  if (isWorkspaceRoute && ws) {
    const workspaceId = route.params.workspaceId as string
    const wsBasePath = workspaceId ? `/workspace/${workspaceId}` : ''

    // First crumb: workspace name
    crumbs.push({ path: wsBasePath || dashboardPath(), label: ws.name })

    // Sub-pages under a module
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
  } else if (route.path.startsWith('/admin')) {
    crumbs.push({ path: '/admin/accounts', label: t('nav.account'), isLast: route.path === '/admin/accounts' })
  } else if (route.path.startsWith('/workspaces')) {
    crumbs.push({ path: '/workspaces', label: t('nav.manageWorkspaces'), isLast: true })
  } else if (route.path.startsWith('/account')) {
    crumbs.push({ path: '/account', label: t('common.profile'), isLast: true })
  }

  return crumbs
})

// --- Settings menu ---
const showSettingsMenu = ref(false)

const settingsMenuItems = computed(() => {
  const items: Array<{ path: string; label: string; adminOnly?: boolean }> = [
    { path: '/workspaces', label: t('nav.manageWorkspaces') },
  ]
  if (user.value.isAdmin) {
    items.push({ path: '/admin/accounts', label: t('nav.account'), adminOnly: true })
  }
  return items
})

function closeSettingsMenu() {
  showSettingsMenu.value = false
}

// --- User menu ---
const showUserMenu = ref(false)

const user = computed(() => ({
  displayName: authStore.displayName || 'User',
  email: authStore.profile?.email || '',
  isAdmin: authStore.isAdmin,
}))

function handleLogout() {
  showUserMenu.value = false
  authLogout()
}

// --- Tab navigation ---
const isWorkspaceRoute = computed(() => route.path.startsWith('/workspace/'))

const workspaceTabs = computed(() => [
  { path: dashboardPath(), label: t('nav.dashboard'), icon: HomeIcon },
  { path: resourcesPath(), label: t('nav.resources'), icon: PuzzleIcon },
  { path: flowsPath(), label: t('nav.flows'), icon: ArrowsRightLeftIcon },
  { path: triggersPath(), label: t('nav.triggers'), icon: PlayIcon },
  { path: toolsPath(), label: t('nav.tools'), icon: WrenchScrewdriverIcon },
  { path: runsPath(), label: t('nav.runList'), icon: QueueListIcon },
])

const isTabActive = (tab: { path: string }) => {
  if (route.path === tab.path) return true
  return route.path.startsWith(tab.path + '/')
}

// --- Non-workspace contextual tabs ---
const nonWsTabs = computed(() => {
  if (route.path.startsWith('/account')) {
    return [{ path: '/account', label: t('common.profile') }]
  }
  if (route.path.startsWith('/workspaces')) {
    return [{ path: '/workspaces', label: t('nav.manageWorkspaces') }]
  }
  if (route.path.startsWith('/admin')) {
    return [{ path: '/admin/accounts', label: t('nav.account') }]
  }
  return []
})

// --- Click outside ---
function handleClickOutside(e: MouseEvent) {
  const target = e.target as HTMLElement
  if (showWsPanel.value && !target.closest('.ws-panel-area')) {
    showWsPanel.value = false
  }
  if (showSettingsMenu.value && !target.closest('.settings-menu-area')) {
    showSettingsMenu.value = false
  }
  if (showUserMenu.value && !target.closest('.user-menu-area')) {
    showUserMenu.value = false
  }
}

// --- Keyboard: close panels on Escape ---
function handleKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') {
    showWsPanel.value = false
    showSettingsMenu.value = false
    showUserMenu.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
  document.addEventListener('keydown', handleKeydown)
})
onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
  document.removeEventListener('keydown', handleKeydown)
})
</script>

<template>
  <header class="app-navbar">
    <!-- Row 1: Global bar -->
    <div class="navbar-row navbar-global">
      <div class="navbar-inner">
        <!-- Left: Hamburger + Breadcrumb -->
        <div class="navbar-left">
          <!-- Hamburger button to toggle workspace panel -->
          <div class="ws-panel-area">
            <button class="hamburger-btn" @click.stop="showWsPanel = !showWsPanel" aria-label="Switch workspace">
              <svg fill="none" viewBox="0 0 24 24" stroke="currentColor" width="20" height="20">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
              </svg>
            </button>

            <!-- Workspace switch panel -->
            <transition name="dropdown">
              <div v-if="showWsPanel" class="ws-panel">
                <div class="ws-panel-header">
                  <span class="ws-panel-title">{{ t('nav.workspace') }}</span>
                </div>
                <div class="ws-panel-list">
                  <button
                    v-for="ws in workspaces"
                    :key="ws.id"
                    class="ws-panel-item"
                    :class="{ 'ws-panel-item-active': ws.id === currentWorkspace?.id }"
                    @click="switchWorkspace(ws)"
                  >
                    <span class="ws-panel-item-avatar">
                      {{ ws.name.split(/[\s\-_]+/).map(p => p[0]).join('').toUpperCase().substring(0, 2) }}
                    </span>
                    <div class="ws-panel-item-info">
                      <span class="ws-panel-item-name">{{ ws.name }}</span>
                    </div>
                    <svg v-if="ws.id === currentWorkspace?.id" class="ws-panel-item-check" fill="none" viewBox="0 0 24 24" stroke="currentColor" width="16" height="16">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                    </svg>
                  </button>
                </div>
                <div class="ws-panel-footer">
                  <button class="ws-panel-manage" @click="goToWorkspaces">
                    <svg fill="none" viewBox="0 0 24 24" stroke="currentColor" width="14" height="14">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                    </svg>
                    {{ t('nav.manageWorkspaces') }}
                  </button>
                </div>
              </div>
            </transition>
          </div>

          <!-- Logo -->
          <div class="logo-area">
            <router-link to="/" class="logo-link">
              <AppLogo size="sm" />
            </router-link>
          </div>

          <!-- Breadcrumb -->
          <nav v-if="breadcrumbs.length > 0" class="breadcrumb-nav">
            <template v-for="(crumb, index) in breadcrumbs" :key="crumb.path">
              <svg v-if="index > 0" class="breadcrumb-sep" fill="none" viewBox="0 0 16 16" width="12" height="12">
                <path fill="currentColor" d="m6.22 3.22 4.25 4.25a.75.75 0 010 1.06l-4.25 4.25a.75.75 0 01-1.06-1.06L8.94 8 5.16 4.28a.75.75 0 111.06-1.06z" />
              </svg>
              <router-link
                v-if="!crumb.isLast"
                :to="crumb.path"
                class="breadcrumb-link"
              >
                {{ crumb.label }}
              </router-link>
              <span v-else class="breadcrumb-current">{{ crumb.label }}</span>
            </template>
          </nav>
        </div>

        <!-- Right: Actions -->
        <div class="navbar-right">
          <!-- Settings menu -->
          <div class="settings-menu-area">
            <button class="settings-menu-btn" @click.stop="showSettingsMenu = !showSettingsMenu" aria-label="Settings">
              <svg fill="none" viewBox="0 0 24 24" stroke="currentColor" width="20" height="20">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.594 3.94c.09-.542.56-.94 1.11-.94h2.593c.55 0 1.02.398 1.11.94l.213 1.281c.063.374.313.686.645.87.074.04.147.083.22.127.325.196.72.257 1.075.124l1.217-.456a1.125 1.125 0 011.37.49l1.296 2.247a1.125 1.125 0 01-.26 1.431l-1.003.827c-.293.241-.438.613-.43.992a7.723 7.723 0 010 .255c-.008.378.137.75.43.991l1.004.827c.424.35.534.955.26 1.43l-1.298 2.247a1.125 1.125 0 01-1.369.491l-1.217-.456c-.355-.133-.75-.072-1.076.124a6.47 6.47 0 01-.22.128c-.331.183-.581.495-.644.869l-.213 1.281c-.09.543-.56.94-1.11.94h-2.594c-.55 0-1.019-.398-1.11-.94l-.212-1.281c-.062-.374-.312-.686-.644-.87a6.52 6.52 0 01-.22-.127c-.325-.196-.72-.257-1.076-.124l-1.217.456a1.125 1.125 0 01-1.369-.49l-1.297-2.247a1.125 1.125 0 01.26-1.431l1.004-.827c.292-.24.437-.613.43-.991a6.932 6.932 0 010-.255c.007-.38-.138-.751-.43-.992l-1.004-.827a1.125 1.125 0 01-.26-1.43l1.297-2.247a1.125 1.125 0 011.37-.491l1.216.456c.356.133.751.072 1.076-.124.072-.044.146-.086.22-.128.332-.183.582-.495.644-.869l.214-1.28z" />
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
              </svg>
            </button>

            <transition name="dropdown">
              <div v-if="showSettingsMenu" class="settings-dropdown">
                <div class="settings-dropdown-list">
                  <router-link
                    v-for="item in settingsMenuItems"
                    :key="item.path"
                    :to="item.path"
                    class="settings-dropdown-item"
                    @click="closeSettingsMenu"
                  >
                    {{ item.label }}
                  </router-link>
                </div>
              </div>
            </transition>
          </div>

          <!-- Language switcher (desktop) -->
          <LanguageSwitcher v-if="!isMobile" />

          <!-- User menu -->
          <div class="user-menu-area">
            <button
              class="user-menu-btn"
              @click.stop="showUserMenu = !showUserMenu"
            >
              <div class="user-avatar">
                {{ user.displayName.split(' ').map(p => p[0]).join('').toUpperCase().substring(0, 2) }}
              </div>
            </button>

            <!-- User dropdown -->
            <transition name="dropdown">
              <div v-if="showUserMenu" class="user-dropdown">
                <div class="user-dropdown-header">
                  <div class="user-dropdown-avatar">
                    {{ user.displayName.split(' ').map(p => p[0]).join('').toUpperCase().substring(0, 2) }}
                  </div>
                  <div class="user-dropdown-info">
                    <span class="user-dropdown-name">{{ user.displayName }}</span>
                    <span v-if="user.email" class="user-dropdown-email">{{ user.email }}</span>
                  </div>
                </div>
                <div class="user-dropdown-list">
                  <router-link
                    to="/account"
                    class="user-dropdown-item"
                    @click="showUserMenu = false"
                  >
                    <svg fill="none" viewBox="0 0 24 24" stroke="currentColor" width="16" height="16">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                    </svg>
                    {{ t('common.profile') }}
                  </router-link>
                </div>
                <div class="user-dropdown-footer">
                  <button class="user-dropdown-item logout-item" @click="handleLogout">
                    <svg fill="none" viewBox="0 0 24 24" stroke="currentColor" width="16" height="16">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
                    </svg>
                    {{ t('common.logout') }}
                  </button>
                </div>
              </div>
            </transition>
          </div>
        </div>
      </div>
    </div>

    <!-- Row 2: Contextual tab bar -->
    <div v-if="isWorkspaceRoute || nonWsTabs.length > 0" class="navbar-row navbar-tabs">
      <div class="navbar-inner">
        <nav class="tabs-nav">
          <!-- Workspace tabs -->
          <template v-if="isWorkspaceRoute">
            <router-link
              v-for="tab in workspaceTabs"
              :key="tab.path"
              :to="tab.path"
              class="tab-item"
              :class="{ 'tab-item-active': isTabActive(tab) }"
            >
              <component v-if="tab.icon" :is="tab.icon" class="tab-icon" />
              <span class="tab-label">{{ tab.label }}</span>
            </router-link>
          </template>
          <!-- Non-workspace tabs -->
          <template v-else>
            <router-link
              v-for="tab in nonWsTabs"
              :key="tab.path"
              :to="tab.path"
              class="tab-item"
              :class="{ 'tab-item-active': route.path === tab.path }"
            >
              <span class="tab-label">{{ tab.label }}</span>
            </router-link>
          </template>
        </nav>
      </div>
    </div>
  </header>
</template>

<style scoped>
/* === Global bar === */
.app-navbar {
  position: sticky;
  top: 0;
  z-index: 50;
  width: 100%;
  background: #fff;
  border-bottom: 1px solid #d0d7de;
}

.navbar-row {
  width: 100%;
}

.navbar-inner {
  max-width: 1280px;
  margin: 0 auto;
  padding: 0 16px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 100%;
}

/* Row 1 */
.navbar-global {
  height: 56px;
  background: #24292f;
}

.navbar-global .navbar-inner {
  max-width: 100%;
  padding: 0 16px;
}

.navbar-left {
  display: flex;
  align-items: center;
  gap: 0;
  min-width: 0;
  flex: 1;
}

.navbar-right {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

/* Hamburger button */
.ws-panel-area {
  position: relative;
  flex-shrink: 0;
}

.hamburger-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border-radius: 6px;
  border: none;
  background: transparent;
  color: rgba(255, 255, 255, 0.85);
  cursor: pointer;
  transition: background 0.15s;
}

.hamburger-btn:hover {
  background: rgba(255, 255, 255, 0.1);
}

/* Workspace panel dropdown */
.ws-panel {
  position: absolute;
  top: calc(100% + 4px);
  left: 0;
  width: 300px;
  background: #fff;
  border: 1px solid #d0d7de;
  border-radius: 12px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
  z-index: 100;
  overflow: hidden;
}

.ws-panel-header {
  padding: 8px 12px;
  border-bottom: 1px solid #d8dee4;
}

.ws-panel-title {
  font-size: 12px;
  font-weight: 600;
  color: #656d76;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.ws-panel-list {
  max-height: 320px;
  overflow-y: auto;
  padding: 4px;
}

.ws-panel-item {
  display: flex;
  align-items: center;
  width: 100%;
  padding: 8px;
  border-radius: 6px;
  border: none;
  background: transparent;
  cursor: pointer;
  transition: background 0.1s;
  color: inherit;
  text-align: left;
  font-size: inherit;
}

.ws-panel-item:hover {
  background: #f3f4f6;
}

.ws-panel-item-active {
  background: #eff6ff;
}

.ws-panel-item-avatar {
  width: 32px;
  height: 32px;
  border-radius: 6px;
  background: linear-gradient(135deg, #7c3aed, #2563eb);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: 700;
  color: #fff;
  flex-shrink: 0;
}

.ws-panel-item-info {
  flex: 1;
  margin-left: 8px;
  min-width: 0;
}

.ws-panel-item-name {
  font-size: 14px;
  font-weight: 500;
  color: #1f2328;
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ws-panel-item-check {
  flex-shrink: 0;
  color: #0969da;
}

.ws-panel-footer {
  padding: 4px;
  border-top: 1px solid #d8dee4;
}

.ws-panel-manage {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
  padding: 8px;
  border-radius: 6px;
  border: none;
  background: transparent;
  cursor: pointer;
  font-size: 13px;
  color: #656d76;
  transition: all 0.1s;
}

.ws-panel-manage:hover {
  background: #f3f4f6;
  color: #1f2328;
}

/* Logo */
.logo-area {
  flex-shrink: 0;
  padding: 4px 12px 4px 8px;
}

.logo-link {
  display: flex;
  align-items: center;
  text-decoration: none;
}

/* Breadcrumb */
.breadcrumb-nav {
  display: flex;
  align-items: center;
  gap: 2px;
  margin-left: 4px;
  overflow: hidden;
}

.breadcrumb-sep {
  color: rgba(255, 255, 255, 0.3);
  flex-shrink: 0;
}

.breadcrumb-link {
  font-size: 14px;
  color: rgba(255, 255, 255, 0.8);
  text-decoration: none;
  white-space: nowrap;
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  transition: color 0.15s;
}

.breadcrumb-link:hover {
  color: #fff;
  text-decoration: none;
}

.breadcrumb-current {
  font-size: 14px;
  font-weight: 600;
  color: #fff;
  white-space: nowrap;
  max-width: 240px;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* Settings menu */
.settings-menu-area {
  position: relative;
}

.settings-menu-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border-radius: 6px;
  border: none;
  background: transparent;
  color: rgba(255, 255, 255, 0.85);
  cursor: pointer;
  transition: background 0.15s;
}

.settings-menu-btn:hover {
  background: rgba(255, 255, 255, 0.1);
}

.settings-dropdown {
  position: absolute;
  top: calc(100% + 4px);
  right: 0;
  width: 200px;
  background: #fff;
  border: 1px solid #d0d7de;
  border-radius: 12px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
  z-index: 100;
  overflow: hidden;
}

.settings-dropdown-list {
  padding: 4px;
}

.settings-dropdown-item {
  display: block;
  padding: 8px 12px;
  border-radius: 6px;
  font-size: 14px;
  color: #1f2328;
  text-decoration: none;
  transition: background 0.1s;
}

.settings-dropdown-item:hover {
  background: #f3f4f6;
  text-decoration: none;
}

/* User menu */
.user-menu-area {
  position: relative;
}

.user-menu-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 8px;
  border-radius: 6px;
  border: 1px solid rgba(255, 255, 255, 0.2);
  background: transparent;
  color: #fff;
  font-size: 14px;
  cursor: pointer;
  transition: background 0.15s;
}

.user-menu-btn:hover {
  background: rgba(255, 255, 255, 0.1);
}

.user-avatar {
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: linear-gradient(135deg, #f97316, #ec4899);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 10px;
  font-weight: 700;
  color: #fff;
  flex-shrink: 0;
}

.user-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 120px;
}

.user-chevron {
  opacity: 0.7;
}

/* User dropdown */
.user-dropdown {
  position: absolute;
  top: calc(100% + 4px);
  right: 0;
  width: 240px;
  background: #fff;
  border: 1px solid #d0d7de;
  border-radius: 12px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
  z-index: 100;
  overflow: hidden;
}

.user-dropdown-header {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px;
  border-bottom: 1px solid #d8dee4;
}

.user-dropdown-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: linear-gradient(135deg, #f97316, #ec4899);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  font-weight: 700;
  color: #fff;
  flex-shrink: 0;
}

.user-dropdown-info {
  min-width: 0;
}

.user-dropdown-name {
  font-size: 14px;
  font-weight: 600;
  color: #1f2328;
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.user-dropdown-email {
  font-size: 12px;
  color: #656d76;
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.user-dropdown-list {
  padding: 4px;
}

.user-dropdown-item {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
  padding: 8px;
  border-radius: 6px;
  border: none;
  background: transparent;
  cursor: pointer;
  font-size: 14px;
  color: #1f2328;
  text-decoration: none;
  text-align: left;
  transition: background 0.1s;
}

.user-dropdown-item:hover {
  background: #f3f4f6;
  text-decoration: none;
}

.logout-item {
  color: #d1242f;
}

.logout-item:hover {
  background: #ffebe9;
  color: #d1242f;
}

.user-dropdown-footer {
  padding: 4px;
  border-top: 1px solid #d8dee4;
}

/* === Tab bar (Row 2) === */
.navbar-tabs {
  height: 48px;
  background: #fff;
  border-bottom: 1px solid #d0d7de;
}

.tabs-nav {
  display: flex;
  align-items: center;
  gap: 0;
  height: 100%;
  overflow-x: auto;
  scrollbar-width: none;
}

.tabs-nav::-webkit-scrollbar {
  display: none;
}

.tab-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  font-size: 14px;
  color: #656d76;
  text-decoration: none;
  border-bottom: 2px solid transparent;
  transition: color 0.15s, border-color 0.15s;
  white-space: nowrap;
  flex-shrink: 0;
}

.tab-item:hover {
  color: #1f2328;
  text-decoration: none;
}

.tab-item-active {
  color: #1f2328;
  font-weight: 600;
  border-bottom-color: #fd8c73;
}

.tab-icon {
  width: 16px;
  height: 16px;
}

.tab-label {
  line-height: 1;
}

/* === Dropdown animation === */
.dropdown-enter-active,
.dropdown-leave-active {
  transition: opacity 0.15s ease, transform 0.15s ease;
  transform-origin: top;
}

.dropdown-enter-from,
.dropdown-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}

/* === Responsive === */
@media (max-width: 767px) {
  .ws-panel {
    width: 260px;
    left: -8px;
  }

  .breadcrumb-link {
    max-width: 120px;
  }

  .breadcrumb-current {
    max-width: 140px;
  }

  .tab-item {
    padding: 8px 12px;
    font-size: 13px;
  }
}

/* Override LanguageSwitcher styles for dark background */
.navbar-global :deep(.language-btn) {
  background: transparent;
  border: 1px solid rgba(255, 255, 255, 0.2);
  color: #fff;
}

.navbar-global :deep(.language-btn:hover) {
  background: rgba(255, 255, 255, 0.1);
  color: #fff;
}

.navbar-global :deep(.language-icon),
.navbar-global :deep(.language-text) {
  color: rgba(255, 255, 255, 0.9);
}

.navbar-global :deep(.arrow-icon) {
  color: rgba(255, 255, 255, 0.6);
}
</style>
