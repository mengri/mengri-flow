<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import { ElMessage, ElMessageBox } from 'element-plus'
import { workspaceAPI } from '@/api/workspaces'
import { useWorkspaceRoute } from '@/composables/useWorkspaceRoute'
import { useI18n } from 'vue-i18n'
import type { WorkspaceMember } from '@/types/workspace'

const { t } = useI18n()
const { workspaceId } = useWorkspaceRoute()

const members = ref<WorkspaceMember[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(false)

// Add member dialog
const addDialogVisible = ref(false)
const addFormRef = ref<FormInstance>()
const addForm = reactive({
  accountId: '',
  role: 'member' as 'member' | 'admin',
})
const addSubmitting = ref(false)

const addRules: FormRules = {
  accountId: [
    { required: true, message: t('workspace.accountIdRequired'), trigger: 'blur' },
  ],
}

const roleOptions = [
  { label: t('workspace.roleMember'), value: 'member' },
  { label: t('workspace.roleAdmin'), value: 'admin' },
]

const roleTagType: Record<string, string> = {
  owner: 'danger',
  admin: 'warning',
  member: 'info',
}

async function fetchMembers() {
  loading.value = true
  try {
    const result = await workspaceAPI.listMembers(workspaceId.value, {
      page: page.value,
      pageSize: pageSize.value,
    })
    members.value = result.list
    total.value = result.total
  } finally {
    loading.value = false
  }
}

function onPageChange(p: number) {
  page.value = p
  fetchMembers()
}

function onPageSizeChange(size: number) {
  page.value = 1
  pageSize.value = size
  fetchMembers()
}

function openAddDialog() {
  addForm.accountId = ''
  addForm.role = 'member'
  addDialogVisible.value = true
}

async function submitAdd() {
  const valid = await addFormRef.value?.validate().catch(() => false)
  if (!valid) return

  addSubmitting.value = true
  try {
    await workspaceAPI.addMember(workspaceId.value, {
      accountId: addForm.accountId,
      role: addForm.role,
    })
    ElMessage.success(t('workspace.addMemberSuccess'))
    addDialogVisible.value = false
    fetchMembers()
  } finally {
    addSubmitting.value = false
  }
}

async function handleRemove(member: WorkspaceMember) {
  try {
    await ElMessageBox.confirm(
      t('workspace.removeMemberConfirm', { name: member.displayName || member.email }),
      t('common.warning'),
      { type: 'warning', confirmButtonText: t('common.confirm'), cancelButtonText: t('common.cancel') },
    )
    await workspaceAPI.removeMember(workspaceId.value, member.accountId)
    ElMessage.success(t('workspace.removeMemberSuccess'))
    fetchMembers()
  } catch {
    // cancelled
  }
}

onMounted(() => {
  fetchMembers()
})
</script>

<template>
  <div class="workspace-members">
    <div class="flex items-center justify-between mb-6">
      <div>
        <h2 class="text-xl font-bold text-gray-800">{{ t('workspace.members') }}</h2>
        <p class="text-sm text-gray-500 mt-1">{{ t('workspace.membersDesc') }}</p>
      </div>
      <el-button type="primary" @click="openAddDialog">
        {{ t('workspace.addMember') }}
      </el-button>
    </div>

    <el-card shadow="never">
      <el-table :data="members" v-loading="loading" stripe>
        <el-table-column :label="t('workspace.displayName')" min-width="120">
          <template #default="{ row }">
            {{ row.displayName || '-' }}
          </template>
        </el-table-column>
        <el-table-column :label="t('workspace.email')" prop="email" min-width="180" />
        <el-table-column :label="t('workspace.role')" width="100">
          <template #default="{ row }">
            <el-tag
              :type="(roleTagType[row.role] as 'danger' | 'warning' | 'info' | 'success') || 'info'"
              size="small"
            >
              {{ row.role }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column :label="t('workspace.joinedAt')" prop="joinedAt" width="180" />
        <el-table-column :label="t('common.actions')" width="100" fixed="right">
          <template #default="{ row }">
            <el-button
              v-if="row.role !== 'owner'"
              size="small"
              link
              type="danger"
              @click="handleRemove(row)"
            >
              {{ t('common.remove') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="flex justify-end mt-4">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next"
          @current-change="onPageChange"
          @size-change="onPageSizeChange"
        />
      </div>
    </el-card>

    <!-- Add Member Dialog -->
    <el-dialog
      v-model="addDialogVisible"
      :title="t('workspace.addMember')"
      width="480px"
    >
      <el-form
        ref="addFormRef"
        :model="addForm"
        :rules="addRules"
        label-position="top"
      >
        <el-form-item :label="t('workspace.accountId')" prop="accountId">
          <el-input
            v-model="addForm.accountId"
            :placeholder="t('workspace.accountIdPlaceholder')"
          />
        </el-form-item>
        <el-form-item :label="t('workspace.role')">
          <el-select v-model="addForm.role" class="w-full">
            <el-option
              v-for="opt in roleOptions"
              :key="opt.value"
              :label="opt.label"
              :value="opt.value"
            />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="addDialogVisible = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="addSubmitting" @click="submitAdd">
          {{ t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>
