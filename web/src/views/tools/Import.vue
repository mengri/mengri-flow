<template>
  <div class="import-tools">
    <div class="header">
      <h1>批量导入工具</h1>
    </div>

    <el-steps :active="step" class="steps">
      <el-step title="选择导入方式" />
      <el-step title="上传文件" />
      <el-step title="预览确认" />
    </el-steps>

    <!-- 步骤1：选择导入方式 -->
    <div v-if="step === 1" class="step-content">
      <el-radio-group v-model="importType">
        <el-radio label="openapi">OpenAPI/Swagger</el-radio>
        <el-radio label="proto">Proto文件</el-radio>
        <el-radio label="sqlc">SQLc文件</el-radio>
      </el-radio-group>

      <div class="actions">
        <el-button type="primary" @click="nextStep" :disabled="!importType">
          下一步
        </el-button>
      </div>
    </div>

    <!-- 步骤2：上传文件 -->
    <div v-if="step === 2" class="step-content">
      <el-upload
        ref="uploadRef"
        drag
        :action="uploadUrl"
        :headers="uploadHeaders"
        :data="uploadData"
        :before-upload="beforeUpload"
        :on-success="handleUploadSuccess"
        :on-error="handleUploadError"
        accept=".json,.yaml,.yml,.proto,.sql"
      >
        <el-icon class="el-icon--upload"><upload-filled /></el-icon>
        <div class="el-upload__text">
          拖拽文件到此处或 <em>点击上传</em>
        </div>
        <template #tip>
          <div class="el-upload__tip">
            支持 OpenAPI/Swagger (.json/.yaml), Proto (.proto), SQLc (.sql) 格式
          </div>
        </template>
      </el-upload>

      <div class="actions">
        <el-button @click="prevStep">上一步</el-button>
      </div>
    </div>

    <!-- 步骤3：预览确认 -->
    <div v-if="step === 3" class="step-content">
      <div class="summary">
        共发现 <strong>{{ previewTools.length }}</strong> 个工具
      </div>

      <el-table :data="previewTools" max-height="400">
        <el-table-column prop="name" label="名称" />
        <el-table-column prop="type" label="类型" />
        <el-table-column prop="method" label="方法" />
        <el-table-column prop="path" label="路径" show-overflow-tooltip />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag v-if="row.exists" type="info">已存在</el-tag>
            <el-tag v-else type="success">新发现</el-tag>
          </template>
        </el-table-column>
      </el-table>

      <div class="actions">
        <el-button @click="prevStep">上一步</el-button>
        <el-button type="primary" @click="confirmImport" :loading="importing">
          确认导入
        </el-button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { toolAPI } from '@/api/tools'
import type { Tool } from '@/types/tool'
import { useWorkspaceRoute } from '@/composables/useWorkspaceRoute'

const router = useRouter()
const { toolsPath } = useWorkspaceRoute()

const step = ref(1)
const importType = ref('')
const previewTools = ref<Tool[]>([])
const importing = ref(false)

const uploadUrl = '/api/v1/tools/import-preview'
const uploadHeaders = {
  Authorization: `Bearer ${localStorage.getItem('token')}`,
}
const uploadData = {
  importType,
}

function nextStep() {
  step.value++
}

function prevStep() {
  step.value--
}

function beforeUpload(file: File) {
  const isValidType = ['application/json', 'application/x-yaml', 'text/plain'].includes(file.type)
  const isValidSize = file.size / 1024 / 1024 < 10 // 10MB

  if (!isValidType) {
    ElMessage.error('不支持的文件类型')
    return false
  }
  if (!isValidSize) {
    ElMessage.error('文件大小不能超过 10MB')
    return false
  }
  return true
}

function handleUploadSuccess(response: any) {
  previewTools.value = response.data || []
  step.value = 3
}

function handleUploadError() {
  ElMessage.error('上传失败')
}

async function confirmImport() {
  importing.value = true
  try {
    const formData = new FormData()
    formData.append('importType', importType.value)
    // 这里需要实际的文件数据，根据upload组件的返回值处理
    
    const tools = await toolAPI.import(formData)
    ElMessage.success(`成功导入 ${tools.length} 个工具`)
    router.push(toolsPath())
  } catch (error) {
    ElMessage.error('导入失败')
  } finally {
    importing.value = false
  }
}
</script>

<style scoped>
.import-tools {
  padding: 20px;
}

.header {
  margin-bottom: 30px;
}

.steps {
  margin-bottom: 40px;
}

.step-content {
  min-height: 400px;
  padding: 40px;
  background: #f5f7fa;
  border-radius: 4px;
}

.actions {
  margin-top: 40px;
  text-align: center;
}

.summary {
  margin-bottom: 20px;
  padding: 20px;
  background: white;
  border-radius: 4px;
  text-align: center;
}
</style>
