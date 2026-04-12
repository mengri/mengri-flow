<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import { useAccountStore } from '@/stores/account'
import { useAccountAdmin } from '@/composables/useAccount'
import type { AccountStatus, AccountResponse, AuditEventFilter } from '@/types'

const store = useAccountStore()
const {
  createDialogVisible,
  detailDialogVisible,
  creating,
  handleCreate,
  handleStatusChange,
  handleResend,
  handleViewDetail,
} = useAccountAdmin()

// --- Filters ---
const statusFilter = ref<AccountStatus | ''>('')
const keyword = ref('')

function onSearch() {
  store.fetchAccounts({ page: 1, status: statusFilter.value, keyword: keyword.value })
}

function onPageChange(page: number) {
  store.fetchAccounts({ page })
}

function onPageSizeChange(pageSize: number) {
  store.fetchAccounts({ page: 1, pageSize })
}

// --- Create Form ---
const createFormRef = ref<FormInstance>()
const createForm = reactive({
  email: '',
  displayName: '',
  username: '',
})

const createRules: FormRules = {
  email: [
    { required: true, message: 'Please enter email', trigger: 'blur' },
    { type: 'email', message: 'Invalid email format', trigger: 'blur' },
  ],
  displayName: [
    { required: true, message: 'Please enter display name', trigger: 'blur' },
    { max: 50, message: 'Max 50 characters', trigger: 'blur' },
  ],
  username: [
    { required: true, message: 'Please enter username', trigger: 'blur' },
    { min: 2, max: 50, message: '2-50 characters', trigger: 'blur' },
  ],
}

async function submitCreate() {
  const valid = await createFormRef.value?.validate().catch(() => false)
  if (!valid) return
  const success = await handleCreate({ ...createForm })
  if (success) {
    createFormRef.value?.resetFields()
  }
}

function onCreateDialogClose() {
  createFormRef.value?.resetFields()
}

// --- Audit Events ---
const auditDialogVisible = ref(false)
const auditAccountId = ref('')

const auditFilters = reactive<AuditEventFilter>({
  accountId: '',
  eventType: '',
  page: 1,
  pageSize: 20,
})

function openAuditDialog(accountId?: string) {
  auditFilters.accountId = accountId || ''
  auditFilters.eventType = ''
  auditFilters.page = 1
  auditAccountId.value = accountId || ''
  auditDialogVisible.value = true
  store.fetchAuditEvents(auditFilters)
}

function onAuditPageChange(page: number) {
  auditFilters.page = page
  store.fetchAuditEvents(auditFilters)
}

// --- Status helpers ---
const statusTagType: Record<string, string> = {
  ACTIVE: 'success',
  PENDING_ACTIVATION: 'warning',
  LOCKED: 'danger',
  DISABLED: 'info',
}

const statusOptions: { label: string; value: AccountStatus | '' }[] = [
  { label: 'All', value: '' },
  { label: 'Pending Activation', value: 'PENDING_ACTIVATION' },
  { label: 'Active', value: 'ACTIVE' },
  { label: 'Locked', value: 'LOCKED' },
  { label: 'Disabled', value: 'DISABLED' },
]

/** 根据当前状态返回可用的操作 */
function getAvailableActions(account: AccountResponse) {
  const actions: { label: string; action: 'lock' | 'unlock' | 'disable' | 'enable'; type: string }[] = []
  switch (account.status) {
    case 'ACTIVE':
      actions.push({ label: 'Lock', action: 'lock', type: 'warning' })
      actions.push({ label: 'Disable', action: 'disable', type: 'danger' })
      break
    case 'LOCKED':
      actions.push({ label: 'Unlock', action: 'unlock', type: 'success' })
      break
    case 'DISABLED':
      actions.push({ label: 'Enable', action: 'enable', type: 'success' })
      break
  }
  return actions
}

// --- Init ---
onMounted(() => {
  store.fetchAccounts()
})
</script>

