<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import { ElMessage } from 'element-plus'
import { getProfile, listMyIdentities, changePassword, getLoginHistory } from '@/api/account'
import type { ProfileResponse, IdentityListResponse, LoginHistoryItem } from '@/types'

const activeTab = ref('profile')

// --- Profile ---
const profile = ref<ProfileResponse | null>(null)
const profileLoading = ref(false)

async function fetchProfile() {
  profileLoading.value = true
  try {
    profile.value = await getProfile()
  } finally {
    profileLoading.value = false
  }
}

// --- Identities ---
const identities = ref<IdentityListResponse | null>(null)
const identitiesLoading = ref(false)

async function fetchIdentities() {
  identitiesLoading.value = true
  try {
    identities.value = await listMyIdentities()
  } finally {
    identitiesLoading.value = false
  }
}

// --- Change Password ---
const passwordFormRef = ref<FormInstance>()
const passwordSubmitting = ref(false)
const passwordForm = reactive({
  oldPassword: '',
  newPassword: '',
  confirmPassword: '',
  revokeOtherSessions: false,
})

const passwordRules: FormRules = {
  oldPassword: [
    { required: true, message: 'Please enter current password', trigger: 'blur' },
  ],
  newPassword: [
    { required: true, message: 'Please enter new password', trigger: 'blur' },
    { min: 8, message: 'Password must be at least 8 characters', trigger: 'blur' },
    {
      validator: (_rule, value: string, callback) => {
        if (!/[A-Z]/.test(value)) {
          callback(new Error('Must contain at least one uppercase letter'))
        } else if (!/[a-z]/.test(value)) {
          callback(new Error('Must contain at least one lowercase letter'))
        } else if (!/\d/.test(value)) {
          callback(new Error('Must contain at least one digit'))
        } else if (!/[!@#$%^&*()_+\-=[\]{};':"\\|,.<>/?]/.test(value)) {
          callback(new Error('Must contain at least one special character'))
        } else {
          callback()
        }
      },
      trigger: 'blur',
    },
  ],
  confirmPassword: [
    { required: true, message: 'Please confirm new password', trigger: 'blur' },
    {
      validator: (_rule, value: string, callback) => {
        if (value !== passwordForm.newPassword) {
          callback(new Error('Passwords do not match'))
        } else {
          callback()
        }
      },
      trigger: 'blur',
    },
  ],
}

async function onChangePassword() {
  const valid = await passwordFormRef.value?.validate().catch(() => false)
  if (!valid) return

  passwordSubmitting.value = true
  try {
    const result = await changePassword({
      oldPassword: passwordForm.oldPassword,
      newPassword: passwordForm.newPassword,
      confirmPassword: passwordForm.confirmPassword,
      revokeOtherSessions: passwordForm.revokeOtherSessions,
    })
    if (result.changed) {
      ElMessage.success(
        result.revokedSessions > 0
          ? `Password changed. ${result.revokedSessions} other session(s) revoked.`
          : 'Password changed successfully',
      )
      passwordFormRef.value?.resetFields()
    }
  } finally {
    passwordSubmitting.value = false
  }
}

// --- Login History ---
const loginHistory = ref<LoginHistoryItem[]>([])
const historyLoading = ref(false)

async function fetchLoginHistory() {
  historyLoading.value = true
  try {
    loginHistory.value = await getLoginHistory()
  } finally {
    historyLoading.value = false
  }
}

// --- Tab Change Handler ---
function onTabChange(tab: string) {
  if (tab === 'profile' && !profile.value) fetchProfile()
  if (tab === 'identities' && !identities.value) fetchIdentities()
  if (tab === 'history') fetchLoginHistory()
}

// --- Init ---
onMounted(() => {
  fetchProfile()
})

const loginTypeLabels: Record<string, string> = {
  password: 'Password',
  sms: 'SMS',
  wechat_qr: 'WeChat',
  lark_qr: 'Lark',
  github_oauth: 'GitHub',
}

const statusTagType: Record<string, string> = {
  ACTIVE: 'success',
  PENDING_ACTIVATION: 'warning',
  LOCKED: 'danger',
  DISABLED: 'info',
}
</script>

<template>
  <div class="max-w-4xl mx-auto py-8 px-4">
    <div class="mb-6">
      <h1 class="text-2xl font-bold text-gray-800">Account Center</h1>
      <p class="text-gray-500 mt-1">Manage your account settings and security</p>
    </div>

    <el-tabs v-model="activeTab" @tab-change="onTabChange">
      <!-- Profile Tab -->
      <el-tab-pane label="Profile" name="profile">
        <el-card v-loading="profileLoading" shadow="never">
          <template v-if="profile">
            <el-descriptions :column="1" border>
              <el-descriptions-item label="Display Name">
                {{ profile.displayName }}
              </el-descriptions-item>
              <el-descriptions-item label="Username">
                {{ profile.username }}
              </el-descriptions-item>
              <el-descriptions-item label="Email">
                {{ profile.email }}
              </el-descriptions-item>
              <el-descriptions-item label="Role">
                <el-tag :type="profile.role === 'admin' ? 'danger' : 'info'" size="small">
                  {{ profile.role }}
                </el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="Status">
                <el-tag
                  :type="(statusTagType[profile.accountStatus] as 'success' | 'warning' | 'danger' | 'info') || 'info'"
                  size="small"
                >
                  {{ profile.accountStatus }}
                </el-tag>
              </el-descriptions-item>
            </el-descriptions>
          </template>
        </el-card>
      </el-tab-pane>

      <!-- Change Password Tab -->
      <el-tab-pane label="Change Password" name="password">
        <el-card shadow="never" class="max-w-lg">
          <el-form
            ref="passwordFormRef"
            :model="passwordForm"
            :rules="passwordRules"
            label-position="top"
          >
            <el-form-item label="Current Password" prop="oldPassword">
              <el-input
                v-model="passwordForm.oldPassword"
                type="password"
                placeholder="Enter current password"
                show-password
              />
            </el-form-item>

            <el-form-item label="New Password" prop="newPassword">
              <el-input
                v-model="passwordForm.newPassword"
                type="password"
                placeholder="Enter new password (min 8 characters)"
                show-password
              />
            </el-form-item>

            <el-form-item label="Confirm New Password" prop="confirmPassword">
              <el-input
                v-model="passwordForm.confirmPassword"
                type="password"
                placeholder="Confirm new password"
                show-password
              />
            </el-form-item>

            <el-form-item>
              <el-checkbox v-model="passwordForm.revokeOtherSessions">
                Revoke all other sessions after password change
              </el-checkbox>
            </el-form-item>

            <el-form-item>
              <el-button
                type="primary"
                :loading="passwordSubmitting"
                @click="onChangePassword"
              >
                Change Password
              </el-button>
            </el-form-item>
          </el-form>

          <div class="text-xs text-gray-400 mt-2">
            Password must contain uppercase, lowercase, digit, and special character.
          </div>
        </el-card>
      </el-tab-pane>

      <!-- Identities Tab -->
      <el-tab-pane label="Login Methods" name="identities">
        <el-card v-loading="identitiesLoading" shadow="never">
          <template v-if="identities">
            <el-table :data="identities.identities" stripe>
              <el-table-column label="Method" prop="loginType" width="150">
                <template #default="{ row }">
                  {{ loginTypeLabels[row.loginType] || row.loginType }}
                </template>
              </el-table-column>
              <el-table-column label="Identifier" prop="maskedIdentifier">
                <template #default="{ row }">
                  {{ row.maskedIdentifier || '-' }}
                </template>
              </el-table-column>
              <el-table-column label="Bound At" prop="boundAt" width="200" />
            </el-table>
            <div v-if="identities.identities.length === 0" class="text-center py-4 text-gray-400">
              No login methods bound
            </div>
          </template>
        </el-card>
      </el-tab-pane>

      <!-- Login History Tab -->
      <el-tab-pane label="Login History" name="history">
        <el-card v-loading="historyLoading" shadow="never">
          <el-table :data="loginHistory" stripe>
            <el-table-column label="Event" prop="eventType" width="160" />
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
          <div v-if="loginHistory.length === 0" class="text-center py-4 text-gray-400">
            No login history
          </div>
        </el-card>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>
