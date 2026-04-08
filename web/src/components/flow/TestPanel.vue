<template>
  <el-drawer
    v-model="visible"
    title="测试运行"
    :size="700"
    direction="rtl"
    @close="$emit('update:modelValue', false)"
  >
    <div class="test-panel" v-if="flow">
      <!-- 输入参数 -->
      <section class="section">
        <h4>输入参数</h4>
        <el-input
          v-model="inputDataStr"
          type="textarea"
          :rows="8"
          placeholder='请输入JSON格式的输入参数，例如：{"key": "value"}'
        />
        <el-button
          type="primary"
          size="small"
          style="margin-top: 10px"
          @click="handleRun"
          :loading="running"
        >
          运行测试
        </el-button>
      </section>

      <!-- 执行结果 -->
      <section class="section" v-if="result">
        <h4>执行结果</h4>
        <el-descriptions :column="2" border>
          <el-descriptions-item label="状态">
            <el-tag :type="result.success ? 'success' : 'danger'">
              {{ result.success ? '成功' : '失败' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="耗时">
            {{ result.durationMs }}ms
          </el-descriptions-item>
        </el-descriptions>

        <!-- 输出数据 -->
        <div class="output-section">
          <h5>输出数据</h5>
          <pre>{{ JSON.stringify(result.output, null, 2) }}</pre>
        </div>

        <!-- 节点执行日志 -->
        <div class="logs-section">
          <h5>节点执行日志</h5>
          <el-timeline>
            <el-timeline-item
              v-for="log in result.nodeLogs"
              :key="log.nodeId"
              :timestamp="formatDate(log.finishedAt)"
              :type="log.success ? 'success' : 'danger'"
            >
              <div class="log-item">
                <div class="node-name">{{ log.nodeName }}</div>
                <div class="node-duration">耗时: {{ log.durationMs }}ms</div>
              </div>
            </el-timeline-item>
          </el-timeline>
        </div>
      </section>
    </div>
  </el-drawer>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import type { Flow } from '@/types/flow'

const props = defineProps<{
  modelValue: boolean
  flow: Flow
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'run', input: Record<string, any>): Promise<any>
  (e: 'close'): void
}>()

const visible = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val),
})

const inputDataStr = ref('{}')
const running = ref(false)
const result = ref<any>(null)

watch(
  () => props.flow,
  (flow) => {
    if (flow?.inputSchema) {
      // 根据Schema生成示例数据
      inputDataStr.value = JSON.stringify(generateSampleData(flow.inputSchema), null, 2)
    }
  },
  { immediate: true }
)

function generateSampleData(schema: Record<string, any>): Record<string, any> {
  const sample: Record<string, any> = {}
  const properties = schema.properties || {}
  
  Object.keys(properties).forEach((key) => {
    const property = properties[key]
    switch (property.type) {
      case 'string':
        sample[key] = ''
        break
      case 'number':
      case 'integer':
        sample[key] = 0
        break
      case 'boolean':
        sample[key] = false
        break
      case 'array':
        sample[key] = []
        break
      case 'object':
        sample[key] = {}
        break
      default:
        sample[key] = null
    }
  })
  
  return sample
}

async function handleRun() {
  let inputData: Record<string, any>
  
  try {
    inputData = JSON.parse(inputDataStr.value)
  } catch (error) {
    ElMessage.error('输入参数格式错误，请输入有效的JSON')
    return
  }
  
  running.value = true
  
  try {
    result.value = await emit('run', inputData)
  } catch (error) {
    ElMessage.error('测试运行失败')
  } finally {
    running.value = false
  }
}

function formatDate(date: string) {
  return new Date(date).toLocaleString('zh-CN')
}
</script>

<style scoped>
.test-panel {
  padding: 20px;
}

.section {
  margin-bottom: 30px;
}

.section h4 {
  margin-bottom: 15px;
  color: #303133;
}

.output-section,
.logs-section {
  margin-top: 20px;
}

.output-section h5,
.logs-section h5 {
  margin-bottom: 10px;
  color: #606266;
}

pre {
  background: #f5f7fa;
  padding: 15px;
  border-radius: 4px;
  overflow-x: auto;
  margin: 0;
}

.log-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.node-name {
  font-weight: 500;
}

.node-duration {
  font-size: 12px;
  color: #909399;
}
</style>
