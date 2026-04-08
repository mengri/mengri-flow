<template>
  <div class="tool-test-panel">
    <el-tabs v-model="activeTab">
      <el-tab-pane label="输入参数" name="input">
        <el-input
          v-model="testInputStr"
          type="textarea"
          :rows="10"
          placeholder='请输入JSON格式的输入参数，例如：{"name": "test"}'
        />
      </el-tab-pane>
      <el-tab-pane label="响应结果" name="output">
        <pre v-if="testOutput">{{ JSON.stringify(testOutput, null, 2) }}</pre>
        <el-empty v-else description="暂无数据" />
      </el-tab-pane>
      <el-tab-pane label="执行日志" name="logs">
        <el-timeline>
          <el-timeline-item
            v-for="log in executionLogs"
            :key="log.id"
            :timestamp="log.timestamp"
            :type="logType(log.level)"
          >
            {{ log.message }}
          </el-timeline-item>
        </el-timeline>
      </el-tab-pane>
    </el-tabs>

    <div class="actions">
      <el-button type="primary" @click="executeTool" :loading="executing">
        执行
      </el-button>
      <el-button @click="reset">重置</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { toolAPI } from '@/api/tools'
import type { Tool } from '@/types/tool'

const props = defineProps<{
  tool: Tool
}>()

const activeTab = ref('input')
const testInputStr = ref('{}')
const testOutput = ref<Record<string, any>>()
const executionLogs = ref<any[]>([])
const executing = ref(false)

async function executeTool() {
  executing.value = true
  executionLogs.value = []
  
  try {
    let input = {}
    try {
      input = JSON.parse(testInputStr.value)
    } catch (error) {
      ElMessage.error('输入参数格式错误，请输入有效的JSON')
      executing.value = false
      return
    }
    
    const result = await toolAPI.test({
      toolId: props.tool.id,
      input,
    })
    testOutput.value = result
    ElMessage.success('执行成功')
  } catch (error) {
    ElMessage.error('执行失败')
  } finally {
    executing.value = false
  }
}

function reset() {
  testInputStr.value = '{}'
  testOutput.value = undefined
  executionLogs.value = []
}

function logType(level: string) {
  const map = {
    info: '',
    success: 'success',
    warning: 'warning',
    error: 'danger',
  }
  return map[level as keyof typeof map] || ''
}
</script>

<style scoped>
.tool-test-panel {
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  padding: 20px;
}

.actions {
  margin-top: 20px;
  text-align: center;
}

pre {
  background: #f5f7fa;
  padding: 10px;
  border-radius: 4px;
  overflow-x: auto;
  margin: 0;
}
</style>
