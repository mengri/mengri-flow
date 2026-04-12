<template>
  <div class="create-flow">
    <div class="header">
      <h1>创建流程</h1>
    </div>

    <el-card>
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="120px"
        @submit.prevent="handleSubmit"
      >
        <el-form-item label="流程名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入流程名称" />
        </el-form-item>

        <el-form-item label="描述" prop="description">
          <el-input
            v-model="form.description"
            type="textarea"
            :rows="3"
            placeholder="请输入流程描述"
          />
        </el-form-item>

        <el-form-item label="工作空间" prop="workspaceId">
          <el-select v-model="form.workspaceId" placeholder="请选择工作空间" style="width: auto; min-width: 180px">
            <el-option
              v-for="ws in workspaces"
              :key="ws.id"
              :label="ws.name"
              :value="ws.id"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="输入Schema" prop="inputSchema">
          <el-input
            v-model="inputSchemaStr"
            type="textarea"
            :rows="5"
            placeholder='{"type":"object","properties":{}}'
            @input="handleSchemaChange"
          />
        </el-form-item>

        <el-form-item label="输出Schema" prop="outputSchema">
          <el-input
            v-model="outputSchemaStr"
            type="textarea"
            :rows="5"
            placeholder='{"type":"object","properties":{}}'
            @input="handleOutputSchemaChange"
          />
        </el-form-item>

        <el-form-item>
          <el-button type="primary" @click="handleSubmit" :loading="submitting">
            创建
          </el-button>
          <el-button @click="handleCancel">取消</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { flowAPI } from '@/api/flows'
import { useWorkspaceStore } from '@/stores/workspace'
import { useWorkspaceRoute } from '@/composables/useWorkspaceRoute'
import type { CreateFlowRequest } from '@/types/flow'

const router = useRouter()
const workspaceStore = useWorkspaceStore()
const { flowDetailPath, flowsPath } = useWorkspaceRoute()

const formRef = ref()

const form = reactive<CreateFlowRequest & { inputSchema?: any; outputSchema?: any }>({
  name: '',
  description: '',
  workspaceId: workspaceStore.currentWorkspaceIdOrThrow,
})

const rules = {
  name: [{ required: true, message: '请输入流程名称', trigger: 'blur' }],
  workspaceId: [{ required: true, message: '请选择工作空间', trigger: 'change' }],
}

const submitting = ref(false)
const inputSchemaStr = ref('{"type":"object","properties":{}}')
const outputSchemaStr = ref('{"type":"object","properties":{}}')

const workspaces = computed(() => workspaceStore.workspaces)

function handleSchemaChange() {
  try {
    form.inputSchema = JSON.parse(inputSchemaStr.value)
  } catch (error) {
    // 解析失败，不更新
  }
}

function handleOutputSchemaChange() {
  try {
    form.outputSchema = JSON.parse(outputSchemaStr.value)
  } catch (error) {
    // 解析失败，不更新
  }
}

async function handleSubmit() {
  try {
    await formRef.value?.validate()
    submitting.value = true
    
    // 确保Schema是有效的JSON
    handleSchemaChange()
    handleOutputSchemaChange()
    
    const response = await flowAPI.create(form)
    ElMessage.success('创建成功')
    router.push(flowDetailPath(response.id))
  } catch (error) {
    ElMessage.error('创建失败')
  } finally {
    submitting.value = false
  }
}

function handleCancel() {
  router.push(flowsPath())
}
</script>

<style scoped>
.create-flow {
  padding: 20px;
}

.header {
  margin-bottom: 20px;
}
</style>
