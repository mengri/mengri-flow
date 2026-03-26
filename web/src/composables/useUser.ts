import { ref, onMounted } from 'vue'
import { useUserStore } from '@/stores/user'
import { storeToRefs } from 'pinia'
import type { CreateUserRequest } from '@/types'

/**
 * useUser — 用户管理 Composable
 * 将复杂业务逻辑从组件中抽离
 */
export function useUser() {
  const store = useUserStore()
  const { users, total, loading } = storeToRefs(store)

  const currentPage = ref(1)
  const pageSize = ref(20)

  async function loadUsers() {
    await store.fetchUsers(currentPage.value, pageSize.value)
  }

  async function handleCreate(req: CreateUserRequest) {
    await store.addUser(req)
  }

  async function handleDelete(id: number) {
    await store.removeUser(id)
  }

  function handlePageChange(page: number) {
    currentPage.value = page
    loadUsers()
  }

  onMounted(() => {
    loadUsers()
  })

  return {
    users,
    total,
    loading,
    currentPage,
    pageSize,
    loadUsers,
    handleCreate,
    handleDelete,
    handlePageChange,
  }
}
