<template>
  <div class="trigger-list">
    <div class="header">
      <h1>触发器管理</h1>
      <div class="actions">
        <el-button type="primary" @click="handleCreate">
          <el-icon><Plus /></el-icon>
          新建触发器
        </el-button>
      </div>
    </div>

    <!-- 筛选 -->
    <div class="filters">
      <el-form :inline="true" :model="filters">
        <el-form-item>
          <el-input
            v-model="filters.search"
            placeholder="搜索触发器名称"
            clearable
            style="width: 300px"
          />
        </el-form-item>
        <el-form-item>
          <el-select v-model="filters.type" placeholder="类型" clearable>
            <el-option label="RESTful" value="restful" />
            <el-option label="定时任务" value="timer" />
            <el-option label="RabbitMQ" value="rabbitmq" />
            <el-option label="Kafka" value="kafka" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-select v-model="filters.status" placeholder="状态" clearable>
            <el-option label="运行中" value="active" />
            <el-option label="已停止" value="inactive" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">搜索</el-button>
        </el-form-item>
      </el-form>
    </div>

    <!-- 表格 -->
    <el-table :data="triggers" v-loading="loading" border>
      <el-table-column prop="name" label="名称" min-width="200" />
      <el-table-column prop="type" label="类型" width="100">
        <template #default="{ row }">
          <el-tag>{{ row.type }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="flowName" label="绑定流程" min-width="200" />
      <el-table-column prop="status" label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="statusTagType(row.status)">
            {{ statusText(row.status) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="clusterId" label="集群" width="150" />
      <el-table-column prop="createdAt" label="创建时间" width="180">
        <template #default="{ row }">
          {{ formatDate(row.createdAt) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="250" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="handleView(row.id)">查看</el-button>
          <el-button
            size="small"
            :type="row.status === 'active' ? 'warning' : 'success'"
            @click="handleToggleStatus(row)"
          >
            {{ row.status === 'active' ? '停止' : '启动' }}
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
import { triggerAPI } from '@/api/triggers'
import type { Trigger } from '@/types/trigger'
import { useWorkspaceStore } from '@/stores/workspace'
import { formatDate } from '@/utils/request'

const router = useRouter()
const workspaceStore = useWorkspaceStore()

const loading = ref(false)
const triggers = ref<Trigger[]>([])

const filters = reactive({
  search: '',
  type: '',
  status: '',
})

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0,
})

async function loadTriggers() {
  loading.value = true
  try {
    const data = await triggerAPI.list({
      workspaceId: workspaceStore.currentWorkspaceIdOrThrow,
    })
    triggers.value = data.list
    pagination.total = data.total
  } catch (error) {
    ElMessage.error('加载触发器失败')
  } finally {
    loading.value = false
  }
}

function handleCreate() {
  router.push('/triggers/new')
}

function handleView(id: string) {
  router.push(`/triggers/${id}`)
}

async function handleToggleStatus(trigger: Trigger) {
  try {
    if (trigger.status === 'active') {
      await triggerAPI.stop(trigger.id)
      ElMessage.success('触发器已停止')
    } else {
      await triggerAPI.start(trigger.id)
      ElMessage.success('触发器已启动')
    }
    loadTriggers()
  } catch (error) {
    ElMessage.error('操作失败')
  }
}

function handleDelete(id: string) {
  ElMessageBox.confirm('确定要删除该触发器吗？', '警告', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning',
  })
    .then(async () => {
      await triggerAPI.delete(id)
      ElMessage.success('删除成功')
      loadTriggers()
    })
    .catch(() => {})
}

function handleSearch() {
  pagination.page = 1
  loadTriggers()
}

function handleSizeChange() {
  loadTriggers()
}

function handleCurrentChange() {
  loadTriggers()
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
  loadTriggers()
})
</script>

<style scoped>
.trigger-list {
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
