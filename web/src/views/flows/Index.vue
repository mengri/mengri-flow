<template>
  <div class="flow-list">
    <div class="header">
      <h1>流程管理</h1>
      <div class="actions">
        <el-button type="primary" @click="handleCreate">
          <el-icon><Plus /></el-icon>
          新建流程
        </el-button>
      </div>
    </div>

    <!-- 筛选 -->
    <div class="filters">
      <el-form :inline="true" :model="filters">
        <el-form-item>
          <el-input
            v-model="filters.search"
            placeholder="搜索流程名称"
            clearable
            style="width: 300px"
          />
        </el-form-item>
        <el-form-item>
          <el-select v-model="filters.status" placeholder="状态" clearable>
            <el-option label="草稿" value="draft" />
            <el-option label="已发布" value="published" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">搜索</el-button>
        </el-form-item>
      </el-form>
    </div>

    <!-- 表格 -->
    <el-table :data="flows" v-loading="loading" border>
      <el-table-column prop="name" label="名称" min-width="200" />
      <el-table-column prop="description" label="描述" min-width="300" show-overflow-tooltip />
      <el-table-column prop="status" label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="statusTagType(row.status)">
            {{ statusText(row.status) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="currentVersion" label="版本" width="80" />
      <el-table-column prop="createdAt" label="创建时间" width="180">
        <template #default="{ row }">
          {{ formatDate(row.createdAt) }}
        </template>
      </el-table-column>
      <el-table-column prop="updatedAt" label="更新时间" width="180">
        <template #default="{ row }">
          {{ formatDate(row.updatedAt) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="300" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="handleCanvas(row.id)">画布</el-button>
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
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { flowAPI } from '@/api/flows'
import type { Flow } from '@/types/flow'
import { useWorkspaceStore } from '@/stores/workspace'
import { formatDate } from '@/utils/request'

const router = useRouter()
const workspaceStore = useWorkspaceStore()

const loading = ref(false)
const flows = ref<Flow[]>([])

const filters = reactive({
  search: '',
  status: '',
})

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0,
})

async function loadFlows() {
  loading.value = true
  try {
    const data = await flowAPI.list({
      workspaceId: workspaceStore.currentWorkspaceIdOrThrow,
      status: filters.status || undefined,
    })
    flows.value = data
    pagination.total = data.length
  } catch (error) {
    ElMessage.error('加载流程失败')
  } finally {
    loading.value = false
  }
}

function handleCreate() {
  router.push('/flows/new')
}

function handleCanvas(id: string) {
  router.push(`/flows/${id}`)
}

function handleEdit(id: string) {
  router.push(`/flows/${id}/edit`)
}

async function handleToggleStatus(flow: Flow) {
  try {
    if (flow.status === 'published') {
      await flowAPI.update(flow.id, { status: 'draft' })
      ElMessage.success('流程已下线')
    } else {
      await flowAPI.publish(flow.id)
      ElMessage.success('流程已发布')
    }
    loadFlows()
  } catch (error) {
    ElMessage.error('操作失败')
  }
}

function handleDelete(id: string) {
  ElMessageBox.confirm('确定要删除该流程吗？', '警告', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning',
  })
    .then(async () => {
      await flowAPI.delete(id)
      ElMessage.success('删除成功')
      loadFlows()
    })
    .catch(() => {})
}

function handleSearch() {
  pagination.page = 1
  loadFlows()
}

function handleSizeChange() {
  loadFlows()
}

function handleCurrentChange() {
  loadFlows()
}

function statusTagType(status: string) {
  const map: Record<string, string> = {
    published: 'success',
    draft: 'info',
  }
  return map[status] || 'info'
}

function statusText(status: string) {
  const map: Record<string, string> = {
    published: '已发布',
    draft: '草稿',
  }
  return map[status] || status
}

onMounted(() => {
  loadFlows()
})
</script>

<style scoped>
.flow-list {
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
