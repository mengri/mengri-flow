<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { useWorkspaceStore } from '@/stores/workspace'
import type { Workspace } from '@/types/workspace'
import AppLogo from '@/components/ui/AppLogo.vue'

const router = useRouter()
const workspaceStore = useWorkspaceStore()

// 状态
const pageLoading = ref(true)
const showCreateDialog = ref(false)
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
}

// 初始化：加载工作空间列表
onMounted(async () => {
  try {
    const status = await workspaceStore.loadWorkspaces()
    // 如果已经有选中的 workspace（从 localStorage 恢复），直接进主页面
    if (status === 'selected') {
      redirectToApp()
    }
  } catch {
    ElMessage.error('加载工作空间失败')
  } finally {
    pageLoading.value = false
  }
})

// 进入主页面
function redirectToApp() {
  const redirect = router.currentRoute.value.query.redirect as string
  const wsId = workspaceStore.currentWorkspaceId
  router.replace(redirect || (wsId ? `/workspace/${wsId}` : '/'))
}

// 选择工作空间
function handleSelect(workspace: Workspace) {
  workspaceStore.setCurrentWorkspace(workspace.id)
  ElMessage.success(`已切换到「${workspace.name}」`)
  redirectToApp()
}

// 创建工作空间
async function handleCreate() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  submitting.value = true
  try {
    const created = await workspaceStore.createWorkspace({
      name: formData.name,
      description: formData.description || undefined,
    })
    ElMessage.success('工作空间已创建')
    // 自动选中新创建的 workspace
    handleSelect(created)
  } catch {
    // client.ts 拦截器已处理错误提示
  } finally {
    submitting.value = false
  }
}

function openCreateDialog() {
  formData.name = ''
  formData.description = ''
  showCreateDialog.value = true
}

// 辅助方法
function getInitials(name: string) {
  if (!name) return ''
  return name
    .split(/[\s-_]+/)
    .map(part => part[0])
    .join('')
    .toUpperCase()
    .substring(0, 2)
}

function getAvatarColor(index: number) {
  const colors = [
    'linear-gradient(135deg, #667eea, #764ba2)',
    'linear-gradient(135deg, #f093fb, #f5576c)',
    'linear-gradient(135deg, #4facfe, #00f2fe)',
    'linear-gradient(135deg, #43e97b, #38f9d7)',
    'linear-gradient(135deg, #fa709a, #fee140)',
    'linear-gradient(135deg, #a18cd1, #fbc2eb)',
    'linear-gradient(135deg, #fccb90, #d57eeb)',
    'linear-gradient(135deg, #e0c3fc, #8ec5fc)',
  ]
  return colors[index % colors.length]
}
</script>

<template>
  <div class="workspace-select-page">
    <div class="workspace-select-container">
      <!-- 顶部 Logo -->
      <div class="logo-section">
        <AppLogo size="lg" />
      </div>

      <!-- 加载状态 -->
      <div v-if="pageLoading" class="loading-section">
        <el-icon class="is-loading" :size="32"><Loading /></el-icon>
        <p class="text-gray-500 mt-4">加载工作空间...</p>
      </div>

      <!-- 无工作空间 -->
      <div v-else-if="workspaceStore.workspaces.length === 0" class="empty-section">
        <el-empty description="还没有工作空间">
          <el-button type="primary" size="large" @click="openCreateDialog">
            创建第一个工作空间
          </el-button>
        </el-empty>
      </div>

      <!-- 工作空间选择 -->
      <div v-else class="select-section">
        <h2 class="section-title">选择工作空间</h2>
        <p class="section-desc">请选择一个工作空间进入</p>

        <div class="workspace-grid">
          <div
            v-for="(workspace, index) in workspaceStore.workspaces"
            :key="workspace.id"
            class="workspace-card"
            :style="{ '--avatar-bg': getAvatarColor(index) }"
            tabindex="0"
            role="button"
            :aria-label="`选择工作空间 ${workspace.name}`"
            @click="handleSelect(workspace)"
            @keyup.enter="handleSelect(workspace)"
          >
            <div class="card-avatar">
              {{ getInitials(workspace.name) }}
            </div>
            <div class="card-info">
              <div class="card-name">{{ workspace.name }}</div>
              <div class="card-desc">{{ workspace.description || '暂无描述' }}</div>
            </div>
            <el-icon class="card-arrow"><ArrowRight /></el-icon>
          </div>

          <!-- 创建新工作空间卡片 -->
          <div
            class="workspace-card workspace-card-new"
            tabindex="0"
            role="button"
            aria-label="创建新工作空间"
            @click="openCreateDialog"
            @keyup.enter="openCreateDialog"
          >
            <div class="card-avatar card-avatar-new">
              <el-icon :size="24"><Plus /></el-icon>
            </div>
            <div class="card-info">
              <div class="card-name">创建新工作空间</div>
              <div class="card-desc">添加一个新的工作空间</div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 创建对话框 -->
    <el-dialog
      v-model="showCreateDialog"
      title="创建工作空间"
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
        <el-button type="primary" :loading="submitting" @click="handleCreate">
          创建并进入
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script lang="ts">
import { Loading, ArrowRight, Plus } from '@element-plus/icons-vue'

export default {
  components: { Loading, ArrowRight, Plus },
}
</script>

<style scoped>
.workspace-select-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
}

.workspace-select-container {
  width: 100%;
  max-width: 720px;
  padding: 40px 24px;
}

.logo-section {
  text-align: center;
  margin-bottom: 48px;
}

.loading-section {
  text-align: center;
  padding: 80px 0;
}

.empty-section {
  text-align: center;
  padding: 60px 0;
}

.select-section {
  text-align: center;
}

.section-title {
  font-size: 24px;
  font-weight: 700;
  color: var(--el-text-color-primary);
  margin: 0 0 8px;
}

.section-desc {
  font-size: 14px;
  color: var(--el-text-color-secondary);
  margin: 0 0 32px;
}

.workspace-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 16px;
  text-align: left;
}

.workspace-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px;
  background: #fff;
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.2s ease;
  border: 1px solid transparent;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.06);
}

.workspace-card:hover {
  border-color: var(--el-color-primary-light-5);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  transform: translateY(-2px);
}

.workspace-card:focus-visible {
  outline: 2px solid var(--el-color-primary);
  outline-offset: 2px;
}

.card-avatar {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  border-radius: 10px;
  background: var(--avatar-bg, linear-gradient(135deg, #667eea, #764ba2));
  color: #fff;
  font-size: 16px;
  font-weight: 700;
  flex-shrink: 0;
}

.card-avatar-new {
  background: linear-gradient(135deg, #e2e8f0, #cbd5e1);
  color: var(--el-text-color-secondary);
}

.card-info {
  flex: 1;
  min-width: 0;
}

.card-name {
  font-size: 15px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  margin-bottom: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.card-desc {
  font-size: 13px;
  color: var(--el-text-color-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.card-arrow {
  color: var(--el-text-color-placeholder);
  transition: transform 0.2s ease;
}

.workspace-card:hover .card-arrow {
  transform: translateX(4px);
  color: var(--el-color-primary);
}
</style>