<template>
  <div class="p-6">
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold text-gray-800">Account Management</h1>
      <div class="flex gap-3">
        <el-button type="info" @click="openAuditDialog()">
          Audit Log
        </el-button>
        <el-button type="primary" @click="createDialogVisible = true">
          Create Account
        </el-button>
      </div>
    </div>

    <!-- Filters -->
    <el-card shadow="never" class="mb-4">
      <div class="flex items-center gap-4">
        <el-input
          v-model="keyword"
          placeholder="Search by name, email, username..."
          clearable
          class="w-64"
          @keyup.enter="onSearch"
          @clear="onSearch"
        />
        <el-select v-model="statusFilter" placeholder="Status" clearable style="width: auto; min-width: 180px" @change="onSearch">
          <el-option
            v-for="opt in statusOptions"
            :key="opt.value"
            :label="opt.label"
            :value="opt.value"
          />
        </el-select>
        <el-button type="primary" @click="onSearch">Search</el-button>
      </div>
    </el-card>

    <!-- Account Table -->
    <el-card shadow="never">
      <el-table :data="store.accounts" v-loading="store.loading" stripe>
        <el-table-column label="Display Name" prop="displayName" min-width="120" />
        <el-table-column label="Username" prop="username" width="120" />
        <el-table-column label="Email" prop="email" min-width="180" />
        <el-table-column label="Role" prop="role" width="80">
          <template #default="{ row }">
            <el-tag :type="row.role === 'admin' ? 'danger' : 'info'" size="small">
              {{ row.role }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="Status" prop="status" width="160">
          <template #default="{ row }">
            <el-tag
              :type="(statusTagType[row.status] as 'success' | 'warning' | 'danger' | 'info') || 'info'"
              size="small"
            >
              {{ row.status }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="Created" prop="createdAt" width="180" />
        <el-table-column label="Actions" width="260" fixed="right">
          <template #default="{ row }">
            <el-button size="small" link type="primary" @click="handleViewDetail(row.accountId)">
              Detail
            </el-button>
            <el-button
              v-if="row.status === 'PENDING_ACTIVATION'"
              size="small"
              link
              type="warning"
              @click="handleResend(row.accountId)"
            >
              Resend
            </el-button>
            <template v-for="act in getAvailableActions(row)" :key="act.action">
              <el-button
                size="small"
                link
                :type="act.type as 'primary' | 'success' | 'warning' | 'danger' | 'info'"
                @click="handleStatusChange(row, act.action)"
              >
                {{ act.label }}
              </el-button>
            </template>
            <el-button size="small" link type="info" @click="openAuditDialog(row.accountId)">
              Audit
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="flex justify-end mt-4">
        <el-pagination
          v-model:current-page="store.filters.page"
          v-model:page-size="store.filters.pageSize"
          :total="store.total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next"
          @current-change="onPageChange"
          @size-change="onPageSizeChange"
        />
      </div>
    </el-card>

    <!-- Create Account Dialog -->
    <el-dialog
      v-model="createDialogVisible"
      title="Create Account"
      width="480px"
      @close="onCreateDialogClose"
    >
      <el-form
        ref="createFormRef"
        :model="createForm"
        :rules="createRules"
        label-position="top"
      >
        <el-form-item label="Email" prop="email">
          <el-input v-model="createForm.email" placeholder="user@example.com" />
        </el-form-item>
        <el-form-item label="Display Name" prop="displayName">
          <el-input v-model="createForm.displayName" placeholder="Display name" />
        </el-form-item>
        <el-form-item label="Username" prop="username">
          <el-input v-model="createForm.username" placeholder="Username (2-50 chars)" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createDialogVisible = false">Cancel</el-button>
        <el-button type="primary" :loading="creating" @click="submitCreate">
          Create
        </el-button>
      </template>
    </el-dialog>

    <!-- Account Detail Dialog -->
    <el-dialog
      v-model="detailDialogVisible"
      title="Account Detail"
      width="600px"
    >
      <template v-if="store.currentDetail">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="Account ID" :span="2">
            {{ store.currentDetail.accountId }}
          </el-descriptions-item>
          <el-descriptions-item label="Display Name">
            {{ store.currentDetail.displayName }}
          </el-descriptions-item>
          <el-descriptions-item label="Username">
            {{ store.currentDetail.username }}
          </el-descriptions-item>
          <el-descriptions-item label="Email" :span="2">
            {{ store.currentDetail.email }}
          </el-descriptions-item>
          <el-descriptions-item label="Role">
            <el-tag :type="store.currentDetail.role === 'admin' ? 'danger' : 'info'" size="small">
              {{ store.currentDetail.role }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="Status">
            <el-tag
              :type="(statusTagType[store.currentDetail.status] as 'success' | 'warning' | 'danger' | 'info') || 'info'"
              size="small"
            >
              {{ store.currentDetail.status }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="Created At">
            {{ store.currentDetail.createdAt }}
          </el-descriptions-item>
          <el-descriptions-item label="Activated At">
            {{ store.currentDetail.activatedAt || '-' }}
          </el-descriptions-item>
        </el-descriptions>

        <h3 class="mt-4 mb-2 font-semibold text-gray-700">Login Methods</h3>
        <el-table :data="store.currentDetail.identities" stripe size="small">
          <el-table-column label="Type" prop="loginType" width="120" />
          <el-table-column label="Identifier" prop="maskedIdentifier">
            <template #default="{ row }">
              {{ row.maskedIdentifier || '-' }}
            </template>
          </el-table-column>
          <el-table-column label="Bound At" prop="boundAt" width="180" />
        </el-table>
        <div
          v-if="store.currentDetail.identities.length === 0"
          class="text-center py-3 text-gray-400 text-sm"
        >
          No login methods bound
        </div>
      </template>
    </el-dialog>

    <!-- Audit Events Dialog -->
    <el-dialog
      v-model="auditDialogVisible"
      title="Audit Events"
      width="800px"
    >
      <el-table :data="store.auditEvents" v-loading="store.auditLoading" stripe size="small">
        <el-table-column label="Event Type" prop="eventType" width="180" />
        <el-table-column label="Result" prop="result" width="100">
          <template #default="{ row }">
            <el-tag
              :type="row.result === 'success' ? 'success' : 'danger'"
              size="small"
            >
              {{ row.result }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="IP" prop="ip" width="140" />
        <el-table-column label="User Agent" prop="ua" show-overflow-tooltip />
        <el-table-column label="Time" prop="createdAt" width="180" />
      </el-table>

      <div class="flex justify-end mt-4">
        <el-pagination
          v-model:current-page="auditFilters.page"
          :total="store.auditTotal"
          :page-size="auditFilters.pageSize"
          layout="total, prev, pager, next"
          @current-change="onAuditPageChange"
        />
      </div>
    </el-dialog>
  </div>
</template>
