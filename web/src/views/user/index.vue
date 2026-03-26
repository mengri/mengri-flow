<script setup lang="ts">
import { useUser } from '@/composables/useUser'
import CreateUserDialog from './components/CreateUserDialog.vue'
import { ref } from 'vue'

const { users, total, loading, currentPage, pageSize, handleDelete, handlePageChange, loadUsers } = useUser()
const showCreateDialog = ref(false)

function onCreated() {
  showCreateDialog.value = false
  loadUsers()
}
</script>

<template>
  <div class="p-6 max-w-5xl mx-auto">
    <div class="flex justify-between items-center mb-6">
      <h2 class="text-2xl font-bold">User Management</h2>
      <el-button type="primary" @click="showCreateDialog = true">
        Create User
      </el-button>
    </div>

    <el-table :data="users" v-loading="loading" stripe border>
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="username" label="Username" />
      <el-table-column prop="email" label="Email" />
      <el-table-column prop="status" label="Status" width="100">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'danger'">
            {{ row.status === 1 ? 'Active' : 'Inactive' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="Created At" width="200" />
      <el-table-column label="Actions" width="150">
        <template #default="{ row }">
          <el-popconfirm title="Are you sure?" @confirm="handleDelete(row.id)">
            <template #reference>
              <el-button type="danger" size="small" text>Delete</el-button>
            </template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>

    <div class="mt-4 flex justify-end">
      <el-pagination
        v-model:current-page="currentPage"
        :page-size="pageSize"
        :total="total"
        layout="prev, pager, next"
        @current-change="handlePageChange"
      />
    </div>

    <CreateUserDialog
      v-model:visible="showCreateDialog"
      @created="onCreated"
    />
  </div>
</template>
