<script setup lang="ts">
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-gray-50">
    <div class="text-center">
      <h1 class="text-4xl font-bold text-gray-800 mb-4">Mengri Flow</h1>
      <p class="text-gray-500 mb-8">Clean Architecture + DDD + Vue 3</p>
      <el-button type="primary" @click="$router.push('/users')">
        User Management
      </el-button>
    </div>
  </div>
</template>
