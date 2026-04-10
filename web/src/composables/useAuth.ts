import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '@/stores/auth'
import { useWorkspaceStore } from '@/stores/workspace'

/** 认证相关的组合式函数 */
export function useAuth() {
  const router = useRouter()
  const authStore = useAuthStore()
  const workspaceStore = useWorkspaceStore()

  /** 执行登录并跳转 */
  async function handleLogin(account: string, password: string): Promise<boolean> {
    try {
      await authStore.login(account, password)
      ElMessage.success('Login successful')

      // 登录后加载工作空间列表
      const status = await workspaceStore.loadWorkspaces()

      if (status === 'none') {
        // 无已选中的 workspace，跳转到选择页
        await router.push('/select-workspace')
        return true
      }

      // 优先跳转到 redirect 指定的页面，否则根据角色跳转
      const redirect = router.currentRoute.value.query.redirect as string
      if (redirect) {
        await router.push(redirect)
      } else if (authStore.isAdmin) {
        await router.push('/admin/accounts')
      } else {
        // 跳转到当前工作空间的 dashboard
        const wsId = workspaceStore.currentWorkspaceId
        await router.push(wsId ? `/workspace/${wsId}` : '/')
      }
      return true
    } catch (error) {
      console.error('Login failed:', error)
      ElMessage.error(error instanceof Error ? error.message : 'Login failed')
      return false
    }
  }

  /** 登出并跳转到登录页 */
  async function handleLogout(): Promise<void> {
    workspaceStore.clearCurrentWorkspace()
    await authStore.logout()
    ElMessage.success('Logged out')
    await router.push('/login')
  }

  return {
    handleLogin,
    handleLogout,
    isAuthenticated: authStore.isAuthenticated,
    isAdmin: authStore.isAdmin,
    profile: authStore.profile,
    displayName: authStore.displayName,
  }
}
