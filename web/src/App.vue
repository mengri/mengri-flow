<script setup lang="ts">
import { onMounted, computed } from 'vue'
import { RouterView, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import AppLayout from '@/layouts/AppLayout.vue'

const route = useRoute()
const authStore = useAuthStore()

/** 公开页面不需要 layout（登录、激活页） */
const isPublicPage = computed(() => route.meta.public === true)

/** 需要显示 layout 的已认证页面 */
const showLayout = computed(() => !isPublicPage.value && authStore.isAuthenticated)

onMounted(async () => {
  // 如果已有 token，尝试加载用户资料
  if (authStore.isAuthenticated) {
    await authStore.fetchProfile()
  }
})
</script>

<template>
  <el-config-provider>
    <!-- 公开页面直接渲染 -->
    <RouterView v-if="isPublicPage" />

    <!-- 已认证页面带 layout -->
    <AppLayout v-else-if="showLayout">
      <RouterView />
    </AppLayout>

    <!-- 未认证的非公开页面也直接渲染（会被路由守卫跳转） -->
    <RouterView v-else />
  </el-config-provider>
</template>
