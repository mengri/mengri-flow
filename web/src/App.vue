<script setup lang="ts">
import { computed, watch, onMounted, ref } from 'vue'
import { RouterView } from 'vue-router'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import { useWorkspaceStore } from '@/stores/workspace'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import en from 'element-plus/es/locale/lang/en'

const router = useRouter()
const authStore = useAuthStore()
const workspaceStore = useWorkspaceStore()
const { locale } = useI18n()

// 应用初始化完成标志（防止 workspace 未加载时页面闪烁）
const initialized = ref(false)

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
  if (authStore.isAuthenticated) {
    // 有 token，先验证有效性（同时拉取 profile）
    const profile = await authStore.fetchProfile()
    if (!profile) {
      // token 无效或已过期，fetchProfile 会收到 401
      // client.ts 的拦截器已处理跳转，这里只需标记初始化完成
      initialized.value = true
      return
    }

    // profile 有效，加载工作空间列表
    try {
      const status = await workspaceStore.loadWorkspaces()
      if (status === 'none') {
        // 无已选中的 workspace，跳转到选择页
        router.replace('/select-workspace')
      }
    } catch (error) {
      console.error('Failed to load workspaces:', error)
    }
  }

  initialized.value = true
})
</script>

<template>
  <el-config-provider :="elConfig">
    <RouterView />
  </el-config-provider>
</template>
