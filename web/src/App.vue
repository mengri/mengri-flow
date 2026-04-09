<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { RouterView } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useWorkspaceStore } from '@/stores/workspace'

const authStore = useAuthStore()
const workspaceStore = useWorkspaceStore()

// Element Plus 配置
const elConfig = computed(() => ({
  size: 'default',
  zIndex: 3000,
}))

// 应用初始化
onMounted(async () => {
  // 如果用户已登录，加载工作空间列表
  if (authStore.isAuthenticated) {
    try {
      await workspaceStore.loadWorkspaces()
    } catch (error) {
      console.error('Failed to load workspaces:', error)
    }
  }
})
</script>

<template>
  <el-config-provider :="elConfig">
    <RouterView />
  </el-config-provider>
</template>
