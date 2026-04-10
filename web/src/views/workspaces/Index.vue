<template>
  <div class="workspace-list">
    <div class="header">
      <h1>工作空间管理</h1>
      <div class="actions">
        <el-button type="primary" @click="showCreateDialog = true">
          <el-icon><Plus /></el-icon>
          新建工作空间
        </el-button>
      </div>
    </div>

    <!-- 表格 -->
    <el-table :data="workspaceStore.workspaces" v-loading="workspaceStore.loading" border>
      <el-table-column prop="name" label="名称" min-width="180">
        <template #default="{ row }">
          <div class="workspace-name-cell">
            <span class="workspace-avatar">
              {{ getInitials(row.name) }}
            </span>
            <span>{{ row.name }}</span>
            <el-tag v-if="row.id === workspaceStore.currentWorkspaceId" size="small" type="success" class="current-tag">
              当前
            </el-tag>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="description" label="描述" min-width="240" show-overflow-tooltip>
        <template #default="{ row }">
          {{ row.description || '-' }}
        </template>
      </el-table-column>
      <el-table-column prop="memberCount" label="成员数" width="90" align="center" />
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
      <el-table-column label="操作" width="220" fixed="right">
        <template #default="{ row }">
          <el-button
            v-if="row.id !== workspaceStore.currentWorkspaceId"
            size="small"
            type="primary"
            @click="handleSwitch(row.id)"
          >
            切换
          </el-button>
          <el-button size="small" @click="handleEdit(row)">编辑</el-button>
          <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 创建/编辑对话框 -->
    <el-dialog
      v-model="showCreateDialog"
      :title="editingWorkspace ? '编辑工作空间' : '新建工作空间'"
      width="480"
      destroy-on-close
    >
      <el-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        label-width="80px"
        label-position="right"
      >
        <el-form-item label="名称" prop="name">
          <el-input
            v-model="formData.name"
            placeholder="输入工作空间名称"
            maxlength="100"
            show-word-limit
          />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input
            v-model="formData.description"
            type="textarea"
            placeholder="输入工作空间描述（可选）"
            :rows="3"
            maxlength="500"
            show-word-limit
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">
          {{ editingWorkspace ? '保存' : '创建' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { useWorkspaceStore } from '@/stores/workspace'
import { formatDate } from '@/utils/request'
import type { Workspace } from '@/types/workspace'

const workspaceStore = useWorkspaceStore()

// 对话框状态
const showCreateDialog = ref(false)
const editingWorkspace = ref<Workspace | null>(null)
const submitting = ref(false)
const formRef = ref<FormInstance>()

const formData = reactive({
  name: '',
  description: '',
})

const formRules: FormRules = {
  name: [
    { required: true, message: '请输入工作空间名称', trigger: 'blur' },
    { max: 100, message: '名称不能超过100个字符', trigger: 'blur' },
  ],
  description: [
    { max: 500, message: '描述不能超过500个字符', trigger: 'blur' },
  ],
}

function resetForm() {
  formData.name = ''
  formData.description = ''
  editingWorkspace.value = null
}

function handleEdit(workspace: Workspace) {
  editingWorkspace.value = workspace
  formData.name = workspace.name
  formData.description = workspace.description
  showCreateDialog.value = true
}

async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  submitting.value = true
  try {
    if (editingWorkspace.value) {
      await workspaceStore.updateWorkspace(editingWorkspace.value.id, {
        name: formData.name,
        description: formData.description || undefined,
      })
      ElMessage.success('工作空间已更新')
    } else {
      const created = await workspaceStore.createWorkspace({
        name: formData.name,
        description: formData.description || undefined,
      })
      ElMessage.success('工作空间已创建')
      // 自动切换到新创建的工作空间
      workspaceStore.setCurrentWorkspace(created.id)
    }
    showCreateDialog.value = false
    resetForm()
  } catch (error) {
    // client.ts 拦截器已处理错误提示
  } finally {
    submitting.value = false
  }
}

function handleSwitch(workspaceId: string) {
  workspaceStore.setCurrentWorkspace(workspaceId)
  ElMessage.success('已切换工作空间')
  // 刷新页面数据
  window.location.reload()
}

function handleDelete(workspace: Workspace) {
  const isCurrent = workspace.id === workspaceStore.currentWorkspaceId

  ElMessageBox.confirm(
    `确定要删除工作空间「${workspace.name}」吗？${isCurrent ? '删除后将自动切换到其他工作空间。' : ''}`,
    '删除工作空间',
    {
      confirmButtonText: '确定删除',
      cancelButtonText: '取消',
      type: 'warning',
      confirmButtonClass: 'el-button--danger',
    },
  )
    .then(async () => {
      await workspaceStore.deleteWorkspace(workspace.id)
      ElMessage.success('工作空间已删除')
    })
    .catch(() => {})
}

function getInitials(name: string) {
  if (!name) return ''
  return name
    .split(/[\s-_]+/)
    .map(part => part[0])
    .join('')
    .toUpperCase()
    .substring(0, 2)
}

onMounted(() => {
  workspaceStore.loadWorkspaces()
})
</script>

<style scoped>
.workspace-list {
  padding: 20px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.workspace-name-cell {
  display: flex;
  align-items: center;
  gap: 10px;
}

.workspace-avatar {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: 6px;
  background: linear-gradient(135deg, var(--el-color-primary-light-3), var(--el-color-primary));
  color: #fff;
  font-size: 12px;
  font-weight: 600;
  flex-shrink: 0;
}

.current-tag {
  margin-left: 4px;
}
</style>
