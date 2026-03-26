<script setup lang="ts">
import { reactive, ref } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import { useAuth } from '@/composables/useAuth'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const { handleLogin } = useAuth()

const formRef = ref<FormInstance>()
const loading = ref(false)

const form = reactive({
  account: '',
  password: '',
})

const rules: FormRules = {
  account: [
    { required: true, message: 'Please enter email or username', trigger: 'blur' },
  ],
  password: [
    { required: true, message: 'Please enter password', trigger: 'blur' },
  ],
}

async function onSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  loading.value = true
  try {
    await handleLogin(form.account, form.password)
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-gray-100">
    <div class="w-full max-w-md">
      <div class="bg-white rounded-lg shadow-lg p-8">
        <div class="text-center mb-8">
          <h1 class="text-2xl font-bold text-gray-800">Mengri Flow</h1>
          <p class="text-gray-500 mt-2">Sign in to your account</p>
        </div>

        <el-form
          ref="formRef"
          :model="form"
          :rules="rules"
          label-position="top"
          size="large"
          @submit.prevent="onSubmit"
        >
          <el-form-item label="Email / Username" prop="account">
            <el-input
              v-model="form.account"
              placeholder="Enter email or username"
              :prefix-icon="UserIcon"
              autofocus
            />
          </el-form-item>

          <el-form-item label="Password" prop="password">
            <el-input
              v-model="form.password"
              type="password"
              placeholder="Enter password"
              :prefix-icon="LockIcon"
              show-password
              @keyup.enter="onSubmit"
            />
          </el-form-item>

          <el-form-item>
            <el-button
              type="primary"
              class="w-full"
              :loading="loading || authStore.loading"
              @click="onSubmit"
            >
              Sign In
            </el-button>
          </el-form-item>
        </el-form>

        <div class="text-center text-sm text-gray-400 mt-4">
          No account? Contact your administrator.
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { User as UserIcon, Lock as LockIcon } from '@element-plus/icons-vue'

export default {
  components: { UserIcon, LockIcon },
}
</script>
