<template>
  <div class="tool-list">
    <div class="header">
      <h1>工具管理</h1>
      <div class="actions">
        <el-button @click="handleImport">
          <el-icon><Upload /></el-icon>
          批量导入
        </el-button>
        <el-button type="primary" @click="handleCreate">
          <el-icon><Plus /></el-icon>
          新建工具
        </el-button>
      </div>
    </div>

    <!-- 筛选 -->
    <div class="filters">
      <el-form :inline="true" :model="filters">
        <el-form-item>
          <el-input
            v-model="filters.search"
            placeholder="搜索工具名称"
            clearable
            style="width: 300px"
          />
        </el-form-item>
        <el-form-item>
          <el-select v-model="filters.resourceId" placeholder="选择资源" clearable style="width: auto; min-width: 180px">
            <el-option
              v-for="resource in resources"
              :key="resource.id"
              :label="resource.name"
              :value="resource.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-select v-model="filters.status" placeholder="状态" clearable style="width: auto; min-width: 180px">
            <el-option label="草稿" value="draft" />
            <el-option label="已发布" value="published" />
            <el-option label="已下线" value="deprecated" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">搜索</el-button>
        </el-form-item>
      </el-form>
    </div>

    <!-- 表格 -->
    <el-table :data="tools" v-loading="loading" border>
      <el-table-column prop="name" label="名称" min-width="150" />
      <el-table-column prop="resourceName" label="所属资源" width="150" />
      <el-table-column prop="type" label="类型" width="100" />
      <el-table-column prop="method" label="方法" width="100" />
      <el-table-column prop="path" label="路径" min-width="200" show-overflow-tooltip />
      <el-table-column prop="status" label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="statusTagType(row.status)">
            {{ statusText(row.status) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="tags" label="标签" width="150">
        <template #default="{ row }">
          <el-tag
            v-for="tag in row.tags.slice(0, 3)"
            :key="tag"
            size="small"
            style="margin-right: 5px"
          >
            {{ tag }}
          </el-tag>
          <span v-if="row.tags.length > 3">...</span>
        </template>
      </el-table-column>
      <el-table-column prop="createdAt" label="创建时间" width="180">
        <template #default="{ row }">
          {{ formatDate(row.createdAt) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="250" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="handleTest(row)">测试</el-button>
          <el-button size="small" @click="handleEdit(row.id)">编辑</el-button>
          <el-button
            size="small"
            :type="row.status === 'published' ? 'warning' : 'success'"
            @click="handleToggleStatus(row)"
          >
            {{ row.status === 'published' ? '下线' : '发布' }}
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
import { toolAPI } from '@/api/tools'
import { resourceAPI } from '@/api/resources'
import type { Tool } from '@/types/tool'
import type { Resource } from '@/types/resource'
import { useWorkspaceStore } from '@/stores/workspace'
import { useWorkspaceRoute } from '@/composables/useWorkspaceRoute'
import { formatDate } from '@/utils/request'

const router = useRouter()
const workspaceStore = useWorkspaceStore()
const { createToolPath, importToolPath, toolDetailPath, toolsPath } = useWorkspaceRoute()

const loading = ref(false)
const tools = ref<Tool[]>([])
const resources = ref<Resource[]>([])

const filters = reactive({
  search: '',
  resourceId: '',
  status: '',
})

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0,
})

async function loadTools() {
  if (!workspaceStore.currentWorkspaceId) return
  loading.value = true
  try {
    const data = await toolAPI.list({
      workspaceId: workspaceStore.currentWorkspaceIdOrThrow,
      resourceId: filters.resourceId || undefined,
      status: filters.status || undefined,
    })
    
    // 加载资源信息
    await loadResources()
    
    // 关联资源名称
    tools.value = (data.list || []).map((tool) => ({
      ...tool,
      resourceName: resources.value.find((r) => r.id === tool.resourceId)?.name || '-',
    }))
    
    pagination.total = data.total
  } catch (error) {
    ElMessage.error('加载工具列表失败')
  } finally {
    loading.value = false
  }
}

async function loadResources() {
  if (!workspaceStore.currentWorkspaceId) return
  try {
    const data = await resourceAPI.list({
      workspaceId: workspaceStore.currentWorkspaceIdOrThrow,
    })
    resources.value = data.list || []
  } catch (error) {
    console.error('加载资源列表失败', error)
  }
}

function handleCreate() {
  router.push(createToolPath())
}

function handleImport() {
  router.push(importToolPath())
}

function handleEdit(id: string) {
  router.push(toolDetailPath(id))
}

async function handleTest(tool: Tool) {
  try {
    await toolAPI.test({
      toolId: tool.id,
      input: {},
    })
    ElMessage.success('工具测试成功')
  } catch (error) {
    ElMessage.error('工具测试失败')
  }
}

async function handleToggleStatus(tool: Tool) {
  try {
    if (tool.status === 'published') {
      await toolAPI.deprecate(tool.id)
      ElMessage.success('工具已下线')
    } else {
      await toolAPI.publish(tool.id)
      ElMessage.success('工具已发布')
    }
    loadTools()
  } catch (error) {
    ElMessage.error('操作失败')
  }
}

function handleDelete(id: string) {
  ElMessageBox.confirm('确定要删除该工具吗？', '警告', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning',
  })
    .then(async () => {
      await toolAPI.delete(id)
      ElMessage.success('删除成功')
      loadTools()
    })
    .catch(() => {})
}

function handleSearch() {
  pagination.page = 1
  loadTools()
}

function handleSizeChange() {
  loadTools()
}

function handleCurrentChange() {
  loadTools()
}

function statusTagType(status: string) {
  const map: Record<string, string> = {
    published: 'success',
    draft: 'info',
    deprecated: 'warning',
  }
  return map[status] || 'info'
}

function statusText(status: string) {
  const map: Record<string, string> = {
    published: '已发布',
    draft: '草稿',
    deprecated: '已下线',
  }
  return map[status] || status
}

onMounted(() => {
  loadTools()
})

watch(() => workspaceStore.workspaces.length, (len) => {
  if (len > 0 && workspaceStore.currentWorkspaceId) loadTools()
})
</script>

<style scoped>
.tool-list {
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
