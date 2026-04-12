<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useWorkspaceStore } from '@/stores/workspace'
import { useWorkspaceRoute } from '@/composables/useWorkspaceRoute'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const workspaceStore = useWorkspaceStore()
const { workspaceId } = useWorkspaceRoute()

const formRef = ref<FormInstance>()
const loading = ref(false)
const saving = ref(false)

const form = reactive({
  name: '',
  description: '',
})

const rules: FormRules = {
  name: [
    { required: true, message: t('workspace.nameRequired'), trigger: 'blur' },
    { max: 100, message: t('workspace.nameMaxLength'), trigger: 'blur' },
  ],
  description: [
    { max: 500, message: t('workspace.descriptionMaxLength'), trigger: 'blur' },
  ],
}

async function fetchWorkspace() {
  const ws = workspaceStore.currentWorkspace
  if (ws) {
    form.name = ws.name
    form.description = ws.description
    return
  }
  loading.value = true
  try {
    const { workspaceAPI } = await import('@/api/workspaces')
    const ws = await workspaceAPI.get(workspaceId.value)
    form.name = ws.name
    form.description = ws.description
  } finally {
    loading.value = false
  }
}

async function handleSave() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  saving.value = true
  try {
    const updated = await workspaceStore.updateWorkspace(workspaceId.value, {
      name: form.name,
      description: form.description,
    })
    form.name = updated.name
    form.description = updated.description
    ElMessage.success(t('workspace.updateSuccess'))
  } finally {
    saving.value = false
  }
}

async function handleDelete() {
  try {
    await ElMessageBox.confirm(
      t('workspace.deleteConfirm'),
      t('common.warning'),
      { type: 'warning', confirmButtonText: t('common.delete'), cancelButtonText: t('common.cancel') },
    )
    await workspaceStore.deleteWorkspace(workspaceId.value)
    ElMessage.success(t('workspace.deleteSuccess'))
  } catch {
    // cancelled
  }
}

onMounted(() => {
  fetchWorkspace()
})
</script>

<template>
  <div class="workspace-settings" v-loading="loading">
    <div class="mb-6">
      <h2 class="text-xl font-bold text-gray-800">{{ t('workspace.settings') }}</h2>
      <p class="text-sm text-gray-500 mt-1">{{ t('workspace.settingsDesc') }}</p>
    </div>

    <el-card shadow="never" class="max-w-2xl">
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-position="top"
      >
        <el-form-item :label="t('workspace.name')" prop="name">
          <el-input v-model="form.name" :placeholder="t('workspace.namePlaceholder')" maxlength="100" show-word-limit />
        </el-form-item>

        <el-form-item :label="t('workspace.description')" prop="description">
          <el-input
            v-model="form.description"
            type="textarea"
            :rows="3"
            :placeholder="t('workspace.descriptionPlaceholder')"
            maxlength="500"
            show-word-limit
          />
        </el-form-item>

        <el-form-item>
          <el-button type="primary" :loading="saving" @click="handleSave">
            {{ t('common.save') }}
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card shadow="never" class="max-w-2xl mt-6">
      <div class="flex items-center justify-between">
        <div>
          <h3 class="text-base font-semibold text-red-600">{{ t('workspace.dangerZone') }}</h3>
          <p class="text-sm text-gray-500 mt-1">{{ t('workspace.deleteWarning') }}</p>
        </div>
        <el-button type="danger" plain @click="handleDelete">
          {{ t('common.delete') }}
        </el-button>
      </div>
    </el-card>
  </div>
</template>
