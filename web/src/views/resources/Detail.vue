<template>
  <div class="resource-detail" v-if="resource">
    <div class="header">
      <router-link :to="resourcesPath()" class="back-link">
        <el-icon><ArrowLeft /></el-icon>
        返回列表
      </router-link>
      <h1>资源详情: {{ resource.name }}</h1>
    </div>

    <el-card>
      <el-descriptions :column="2" border>
        <el-descriptions-item label="名称">{{ resource.name }}</el-descriptions-item>
        <el-descriptions-item label="类型">
          <el-tag>{{ resource.type }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="statusTagType(resource.status)">
            {{ statusText(resource.status) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="工作空间">{{ resource.workspaceId }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatDate(resource.createdAt) }}</el-descriptions-item>
        <el-descriptions-item label="更新时间">{{ formatDate(resource.updatedAt) }}</el-descriptions-item>
        <el-descriptions-item label="描述" :span="2">{{ resource.description }}</el-descriptions-item>
        <el-descriptions-item label="配置" :span="2">
          <pre>{{ JSON.stringify(resource.config, null, 2) }}</pre>
        </el-descriptions-item>
      </el-descriptions>

      <div class="actions">
        <el-button @click="handleTest">测试连接</el-button>
        <el-button type="primary" @click="handleEdit">编辑</el-button>
        <el-button type="danger" @click="handleDelete">删除</el-button>
      </div>
    </el-card>

    <!-- 关联的工具 -->
    <el-card class="tools-section">
      <template #header>
        <span>关联工具</span>
      </template>
      <el-table :data="tools" v-loading="loadingTools">
        <el-table-column prop="name" label="工具名称" />
        <el-table-column prop="type" label="类型" />
        <el-table-column prop="method" label="方法" />
        <el-table-column prop="status" label="状态">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)">
              {{ statusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150">
          <template #default="{ row }">
            <el-button size="small" @click="handleViewTool(row.id)">查看</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useWorkspaceStore } from '@/stores/workspace'
import { resourceAPI } from '@/api/resources'
import { toolAPI } from '@/api/tools'
import type { Resource } from '@/types/resource'
import type { Tool } from '@/types/tool'
import { formatDate } from '@/utils/request'
import { useWorkspaceRoute } from '@/composables/useWorkspaceRoute'

const route = useRoute()
const router = useRouter()
const { resourcesPath, resourceDetailPath, toolDetailPath } = useWorkspaceRoute()

const resource = ref<Resource>()
const tools = ref<Tool[]>([])
const loadingTools = ref(false)

async function loadResource() {
  const id = route.params.id as string
  try {
    resource.value = await resourceAPI.get(id)
    loadTools()
  } catch (error) {
    ElMessage.error('加载资源详情失败')
  }
}

async function loadTools() {
  if (!resource.value) return
  loadingTools.value = true
  try {
    const workspaceStore = useWorkspaceStore()
    const data = await toolAPI.list({
      workspaceId: workspaceStore.currentWorkspaceIdOrThrow,
      resourceId: resource.value.id,
    })
    tools.value = data.list || []
  } catch (error) {
    ElMessage.error('加载工具列表失败')
  } finally {
    loadingTools.value = false
  }
}

async function handleTest() {
  if (!resource.value) return
  try {
    await resourceAPI.testConnection({
      type: resource.value.type,
      config: resource.value.config,
    })
    ElMessage.success('连接测试成功')
  } catch (error) {
    ElMessage.error('连接测试失败')
  }
}

function handleEdit() {
  if (!resource.value) return
  router.push(resourceDetailPath(resource.value.id) + '/edit')
}

function handleDelete() {
  if (!resource.value) return
  ElMessageBox.confirm('确定要删除该资源吗？', '警告', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning',
  })
    .then(async () => {
      await resourceAPI.delete(resource.value!.id)
      ElMessage.success('删除成功')
      router.push(resourcesPath())
    })
    .catch(() => {})
}

function handleViewTool(id: string) {
  router.push(toolDetailPath(id))
}

function statusTagType(status: string) {
  const map: Record<string, string> = {
    active: 'success',
    published: 'success',
    error: 'danger',
    failed: 'danger',
    inactive: 'info',
    draft: 'info',
    running: 'warning',
    timeout: 'info',
  }
  return map[status] || 'info'
}

function statusText(status: string) {
  const map: Record<string, string> = {
    active: '正常',
    published: '已发布',
    error: '异常',
    failed: '失败',
    inactive: '未激活',
    draft: '草稿',
    running: '运行中',
    timeout: '超时',
  }
  return map[status] || status
}

onMounted(() => {
  loadResource()
})
</script>

<style scoped>
.resource-detail {
  padding: 20px;
}

.header {
  display: flex;
  align-items: center;
  margin-bottom: 20px;
}

.back-link {
  display: flex;
  align-items: center;
  color: #409eff;
  text-decoration: none;
  margin-right: 20px;
}

.actions {
  margin-top: 20px;
  text-align: center;
}

.tools-section {
  margin-top: 30px;
}

pre {
  background: #f5f7fa;
  padding: 10px;
  border-radius: 4px;
  overflow-x: auto;
}
</style>
