<template>
  <div class="create-resource">
    <div class="header">
      <h1>创建资源</h1>
    </div>

    <el-card>
      <ResourceForm @success="handleSuccess" @cancel="handleCancel" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import ResourceForm from '@/components/ResourceForm.vue'

const router = useRouter()

function handleSuccess() {
  router.push('/resources')
}

function handleCancel() {
  router.push('/resources')
}
</script>

<style scoped>
.create-resource {
  padding: 20px;
}

.header {
  margin-bottom: 20px;
}
</style>
