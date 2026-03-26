import { defineStore } from 'pinia'
import { ref, reactive } from 'vue'
import type {
  AccountResponse,
  AccountDetailResponse,
  ListAccountsRequest,
  CreateAccountRequest,
  ChangeStatusRequest,
  AuditEventItem,
  AuditEventFilter,
} from '@/types'
import {
  listAccounts,
  getAccountDetail,
  createAccount,
  changeAccountStatus,
  resendActivation,
  listAuditEvents,
} from '@/api/account'

export const useAccountStore = defineStore('account', () => {
  // --- State ---
  const accounts = ref<AccountResponse[]>([])
  const total = ref(0)
  const loading = ref(false)
  const currentDetail = ref<AccountDetailResponse | null>(null)

  const auditEvents = ref<AuditEventItem[]>([])
  const auditTotal = ref(0)
  const auditLoading = ref(false)

  const filters = reactive<ListAccountsRequest>({
    page: 1,
    pageSize: 20,
    status: '',
    keyword: '',
  })

  // --- Actions ---

  /** 拉取账号列表 */
  async function fetchAccounts(params?: Partial<ListAccountsRequest>): Promise<void> {
    if (params) {
      Object.assign(filters, params)
    }
    loading.value = true
    try {
      const { data } = await listAccounts(filters)
      accounts.value = data.data.items
      total.value = data.data.total
    } finally {
      loading.value = false
    }
  }

  /** 获取账号详情 */
  async function fetchDetail(accountId: string): Promise<AccountDetailResponse> {
    const { data } = await getAccountDetail(accountId)
    currentDetail.value = data.data
    return data.data
  }

  /** 创建账号 */
  async function create(req: CreateAccountRequest): Promise<AccountResponse> {
    const { data } = await createAccount(req)
    await fetchAccounts()
    return data.data
  }

  /** 变更状态 */
  async function changeStatus(accountId: string, req: ChangeStatusRequest): Promise<void> {
    await changeAccountStatus(accountId, req)
    await fetchAccounts()
  }

  /** 重发激活邮件 */
  async function resend(accountId: string): Promise<void> {
    await resendActivation(accountId)
  }

  /** 查询审计事件 */
  async function fetchAuditEvents(params: AuditEventFilter): Promise<void> {
    auditLoading.value = true
    try {
      const { data } = await listAuditEvents(params)
      auditEvents.value = data.data.items
      auditTotal.value = data.data.total
    } finally {
      auditLoading.value = false
    }
  }

  function reset(): void {
    accounts.value = []
    total.value = 0
    currentDetail.value = null
    auditEvents.value = []
    auditTotal.value = 0
    filters.page = 1
    filters.pageSize = 20
    filters.status = ''
    filters.keyword = ''
  }

  return {
    // state
    accounts,
    total,
    loading,
    currentDetail,
    auditEvents,
    auditTotal,
    auditLoading,
    filters,
    // actions
    fetchAccounts,
    fetchDetail,
    create,
    changeStatus,
    resend,
    fetchAuditEvents,
    reset,
  }
})
