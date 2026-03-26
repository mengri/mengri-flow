<script setup lang="ts">
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
</script>

<template>
  <div class="flex items-center justify-center py-20">
    <div class="text-center">
      <h1 class="text-4xl font-bold text-gray-800 mb-4">Mengri Flow</h1>
      <p class="text-gray-500 mb-8">Clean Architecture + DDD + Vue 3</p>

      <div v-if="authStore.isAuthenticated" class="flex gap-4 justify-center">
        <el-button type="primary" @click="$router.push('/account')">
          Account Center
        </el-button>
        <el-button v-if="authStore.isAdmin" type="warning" @click="$router.push('/admin/accounts')">
          Account Management
        </el-button>
      </div>
      <div v-else>
        <el-button type="primary" @click="$router.push('/login')">
          Sign In
        </el-button>
      </div>
    </div>
  </div>
</template>
