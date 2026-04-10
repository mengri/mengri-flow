<script setup lang="ts">
import { computed, watch, onMounted } from 'vue'
import { RouterView } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import { useWorkspaceStore } from '@/stores/workspace'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import en from 'element-plus/es/locale/lang/en'

const authStore = useAuthStore()
const workspaceStore = useWorkspaceStore()
const { locale } = useI18n()

// Element Plus 语言映射
const elementLocales: Record<string, any> = {
  zh: zhCn,
  en: en,
}

// Element Plus 配置
const elConfig = computed(() => ({
  size: 'default',
  zIndex: 3000,
  locale: elementLocales[locale.value] || en,
}))

// 监听语言变化，同步到 localStorage
watch(locale, (newLocale) => {
  localStorage.setItem('locale', newLocale)
}, { immediate: true })

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
