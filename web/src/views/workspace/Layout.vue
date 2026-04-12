<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useWorkspaceRoute } from '@/composables/useWorkspaceRoute'
import { useI18n } from 'vue-i18n'

const route = useRoute()
const { t } = useI18n()
const { settingsPath, settingsMembersPath } = useWorkspaceRoute()

const menuItems = computed(() => [
  { path: settingsPath(), label: t('workspace.settings'), icon: 'settings' },
  { path: settingsMembersPath(), label: t('workspace.members'), icon: 'users' },
])

function isActive(item: { path: string }) {
  return route.path === item.path
}
</script>

<template>
  <div class="workspace-settings-layout">
    <div class="settings-sidebar">
      <nav class="settings-nav">
        <router-link
          v-for="item in menuItems"
          :key="item.path"
          :to="item.path"
          class="settings-nav-item"
          :class="{ 'settings-nav-item-active': isActive(item) }"
        >
          <!-- Settings icon -->
          <svg v-if="item.icon === 'settings'" class="nav-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" width="18" height="18">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9.594 3.94c.09-.542.56-.94 1.11-.94h2.593c.55 0 1.02.398 1.11.94l.213 1.281c.063.374.313.686.645.87.074.04.147.083.22.127.325.196.72.257 1.075.124l1.217-.456a1.125 1.125 0 011.37.49l1.296 2.247a1.125 1.125 0 01-.26 1.431l-1.003.827c-.293.241-.438.613-.43.992a7.723 7.723 0 010 .255c-.008.378.137.75.43.991l1.004.827c.424.35.534.955.26 1.43l-1.298 2.247a1.125 1.125 0 01-1.369.491l-1.217-.456c-.355-.133-.75-.072-1.076.124a6.47 6.47 0 01-.22.128c-.331.183-.581.495-.644.869l-.213 1.281c-.09.543-.56.94-1.11.94h-2.594c-.55 0-1.019-.398-1.11-.94l-.212-1.281c-.062-.374-.312-.686-.644-.87a6.52 6.52 0 01-.22-.127c-.325-.196-.72-.257-1.076-.124l-1.217.456a1.125 1.125 0 01-1.369-.49l-1.297-2.247a1.125 1.125 0 01.26-1.431l1.004-.827c.292-.24.437-.613.43-.991a6.932 6.932 0 010-.255c.007-.38-.138-.751-.43-.992l-1.004-.827a1.125 1.125 0 01-.26-1.43l1.297-2.247a1.125 1.125 0 011.37-.491l1.216.456c.356.133.751.072 1.076-.124.072-.044.146-.086.22-.128.332-.183.582-.495.644-.869l.214-1.28z" />
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
          </svg>
          <!-- Users icon -->
          <svg v-else-if="item.icon === 'users'" class="nav-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" width="18" height="18">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15 19.128a9.38 9.38 0 002.625.372 9.337 9.337 0 004.121-.952 4.125 4.125 0 00-7.533-2.493M15 19.128v-.003c0-1.113-.285-2.16-.786-3.07M15 19.128v.106A12.318 12.318 0 018.624 21c-2.331 0-4.512-.645-6.374-1.766l-.001-.109a6.375 6.375 0 0111.964-3.07M12 6.375a3.375 3.375 0 11-6.75 0 3.375 3.375 0 016.75 0zm8.25 2.25a2.625 2.625 0 11-5.25 0 2.625 2.625 0 015.25 0z" />
          </svg>
          <span>{{ item.label }}</span>
        </router-link>
      </nav>
    </div>
    <div class="settings-content">
      <RouterView />
    </div>
  </div>
</template>

<style scoped>
.workspace-settings-layout {
  display: flex;
  gap: 24px;
  min-height: 500px;
}

.settings-sidebar {
  width: 200px;
  flex-shrink: 0;
}

.settings-nav {
  display: flex;
  flex-direction: column;
  gap: 2px;
  position: sticky;
  top: 24px;
}

.settings-nav-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border-radius: 6px;
  font-size: 14px;
  color: #656d76;
  text-decoration: none;
  transition: all 0.15s;
}

.settings-nav-item:hover {
  background: #f3f4f6;
  color: #1f2328;
  text-decoration: none;
}

.settings-nav-item-active {
  background: #eff6ff;
  color: #0969da;
  font-weight: 600;
}

.settings-nav-item-active:hover {
  background: #eff6ff;
  color: #0969da;
}

.nav-icon {
  flex-shrink: 0;
}

.settings-content {
  flex: 1;
  min-width: 0;
}

@media (max-width: 767px) {
  .workspace-settings-layout {
    flex-direction: column;
    gap: 16px;
  }

  .settings-sidebar {
    width: 100%;
  }

  .settings-nav {
    flex-direction: row;
    position: static;
    overflow-x: auto;
    scrollbar-width: none;
  }

  .settings-nav::-webkit-scrollbar {
    display: none;
  }

  .settings-nav-item {
    white-space: nowrap;
  }
}
</style>
