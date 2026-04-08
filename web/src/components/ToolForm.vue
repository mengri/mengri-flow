<template>
  <el-form
    ref="formRef"
    :model="form"
    :rules="rules"
    label-width="120px"
    @submit.prevent="handleSubmit"
  >
    <el-form-item label="工具名称" prop="name">
      <el-input v-model="form.name" placeholder="请输入工具名称" />
    </el-form-item>

    <el-form-item label="所属资源" prop="resourceId">
      <el-select v-model="form.resourceId" placeholder="请选择资源">
        <el-option
          v-for="resource in resources"
          :key="resource.id"
          :label="resource.name"
          :value="resource.id"
        />
      </el-select>
    </el-form-item>

    <el-form-item label="工具类型" prop="type">
      <el-input v-model="form.type" placeholder="例如：restful, grpc, sql" />
    </el-form-item>

    <el-form-item label="HTTP方法" prop="method" v-if="form.type === 'restful'">
      <el-select v-model="form.method" placeholder="请选择HTTP方法">
        <el-option label="GET" value="GET" />
        <el-option label="POST" value="POST" />
        <el-option label="PUT" value="PUT" />
        <el-option label="DELETE" value="DELETE" />
        <el-option label="PATCH" value="PATCH" />
      </el-select>
    </el-form-item>

    <el-form-item label="路径" prop="path" v-if="form.type === 'restful'">
      <el-input v-model="form.path" placeholder="例如：/api/users/{id}" />
    </el-form-item>

    <el-form-item label="输入Schema" prop="inputSchema">
      <el-input
        v-model="inputSchemaStr"
        type="textarea"
        :rows="5"
        placeholder='{"type":"object","properties":{}}'
        @input="handleSchemaChange('input')"
      />
    </el-form-item>

    <el-form-item label="输出Schema" prop="outputSchema">
      <el-input
        v-model="outputSchemaStr"
        type="textarea"
        :rows="5"
        placeholder='{"type":"object","properties":{}}'
        @input="handleSchemaChange('output')"
      />
    </el-form-item>

    <el-form-item label="描述" prop="description">
      <el-input
        v-model="form.description"
        type="textarea"
        :rows="3"
        placeholder="请输入工具描述"
      />
    </el-form-item>

    <el-form-item label="标签" prop="tags">
      <el-select
        v-model="form.tags"
        multiple
        filterable
        allow-create
        placeholder="输入标签后按回车"
      />
    </el-form-item>

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

    <el-form-item>
      <el-button type="primary" @click="handleSubmit" :loading="submitting">
        保存
      </el-button>
      <el-button @click="handleCancel">取消</el-button>
    </el-form-item>
  </el-form>
</template>

<script setup lang="ts">
import { ref, reactive, defineProps, defineEmits, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { toolAPI } from '@/api/tools'
import { resourceAPI } from '@/api/resources'
import { useWorkspaceStore } from '@/stores/workspace'
import type { Tool, CreateToolRequest } from '@/types/tool'
import type { Resource } from '@/types/resource'

const props = defineProps<{
  initialData?: Partial<CreateToolRequest>
}>()

const emit = defineEmits<{
  (e: 'success', data: Tool): void
  (e: 'cancel'): void
}>()

const formRef = ref()
const workspaceStore = useWorkspaceStore()

const form = reactive<CreateToolRequest>({
  name: props.initialData?.name || '',
  resourceId: props.initialData?.resourceId || '',
  type: props.initialData?.type || 'restful',
  method: props.initialData?.method || 'GET',
  path: props.initialData?.path || '',
  inputSchema: props.initialData?.inputSchema || {},
  outputSchema: props.initialData?.outputSchema || {},
  description: props.initialData?.description || '',
  tags: props.initialData?.tags || [],
  workspaceId: props.initialData?.workspaceId || workspaceStore.currentWorkspace,
})

const rules = {
  name: [{ required: true, message: '请输入工具名称', trigger: 'blur' }],
  resourceId: [{ required: true, message: '请选择资源', trigger: 'change' }],
  type: [{ required: true, message: '请输入工具类型', trigger: 'blur' }],
}

const submitting = ref(false)
const resources = ref<Resource[]>([])
const inputSchemaStr = ref(JSON.stringify(form.inputSchema, null, 2))
const outputSchemaStr = ref(JSON.stringify(form.outputSchema, null, 2))

async function loadResources() {
  try {
    const data = await resourceAPI.list({
      workspaceId: workspaceStore.currentWorkspace,
    })
    resources.value = data
  } catch (error) {
    ElMessage.error('加载资源列表失败')
  }
}

function handleSchemaChange(type: 'input' | 'output') {
  try {
    if (type === 'input') {
      form.inputSchema = JSON.parse(inputSchemaStr.value)
    } else {
      form.outputSchema = JSON.parse(outputSchemaStr.value)
    }
  } catch (error) {
    // 解析失败，不更新
  }
}

async function handleSubmit() {
  try {
    await formRef.value?.validate()
    submitting.value = true
    
    // 确保Schema是有效的JSON
    handleSchemaChange('input')
    handleSchemaChange('output')
    
    const tool = await toolAPI.create(form)
    ElMessage.success('创建成功')
    emit('success', tool)
  } catch (error) {
    ElMessage.error('创建失败')
  } finally {
    submitting.value = false
  }
}

function handleCancel() {
  emit('cancel')
}

onMounted(() => {
  loadResources()
  
  // 如果有初始数据，更新Schema字符串
  if (props.initialData?.inputSchema) {
    inputSchemaStr.value = JSON.stringify(props.initialData.inputSchema, null, 2)
  }
  if (props.initialData?.outputSchema) {
    outputSchemaStr.value = JSON.stringify(props.initialData.outputSchema, null, 2)
  }
})
</script>
