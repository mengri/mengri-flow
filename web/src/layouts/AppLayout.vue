<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useAuth } from '@/composables/useAuth'

const route = useRoute()
const authStore = useAuthStore()
const { handleLogout } = useAuth()

const menuItems = computed(() => {
  const items = [
    { path: '/account', label: 'Account Center', icon: 'User' },
  ]
  if (authStore.isAdmin) {
    items.push({ path: '/admin/accounts', label: 'Account Management', icon: 'Setting' })
  }
  return items
})

const activeMenu = computed(() => route.path)
</script>

<template>
  <div class="min-h-screen flex flex-col">
    <!-- Top navbar -->
    <header class="bg-white border-b border-gray-200 px-6 h-14 flex items-center justify-between shrink-0">
      <div class="flex items-center gap-6">
        <router-link to="/" class="text-lg font-bold text-gray-800 no-underline">
          Mengri Flow
        </router-link>
        <nav class="flex gap-1">
          <router-link
            v-for="item in menuItems"
            :key="item.path"
            :to="item.path"
            class="px-3 py-2 rounded text-sm no-underline transition-colors"
            :class="[
              activeMenu === item.path
                ? 'bg-blue-50 text-blue-600 font-medium'
                : 'text-gray-600 hover:bg-gray-100'
            ]"
          >
            {{ item.label }}
          </router-link>
        </nav>
      </div>

      <div class="flex items-center gap-4">
        <span class="text-sm text-gray-600">
          {{ authStore.displayName }}
        </span>
        <el-tag v-if="authStore.isAdmin" type="danger" size="small">admin</el-tag>
        <el-button size="small" type="info" text @click="handleLogout">
          Logout
        </el-button>
      </div>
    </header>

    <!-- Main content -->
    <main class="flex-1 bg-gray-50">
      <slot />
    </main>
  </div>
</template>
