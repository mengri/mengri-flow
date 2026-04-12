<template>
  <div class="create-trigger">
    <div class="header">
      <h1>创建触发器</h1>
    </div>

    <el-steps :active="step" class="steps">
      <el-step title="基本信息" />
      <el-step title="绑定流程" />
      <el-step title="输入输出映射" />
      <el-step title="错误处理" />
    </el-steps>

    <div class="step-content">
      <!-- 步骤1：基本信息 -->
      <div v-if="step === 1">
        <el-form ref="formRef" :model="form" :rules="rules" label-width="120px">
          <el-form-item label="名称" prop="name">
            <el-input v-model="form.name" placeholder="请输入触发器名称" />
          </el-form-item>
          <el-form-item label="类型" prop="type">
            <el-radio-group v-model="form.type">
              <el-radio label="restful">RESTful</el-radio>
              <el-radio label="timer">定时任务</el-radio>
              <el-radio label="rabbitmq">RabbitMQ</el-radio>
              <el-radio label="kafka">Kafka</el-radio>
            </el-radio-group>
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
        </el-form>

        <!-- 触发器类型特定配置 -->
        <component
          :is="getTriggerConfigComponent(form.type)"
          v-model="form.config"
          :schema="getTriggerConfigSchema(form.type)"
        />
      </div>

      <!-- 步骤2：绑定流程 -->
      <div v-if="step === 2">
        <el-alert
          title="请选择要绑定的已发布流程"
          type="info"
          :closable="false"
          style="margin-bottom: 20px"
        />

        <el-table :data="publishedFlows" v-loading="loadingFlows">
          <el-table-column prop="name" label="流程名称" />
          <el-table-column prop="description" label="描述" show-overflow-tooltip />
          <el-table-column prop="currentVersion" label="版本" width="100" />
          <el-table-column label="操作" width="100">
            <template #default="{ row }">
              <el-button
                type="primary"
                size="small"
                @click="selectFlow(row)"
              >
                选择
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <!-- 步骤3：输入输出映射 -->
      <div v-if="step === 3">
        <!-- 输入映射 -->
        <section class="mapping-section">
          <h4>输入映射</h4>
          <p class="description">将触发器输入映射到流程输入参数</p>

          <DataMappingEditor
            v-model="form.inputMapping"
            :source-fields="triggerInputFields"
            :target-fields="flowInputFields"
          />
        </section>

        <!-- 输出映射 -->
        <section class="mapping-section">
          <h4>输出映射</h4>
          <p class="description">将流程输出映射到触发器响应</p>

          <DataMappingEditor
            v-model="form.outputMapping"
            :source-fields="flowOutputFields"
            :target-fields="triggerOutputFields"
          />
        </section>
      </div>

      <!-- 步骤4：错误处理 -->
      <div v-if="step === 4">
        <el-form label-width="120px">
          <el-form-item label="错误策略">
            <el-radio-group v-model="form.errorHandling.strategy">
              <el-radio label="default">默认（透传流程错误）</el-radio>
              <el-radio label="custom">自定义错误格式</el-radio>
            </el-radio-group>
          </el-form-item>

          <template v-if="form.errorHandling.strategy === 'custom'">
            <el-form-item label="错误格式">
              <JsonEditor v-model="form.errorHandling.customErrorFormat" />
            </el-form-item>
          </template>

          <el-form-item label="失败重试">
            <el-switch v-model="form.errorHandling.retryOnFailure" />
          </el-form-item>
        </el-form>
      </div>
    </div>

    <div class="actions">
      <el-button v-if="step > 1" @click="prevStep">上一步</el-button>
      <el-button v-if="step < 4" type="primary" @click="nextStep">
        下一步
      </el-button>
      <el-button v-if="step === 4" type="success" @click="handleCreate">
        创建触发器
      </el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { flowAPI } from '@/api/flows'
import { triggerAPI } from '@/api/triggers'
import type { Flow } from '@/types/flow'
import type { CreateTriggerRequest } from '@/types/trigger'
import { useWorkspaceStore } from '@/stores/workspace'
import { useWorkspaceRoute } from '@/composables/useWorkspaceRoute'

const router = useRouter()
const workspaceStore = useWorkspaceStore()
const { triggersPath } = useWorkspaceRoute()

const step = ref(1)
const loadingFlows = ref(false)
const publishedFlows = ref<Flow[]>([])
const workspaces = computed(() => workspaceStore.workspaces)

const rules = {
  name: [{ required: true, message: '请输入触发器名称', trigger: 'blur' }],
  type: [{ required: true, message: '请选择触发器类型', trigger: 'change' }],
  workspaceId: [{ required: true, message: '请选择工作空间', trigger: 'change' }],
}

const form = reactive<CreateTriggerRequest>({
  name: '',
  type: 'restful',
  config: {},
  flowId: '',
  flowVersion: 0,
  clusterId: '',
  inputMapping: {},
  outputMapping: {},
  errorHandling: {
    strategy: 'default',
    retryOnFailure: false,
  },
  workspaceId: workspaceStore.currentWorkspaceIdOrThrow,
})

const triggerInputFields = computed(() => {
  // 根据触发器类型返回输入字段
  switch (form.type) {
    case 'restful':
      return ['headers', 'query', 'body', 'pathParams']
    case 'timer':
      return ['triggerTime', 'executionId', 'triggerId']
    default:
      return []
  }
})

const flowInputFields = computed(() => {
  if (!form.flowId) return []
  const flow = publishedFlows.value.find(f => f.id === form.flowId)
  if (!flow?.inputSchema?.properties) return []
  return Object.keys(flow.inputSchema.properties)
})

const flowOutputFields = computed(() => {
  if (!form.flowId) return []
  const flow = publishedFlows.value.find(f => f.id === form.flowId)
  if (!flow?.outputSchema?.properties) return []
  return Object.keys(flow.outputSchema.properties)
})

const triggerOutputFields = computed(() => {
  // 触发器输出字段通常是固定的
  return ['status', 'data', 'message', 'executionId']
})

function nextStep() {
  step.value++
}

function prevStep() {
  step.value--
}

async function selectFlow(flow: Flow) {
  form.flowId = flow.id
  form.flowVersion = flow.currentVersion
  nextStep()
}

function getTriggerConfigComponent(_type: string) {
  // 返回不同类型的配置组件
  return 'div'
}

function getTriggerConfigSchema(_type: string) {
  // 返回对应类型的Schema
  return {}
}

async function handleCreate() {
  try {
    await triggerAPI.create(form)
    ElMessage.success('创建成功')
    router.push(triggersPath())
  } catch (error) {
    ElMessage.error('创建失败')
  }
}

onMounted(async () => {
  // 加载已发布的流程
  loadingFlows.value = true
  try {
    const result = await flowAPI.list({
      workspaceId: workspaceStore.currentWorkspaceIdOrThrow,
      status: 'active',
    })
    publishedFlows.value = result.list || []
  } finally {
    loadingFlows.value = false
  }
})
</script>

<style scoped>
.create-trigger {
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

.mapping-section {
  margin-bottom: 40px;
}

.mapping-section h4 {
  margin-bottom: 10px;
  color: #303133;
}

.mapping-section .description {
  color: #909399;
  font-size: 14px;
  margin-bottom: 20px;
}

.actions {
  margin-top: 40px;
  text-align: center;
}
</style>
