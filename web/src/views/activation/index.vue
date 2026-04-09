<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import type { FormInstance, FormRules } from 'element-plus'
import { ElMessage } from 'element-plus'
import { validateActivation, confirmActivation } from '@/api/auth'
import type { ActivationValidateResponse } from '@/types'

const route = useRoute()
const router = useRouter()

const state = ref<'loading' | 'valid' | 'invalid' | 'expired' | 'already_activated'>('loading')
const activationInfo = ref<ActivationValidateResponse | null>(null)
const formRef = ref<FormInstance>()
const submitting = ref(false)

const form = reactive({
  password: '',
  confirmPassword: '',
})

const rules: FormRules = {
  password: [
    { required: true, message: 'Please set a password', trigger: 'blur' },
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
    { required: true, message: 'Please confirm your password', trigger: 'blur' },
    {
      validator: (_rule, value: string, callback) => {
        if (value !== form.password) {
          callback(new Error('Passwords do not match'))
        } else {
          callback()
        }
      },
      trigger: 'blur',
    },
  ],
}

const token = (route.query.token as string) || ''

onMounted(async () => {
  if (!token) {
    state.value = 'invalid'
    return
  }
  try {
    const result = await validateActivation(token)
    activationInfo.value = result
    if (result.alreadyActivated) {
      state.value = 'already_activated'
    } else if (result.valid) {
      state.value = 'valid'
    } else {
      state.value = 'expired'
    }
  } catch {
    state.value = 'invalid'
  }
})

async function onSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  submitting.value = true
  try {
    await confirmActivation({
      token,
      password: form.password,
      confirmPassword: form.confirmPassword,
    })
    ElMessage.success('Account activated! Please sign in.')
    await router.push('/login')
  } catch {
    // error handled by interceptor
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-gray-100">
    <div class="w-full max-w-md">
      <div class="bg-white rounded-lg shadow-lg p-8">
        <div class="text-center mb-6">
          <h1 class="text-2xl font-bold text-gray-800">Activate Account</h1>
        </div>

        <!-- Loading -->
        <div v-if="state === 'loading'" class="text-center py-8">
          <el-icon class="is-loading text-4xl text-blue-500"><Loading /></el-icon>
          <p class="text-gray-500 mt-4">Validating activation link...</p>
        </div>

        <!-- Valid: Show password form -->
        <template v-else-if="state === 'valid'">
          <p class="text-gray-600 mb-4 text-center">
            Set a password for
            <strong>{{ activationInfo?.emailMasked }}</strong>
          </p>

          <el-form
            ref="formRef"
            :model="form"
            :rules="rules"
            label-position="top"
            size="large"
          >
            <el-form-item label="Password" prop="password">
              <el-input
                v-model="form.password"
                type="password"
                placeholder="Enter password (min 8 characters)"
                show-password
              />
            </el-form-item>

            <el-form-item label="Confirm Password" prop="confirmPassword">
              <el-input
                v-model="form.confirmPassword"
                type="password"
                placeholder="Confirm password"
                show-password
                @keyup.enter="onSubmit"
              />
            </el-form-item>

            <el-form-item>
              <el-button
                type="primary"
                class="w-full"
                :loading="submitting"
                @click="onSubmit"
              >
                Activate Account
              </el-button>
            </el-form-item>
          </el-form>

          <div class="text-xs text-gray-400 mt-2">
            Password must contain uppercase, lowercase, digit, and special character.
          </div>
        </template>

        <!-- Invalid token -->
        <div v-else-if="state === 'invalid'" class="text-center py-8">
          <el-result icon="error" title="Invalid Link">
            <template #sub-title>
              <p>This activation link is invalid. Please contact your administrator to resend.</p>
            </template>
            <template #extra>
              <el-button type="primary" @click="$router.push('/login')">Go to Login</el-button>
            </template>
          </el-result>
        </div>

        <!-- Expired token -->
        <div v-else-if="state === 'expired'" class="text-center py-8">
          <el-result icon="warning" title="Link Expired">
            <template #sub-title>
              <p>This activation link has expired. Please contact your administrator to resend.</p>
            </template>
            <template #extra>
              <el-button type="primary" @click="$router.push('/login')">Go to Login</el-button>
            </template>
          </el-result>
        </div>

        <!-- Already activated -->
        <div v-else-if="state === 'already_activated'" class="text-center py-8">
          <el-result icon="success" title="Already Activated">
            <template #sub-title>
              <p>This account has already been activated. You can sign in directly.</p>
            </template>
            <template #extra>
              <el-button type="primary" @click="$router.push('/login')">Go to Login</el-button>
            </template>
          </el-result>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { Loading } from '@element-plus/icons-vue'

export default {
  components: { Loading },
}
</script>
