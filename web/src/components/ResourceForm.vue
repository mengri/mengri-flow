<template>
  <el-form
    ref="formRef"
    :model="form"
    :rules="rules"
    label-width="120px"
    @submit.prevent="handleSubmit"
  >
    <el-form-item label="资源名称" prop="name">
      <el-input v-model="form.name" placeholder="请输入资源名称" />
    </el-form-item>

    <el-form-item label="资源类型" prop="type">
      <el-select
        v-model="form.type"
        placeholder="请选择资源类型"
        @change="handleTypeChange"
      >
        <el-option label="HTTP服务" value="http" />
        <el-option label="gRPC服务" value="grpc" />
        <el-option label="MySQL数据库" value="mysql" />
        <el-option label="PostgreSQL数据库" value="postgres" />
      </el-select>
    </el-form-item>

    <!-- 动态配置表单 -->
    <component
      :is="getConfigFormComponent(form.type)"
      v-model="form.config"
      :schema="getConfigSchema(form.type)"
    />

    <el-form-item label="工作空间" prop="workspaceId">
      <el-select v-model="form.workspaceId" placeholder="请选择工作空间">
        <el-option
          v-for="ws in workspaces"
          :key="ws.id"
          :label="ws.name"
          :value="ws.id"
        />
      </el-select>
    </el-form-item>

    <el-form-item label="描述" prop="description">
      <el-input
        v-model="form.description"
        type="textarea"
        :rows="3"
        placeholder="请输入资源描述"
      />
    </el-form-item>

    <el-form-item>
      <el-button @click="handleTest" :loading="testing">测试连接</el-button>
      <el-button type="primary" @click="handleSubmit" :loading="submitting">
        保存
      </el-button>
      <el-button @click="handleCancel">取消</el-button>
    </el-form-item>
  </el-form>
</template>

<script setup lang="ts">
import { ref, reactive, defineProps, defineEmits } from 'vue'
import { ElMessage } from 'element-plus'
import { resourceAPI } from '@/api/resources'
import { useWorkspaceStore } from '@/stores/workspace'
import type { Resource, CreateResourceRequest } from '@/types/resource'

const props = defineProps<{
  initialData?: Partial<CreateResourceRequest>
}>()

const emit = defineEmits<{
  (e: 'success', data: Resource): void
  (e: 'cancel'): void
}>()

const formRef = ref()
const workspaceStore = useWorkspaceStore()

const form = reactive<CreateResourceRequest>({
  name: props.initialData?.name || '',
  type: props.initialData?.type || 'http',
  config: props.initialData?.config || {},
  workspaceId: props.initialData?.workspaceId || workspaceStore.currentWorkspace,
  description: props.initialData?.description || '',
})

const rules = {
  name: [{ required: true, message: '请输入资源名称', trigger: 'blur' }],
  type: [{ required: true, message: '请选择资源类型', trigger: 'change' }],
  workspaceId: [{ required: true, message: '请选择工作空间', trigger: 'change' }],
}

const testing = ref(false)
const submitting = ref(false)

async function handleTest() {
  try {
    await formRef.value?.validate()
    testing.value = true
    await resourceAPI.testConnection({ type: form.type, config: form.config })
    ElMessage.success('连接测试成功')
  } catch (error) {
    ElMessage.error('连接测试失败')
  } finally {
    testing.value = false
  }
}

async function handleSubmit() {
  try {
    await formRef.value?.validate()
    submitting.value = true
    const resource = await resourceAPI.create(form)
    ElMessage.success('创建成功')
    emit('success', resource)
  } catch (error) {
    ElMessage.error('创建失败')
  } finally {
    submitting.value = false
  }
}

function handleCancel() {
  emit('cancel')
}

function handleTypeChange(type: string) {
  // 清空配置
  form.config = {}
}

function getConfigFormComponent(type: string) {
  const components = {
    http: 'HttpConfigForm',
    grpc: 'GrpcConfigForm',
    mysql: 'MysqlConfigForm',
    postgres: 'PostgresConfigForm',
  }
  return components[type] || 'div'
}

function getConfigSchema(type: string) {
  // 返回对应类型的Schema
  // 可以从后端获取或前端静态定义
  return {}
}
</script>
