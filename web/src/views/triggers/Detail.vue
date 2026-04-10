<template>
  <div class="trigger-detail" v-if="trigger">
    <div class="header">
      <router-link :to="triggersPath()" class="back-link">
        <el-icon><ArrowLeft /></el-icon>
        返回列表
      </router-link>
      <h1>触发器详情: {{ trigger.name }}</h1>
      <div class="actions">
        <el-button
          :type="trigger.status === 'active' ? 'warning' : 'success'"
          @click="handleToggleStatus"
        >
          {{ trigger.status === 'active' ? '停止' : '启动' }}
        </el-button>
        <el-button type="danger" @click="handleDelete">删除</el-button>
      </div>
    </div>

    <!-- 基本信息 -->
    <el-card class="info-card">
      <template #header>
        <span>基本信息</span>
      </template>
      <el-descriptions :column="2" border>
        <el-descriptions-item label="名称">{{ trigger.name }}</el-descriptions-item>
        <el-descriptions-item label="类型">
          <el-tag>{{ trigger.type }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="statusTagType(trigger.status)">
            {{ statusText(trigger.status) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="绑定流程">{{ trigger.flowId }}</el-descriptions-item>
        <el-descriptions-item label="集群">{{ trigger.clusterId }}</el-descriptions-item>
        <el-descriptions-item label="版本">v{{ trigger.flowVersion }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatDate(trigger.createdAt) }}</el-descriptions-item>
        <el-descriptions-item label="更新时间">{{ formatDate(trigger.updatedAt) }}</el-descriptions-item>
      </el-descriptions>
    </el-card>

    <!-- 配置信息 -->
    <el-card class="config-card">
      <template #header>
        <span>配置信息</span>
      </template>
      <el-descriptions :column="1" border>
        <el-descriptions-item label="配置">
          <pre>{{ JSON.stringify(trigger.config, null, 2) }}</pre>
        </el-descriptions-item>
        <el-descriptions-item label="输入映射">
          <pre>{{ JSON.stringify(trigger.inputMapping, null, 2) }}</pre>
        </el-descriptions-item>
        <el-descriptions-item label="输出映射">
          <pre>{{ JSON.stringify(trigger.outputMapping, null, 2) }}</pre>
        </el-descriptions-item>
        <el-descriptions-item label="错误处理">
          <pre>{{ JSON.stringify(trigger.errorHandling, null, 2) }}</pre>
        </el-descriptions-item>
      </el-descriptions>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { triggerAPI } from '@/api/triggers'
import type { Trigger } from '@/types/trigger'
import { formatDate } from '@/utils/request'
import { useWorkspaceRoute } from '@/composables/useWorkspaceRoute'

const route = useRoute()
const router = useRouter()
const { triggersPath } = useWorkspaceRoute()

const trigger = ref<Trigger>()

async function loadTrigger() {
  const id = route.params.id as string
  try {
    trigger.value = await triggerAPI.get(id)
  } catch (error) {
    ElMessage.error('加载触发器详情失败')
  }
}

async function handleToggleStatus() {
  if (!trigger.value) return
  
  try {
    if (trigger.value.status === 'active') {
      await triggerAPI.stop(trigger.value.id)
      ElMessage.success('触发器已停止')
    } else {
      await triggerAPI.start(trigger.value.id)
      ElMessage.success('触发器已启动')
    }
    loadTrigger()
  } catch (error) {
    ElMessage.error('操作失败')
  }
}

function handleDelete() {
  if (!trigger.value) return
  
  ElMessageBox.confirm('确定要删除该触发器吗？', '警告', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning',
  })
    .then(async () => {
      await triggerAPI.delete(trigger.value!.id)
      ElMessage.success('删除成功')
      router.push(triggersPath())
    })
    .catch(() => {})
}

function statusTagType(status: string) {
  const map: Record<string, string> = {
    active: 'success',
    inactive: 'info',
  }
  return map[status] || 'info'
}

function statusText(status: string) {
  const map: Record<string, string> = {
    active: '运行中',
    inactive: '已停止',
  }
  return map[status] || status
}

onMounted(() => {
  loadTrigger()
})
</script>

<style scoped>
.trigger-detail {
  padding: 20px;
}

.header {
  display: flex;
  align-items: center;
  margin-bottom: 30px;
}

.back-link {
  display: flex;
  align-items: center;
  color: #409eff;
  text-decoration: none;
  margin-right: 20px;
}

.actions {
  margin-left: auto;
}

.info-card {
  margin-bottom: 30px;
}

.config-card {
  margin-bottom: 30px;
}

pre {
  background: #f5f7fa;
  padding: 10px;
  border-radius: 4px;
  overflow-x: auto;
  margin: 0;
}
</style>
