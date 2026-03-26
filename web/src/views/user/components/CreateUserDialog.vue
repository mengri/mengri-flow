<script setup lang="ts">
import { reactive, ref } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import { createUser } from '@/api/user'
import { ElMessage } from 'element-plus'
import type { CreateUserRequest } from '@/types'

defineProps<{
  visible: boolean
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  'created': []
}>()

const formRef = ref<FormInstance>()
const submitting = ref(false)

const form = reactive<CreateUserRequest>({
  username: '',
  email: '',
  password: '',
})

const rules: FormRules = {
  username: [
    { required: true, message: 'Please enter username', trigger: 'blur' },
    { min: 2, max: 50, message: '2-50 characters', trigger: 'blur' },
  ],
  email: [
    { required: true, message: 'Please enter email', trigger: 'blur' },
    { type: 'email', message: 'Invalid email', trigger: 'blur' },
  ],
  password: [
    { required: true, message: 'Please enter password', trigger: 'blur' },
    { min: 8, message: 'At least 8 characters', trigger: 'blur' },
  ],
}

async function handleSubmit() {
  if (!formRef.value) return
  await formRef.value.validate()

  submitting.value = true
  try {
    await createUser(form)
    ElMessage.success('User created')
    emit('created')
    resetForm()
  } finally {
    submitting.value = false
  }
}

function resetForm() {
  form.username = ''
  form.email = ''
  form.password = ''
  formRef.value?.resetFields()
}

function handleClose() {
  resetForm()
  emit('update:visible', false)
}
</script>

<template>
  <el-dialog
    title="Create User"
    :model-value="visible"
    width="480px"
    @close="handleClose"
  >
    <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
      <el-form-item label="Username" prop="username">
        <el-input v-model="form.username" placeholder="Enter username" />
      </el-form-item>
      <el-form-item label="Email" prop="email">
        <el-input v-model="form.email" placeholder="Enter email" />
      </el-form-item>
      <el-form-item label="Password" prop="password">
        <el-input v-model="form.password" type="password" placeholder="Enter password" show-password />
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button @click="handleClose">Cancel</el-button>
      <el-button type="primary" :loading="submitting" @click="handleSubmit">
        Create
      </el-button>
    </template>
  </el-dialog>
</template>
