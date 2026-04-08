<template>
  <div class="tool-detail" v-if="tool">
    <div class="header">
      <router-link to="/tools" class="back-link">
        <el-icon><ArrowLeft /></el-icon>
        返回列表
      </router-link>
      <h1>工具详情: {{ tool.name }}</h1>
      <div class="actions">
        <el-button @click="handleTest">测试</el-button>
        <el-button type="primary" @click="handleEdit">编辑</el-button>
        <el-button
          :type="tool.status === 'published' ? 'warning' : 'success'"
          @click="handleToggleStatus"
        >
          {{ tool.status === 'published' ? '下线' : '发布' }}
        </el-button>
        <el-button type="danger" @click="handleDelete">删除</el-button>
      </div>
    </div>

    <!-- 基本信息 -->
    <el-card class="info-card">
      <template #header>
        <span>基本信息</span>
      </template>
      <el-descriptions :column="2" border>
        <el-descriptions-item label="名称">{{ tool.name }}</el-descriptions-item>
        <el-descriptions-item label="类型">{{ tool.type }}</el-descriptions-item>
        <el-descriptions-item label="方法">{{ tool.method || '-' }}</el-descriptions-item>
        <el-descriptions-item label="路径">{{ tool.path || '-' }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="statusTagType(tool.status)">
            {{ statusText(tool.status) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="版本">v{{ tool.currentVersion }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatDate(tool.createdAt) }}</el-descriptions-item>
        <el-descriptions-item label="更新时间">{{ formatDate(tool.updatedAt) }}</el-descriptions-item>
        <el-descriptions-item label="描述" :span="2">{{ tool.description }}</el-descriptions-item>
        <el-descriptions-item label="标签" :span="2">
          <el-tag
            v-for="tag in tool.tags"
            :key="tag"
            size="small"
            style="margin-right: 5px"
          >
            {{ tag }}
          </el-tag>
        </el-descriptions-item>
      </el-descriptions>
    </el-card>

    <!-- Schema信息 -->
    <div class="schema-section">
      <el-card class="schema-card">
        <template #header>
          <span>输入Schema</span>
        </template>
        <pre>{{ JSON.stringify(tool.inputSchema, null, 2) }}</pre>
      </el-card>

      <el-card class="schema-card">
        <template #header>
          <span>输出Schema</span>
        </template>
        <pre>{{ JSON.stringify(tool.outputSchema, null, 2) }}</pre>
      </el-card>
    </div>

    <!-- 测试面板 -->
    <ToolTestPanel :tool="tool" class="test-panel" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { toolAPI } from '@/api/tools'
import ToolTestPanel from '@/components/ToolTestPanel.vue'
import type { Tool } from '@/types/tool'
import { formatDate } from '@/utils/request'

const route = useRoute()
const router = useRouter()

const tool = ref<Tool>()

async function loadTool() {
  const id = route.params.id as string
  try {
    tool.value = await toolAPI.get(id)
  } catch (error) {
    ElMessage.error('加载工具详情失败')
  }
}

async function handleTest() {
  if (!tool.value) return
  try {
    await toolAPI.test({
      toolId: tool.value.id,
      input: {},
    })
    ElMessage.success('工具测试成功')
  } catch (error) {
    ElMessage.error('工具测试失败')
  }
}

function handleEdit() {
  if (!tool.value) return
  router.push(`/tools/${tool.value.id}/edit`)
}

async function handleToggleStatus() {
  if (!tool.value) return
  try {
    if (tool.value.status === 'published') {
      await toolAPI.deprecate(tool.value.id)
      ElMessage.success('工具已下线')
    } else {
      await toolAPI.publish(tool.value.id)
      ElMessage.success('工具已发布')
    }
    loadTool()
  } catch (error) {
    ElMessage.error('操作失败')
  }
}

function handleDelete() {
  if (!tool.value) return
  ElMessageBox.confirm('确定要删除该工具吗？', '警告', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning',
  })
    .then(async () => {
      await toolAPI.delete(tool.value!.id)
      ElMessage.success('删除成功')
      router.push('/tools')
    })
    .catch(() => {})
}

function statusTagType(status: string) {
  const map = {
    published: 'success',
    draft: 'info',
    deprecated: 'warning',
  }
  return map[status] || 'info'
}

function statusText(status: string) {
  const map = {
    published: '已发布',
    draft: '草稿',
    deprecated: '已下线',
  }
  return map[status] || status
}

onMounted(() => {
  loadTool()
})
</script>

<style scoped>
.tool-detail {
  padding: 20px;
}

.header {
  display: flex;
  align-items: center;
  margin-bottom: 30px;
}

.back-link {
  display: flex;
  align-items: center;
  color: #409eff;
  text-decoration: none;
  margin-right: 20px;
}

.actions {
  margin-left: auto;
}

.info-card {
  margin-bottom: 30px;
}

.schema-section {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
  margin-bottom: 30px;
}

.schema-card {
  height: 400px;
  overflow: auto;
}

.test-panel {
  margin-bottom: 30px;
}

pre {
  background: #f5f7fa;
  padding: 10px;
  border-radius: 4px;
  overflow-x: auto;
  margin: 0;
}
</style>
