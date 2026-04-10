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
import { useWorkspaceRoute } from '@/composables/useWorkspaceRoute'

const router = useRouter()
const { toolsPath } = useWorkspaceRoute()

function handleSuccess() {
  router.push(toolsPath())
}

function handleCancel() {
  router.push(toolsPath())
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
