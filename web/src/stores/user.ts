import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { User } from '@/types'
import { getUserList, createUser, deleteUser } from '@/api/user'
import type { CreateUserRequest } from '@/types'

export const useUserStore = defineStore('user', () => {
  const users = ref<User[]>([])
  const total = ref(0)
  const loading = ref(false)

  async function fetchUsers(page = 1, pageSize = 20) {
    loading.value = true
    try {
      const { data } = await getUserList(page, pageSize)
      users.value = data.data.items
      total.value = data.data.total
    } finally {
      loading.value = false
    }
  }

  async function addUser(req: CreateUserRequest) {
    await createUser(req)
    await fetchUsers()
  }

  async function removeUser(id: number) {
    await deleteUser(id)
    await fetchUsers()
  }

  return {
    users,
    total,
    loading,
    fetchUsers,
    addUser,
    removeUser,
  }
})
