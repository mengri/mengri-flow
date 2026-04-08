<template>
  <el-drawer
    v-model="visible"
    title="节点配置"
    :size="600"
    @close="$emit('update:modelValue', false)"
  >
    <div class="config-drawer" v-if="node">
      <h3>{{ node.data.toolName }}</h3>

      <!-- 参数映射 -->
      <section class="section">
        <h4>参数映射</h4>
        <p class="description">将上游输出或流程输入映射到工具输入参数</p>

        <el-table :data="paramMappings" border>
          <el-table-column label="工具参数" prop="target" />
          <el-table-column label="映射来源">
            <template #default="{ row }">
              <el-select
                v-model="row.source"
                placeholder="选择来源"
                clearable
                @change="(val) => updateMapping(row, val)"
              >
                <el-option-group label="流程输入">
                  <el-option
                    v-for="field in flowInputFields"
                    :key="`flow.${field}`"
                    :label="`flow.${field}`"
                    :value="`flow.${field}`"
                  />
                </el-option-group>
                <el-option-group label="上游节点">
                  <el-option
                    v-for="node in upstreamNodes"
                    :key="node.id"
                    :label="node.name"
                    :value="node.id"
                  >
                    <span class="node-option">
                      {{ node.name }}
                      <el-tag size="small" type="info">{{ node.id }}</el-tag>
                    </span>
                  </el-option>
                </el-option-group>
              </el-select>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="100">
            <template #default="{ $index }">
              <el-button
                type="danger"
                size="small"
                text
                @click="removeMapping($index)"
              >
                删除
              </el-button>
            </template>
          </el-table-column>
        </el-table>

        <el-button
          type="primary"
          size="small"
          style="margin-top: 10px"
          @click="addMapping"
        >
          添加映射
        </el-button>
      </section>

      <!-- 高级配置 -->
      <section class="section">
        <h4>高级配置</h4>
        <el-form label-width="120px">
          <el-form-item label="超时时间(ms)">
            <el-input-number
              v-model="nodeConfig.timeout"
              :min="1000"
              :max="300000"
              :step="1000"
            />
          </el-form-item>
          <el-form-item label="重试次数">
            <el-input-number v-model="nodeConfig.retry" :min="0" :max="10" />
          </el-form-item>
          <el-form-item label="执行条件">
            <el-input
              v-model="nodeConfig.condition"
              placeholder="例如: upstream.success === true"
            />
          </el-form-item>
        </el-form>
      </section>

      <!-- 保存 -->
      <div class="actions">
        <el-button type="primary" @click="handleSave">保存配置</el-button>
      </div>
    </div>
  </el-drawer>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import type { Node } from '@vue-flow/core'

const props = defineProps<{
  modelValue: boolean
  node?: Node
  flowInputSchema: Record<string, any>
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'save', nodeId: string, config: any): void
}>()

const visible = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val),
})

const paramMappings = ref<Array<{ source: string; target: string }>>([])
const nodeConfig = ref({
  timeout: 30000,
  retry: 0,
  condition: '',
})

const flowInputFields = computed(() => {
  const properties = props.flowInputSchema.properties || {}
  return Object.keys(properties)
})

const upstreamNodes = computed(() => {
  // 获取上游节点
  // 根据edges计算
  return []
})

function addMapping() {
  paramMappings.value.push({ source: '', target: '' })
}

function removeMapping(index: number) {
  paramMappings.value.splice(index, 1)
}

function updateMapping(row: any, val: string) {
  row.source = val
}

function handleSave() {
  emit('save', props.node!.id, {
    inputMapping: paramMappings.value,
    config: nodeConfig.value,
  })
}

watch(
  () => props.node,
  (node) => {
    if (node?.data) {
      paramMappings.value = node.data.inputMapping || []
      nodeConfig.value = {
        timeout: node.data.config?.timeout || 30000,
        retry: node.data.config?.retry || 0,
        condition: node.data.config?.condition || '',
      }
    }
  },
  { immediate: true }
)
</script>

<style scoped>
.config-drawer {
  padding: 20px;
}

.section {
  margin-bottom: 30px;
}

.section h4 {
  margin-bottom: 10px;
  color: #303133;
}

.description {
  color: #909399;
  font-size: 14px;
  margin-bottom: 15px;
}

.actions {
  text-align: center;
  padding: 20px 0;
}

.node-option {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}
</style>
