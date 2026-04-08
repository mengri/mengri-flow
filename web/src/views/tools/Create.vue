<template>
  <div class="create-tool">
    <div class="header">
      <h1>创建工具</h1>
    </div>

    <el-card>
      <ToolForm @success="handleSuccess" @cancel="handleCancel" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import ToolForm from '@/components/ToolForm.vue'

const router = useRouter()

function handleSuccess() {
  router.push('/tools')
}

function handleCancel() {
  router.push('/tools')
}
</script>

<style scoped>
.create-tool {
  padding: 20px;
}

.header {
  margin-bottom: 20px;
}
</style>
