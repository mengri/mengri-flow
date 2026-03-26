import { ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type {
  CreateAccountRequest,
  ChangeStatusRequest,
  AccountResponse,
} from '@/types'
import { useAccountStore } from '@/stores/account'

/** 管理员账号管理的组合式函数 */
export function useAccountAdmin() {
  const store = useAccountStore()
  const createDialogVisible = ref(false)
  const detailDialogVisible = ref(false)
  const creating = ref(false)

  /** 创建账号 */
  async function handleCreate(form: CreateAccountRequest): Promise<boolean> {
    creating.value = true
    try {
      await store.create(form)
      ElMessage.success('Account created, activation email sent')
      createDialogVisible.value = false
      return true
    } catch {
      return false
    } finally {
      creating.value = false
    }
  }

  /** 变更账号状态，带确认对话框 */
  async function handleStatusChange(
    account: AccountResponse,
    action: ChangeStatusRequest['action'],
  ): Promise<boolean> {
    const actionLabels: Record<string, string> = {
      lock: 'Lock',
      unlock: 'Unlock',
      disable: 'Disable',
      enable: 'Enable',
    }
    const label = actionLabels[action] ?? action

    try {
      const { value: reason } = await ElMessageBox.prompt(
        `Are you sure to ${label.toLowerCase()} account "${account.displayName}"?`,
        `${label} Account`,
        {
          confirmButtonText: label,
          cancelButtonText: 'Cancel',
          inputPlaceholder: 'Reason (optional)',
          inputType: 'textarea',
        },
      )
      await store.changeStatus(account.accountId, { action, reason: reason || '' })
      ElMessage.success(`Account ${label.toLowerCase()}ed successfully`)
      return true
    } catch {
      // 用户取消
      return false
    }
  }

  /** 重发激活邮件 */
  async function handleResend(accountId: string): Promise<boolean> {
    try {
      await store.resend(accountId)
      ElMessage.success('Activation email resent')
      return true
    } catch {
      return false
    }
  }

  /** 查看详情 */
  async function handleViewDetail(accountId: string): Promise<void> {
    await store.fetchDetail(accountId)
    detailDialogVisible.value = true
  }

  return {
    createDialogVisible,
    detailDialogVisible,
    creating,
    handleCreate,
    handleStatusChange,
    handleResend,
    handleViewDetail,
  }
}
