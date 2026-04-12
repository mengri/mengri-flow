<template>
  <div class="resource-list">
    <div class="header">
      <h1>资源管理</h1>
      <div class="actions">
        <el-button type="primary" @click="handleCreate">
          <el-icon><Plus /></el-icon>
          新建资源
        </el-button>
      </div>
    </div>

    <!-- 筛选 -->
    <div class="filters">
      <el-form :inline="true" :model="filters">
        <el-form-item label="类型">
          <el-select v-model="filters.type" placeholder="全部" clearable style="width: auto; min-width: 180px">
            <el-option label="HTTP" value="http" />
            <el-option label="gRPC" value="grpc" />
            <el-option label="MySQL" value="mysql" />
            <el-option label="PostgreSQL" value="postgres" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: auto; min-width: 180px">
            <el-option label="正常" value="active" />
            <el-option label="异常" value="error" />
            <el-option label="未激活" value="inactive" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-input v-model="filters.search" placeholder="搜索资源名称" clearable />
        </el-form-item>
        <el-form-item>
          <el-button @click="handleSearch">搜索</el-button>
        </el-form-item>
      </el-form>
    </div>

    <!-- 表格 -->
    <el-table :data="resources" v-loading="loading" border>
      <el-table-column prop="name" label="名称" min-width="150" />
      <el-table-column prop="type" label="类型" width="100">
        <template #default="{ row }">
          <el-tag>{{ row.type }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="status" label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="statusTagType(row.status)">
            {{ statusText(row.status) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
      <el-table-column prop="createdAt" label="创建时间" width="180">
        <template #default="{ row }">
          {{ formatDate(row.createdAt) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="250" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="handleTest(row.id)">测试</el-button>
          <el-button size="small" @click="handleEdit(row.id)">编辑</el-button>
          <el-button size="small" type="primary" @click="handleExtractTools(row.id)">
            提取工具
          </el-button>
          <el-button size="small" type="danger" @click="handleDelete(row.id)">
            删除
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 分页 -->
    <div class="pagination">
      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { resourceAPI } from '@/api/resources'
import type { Resource } from '@/types/resource'
import { useWorkspaceStore } from '@/stores/workspace'
import { useWorkspaceRoute } from '@/composables/useWorkspaceRoute'
import { formatDate } from '@/utils/request'

const router = useRouter()
const workspaceStore = useWorkspaceStore()
const { createResourcePath, resourceDetailPath, toolsPath } = useWorkspaceRoute()

const loading = ref(false)
const resources = ref<Resource[]>([])

const filters = reactive({
  type: '',
  status: '',
  search: '',
})

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0,
})

async function loadResources() {
  if (!workspaceStore.currentWorkspaceId) return
  loading.value = true
  try {
    const data = await resourceAPI.list({
      workspaceId: workspaceStore.currentWorkspaceIdOrThrow,
      type: filters.type || undefined,
      status: filters.status || undefined,
    })
    resources.value = data.list || []
    pagination.total = data.total
  } catch (error) {
    ElMessage.error('加载资源失败')
  } finally {
    loading.value = false
  }
}

function handleCreate() {
  router.push(createResourcePath())
}

function handleEdit(id: string) {
  router.push(resourceDetailPath(id))
}

async function handleTest(id: string) {
  try {
    const resource = await resourceAPI.get(id)
    await resourceAPI.testConnection({ type: resource.type, config: resource.config })
    ElMessage.success('连接测试成功')
  } catch (error) {
    ElMessage.error('连接测试失败')
  }
}

async function handleExtractTools(id: string) {
  try {
    const tools = await resourceAPI.extractTools(id)
    ElMessage.success(`成功提取 ${tools.length} 个工具`)
    router.push(toolsPath())
  } catch (error) {
    ElMessage.error('提取工具失败')
  }
}

function handleDelete(id: string) {
  ElMessageBox.confirm('确定要删除该资源吗？', '警告', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning',
  })
    .then(async () => {
      await resourceAPI.delete(id)
      ElMessage.success('删除成功')
      loadResources()
    })
    .catch(() => {})
}

function handleSearch() {
  pagination.page = 1
  loadResources()
}

function handleSizeChange() {
  loadResources()
}

function handleCurrentChange() {
  loadResources()
}

function statusTagType(status: string) {
  const map = {
    active: 'success',
    error: 'danger',
    inactive: 'info',
  }
  return map[status as keyof typeof map] || 'info'
}

function statusText(status: string) {
  const map = {
    active: '正常',
    error: '异常',
    inactive: '未激活',
  }
  return map[status as keyof typeof map] || status
}

onMounted(() => {
  loadResources()
})

watch(() => workspaceStore.workspaces.length, (len) => {
  if (len > 0 && workspaceStore.currentWorkspaceId) loadResources()
})
</script>

<style scoped>
.resource-list {
  padding: 20px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.filters {
  margin-bottom: 20px;
  padding: 20px;
  background: #f5f7fa;
  border-radius: 4px;
}

.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}
</style>
