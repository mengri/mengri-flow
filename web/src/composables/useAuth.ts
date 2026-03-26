import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '@/stores/auth'

/** 认证相关的组合式函数 */
export function useAuth() {
  const router = useRouter()
  const authStore = useAuthStore()

  /** 执行登录并跳转 */
  async function handleLogin(account: string, password: string): Promise<boolean> {
    try {
      await authStore.login(account, password)
      ElMessage.success('Login successful')
      // 管理员跳后台，普通用户跳个人中心
      if (authStore.isAdmin) {
        await router.push('/admin/accounts')
      } else {
        await router.push('/account')
      }
      return true
    } catch {
      return false
    }
  }

  /** 登出并跳转到登录页 */
  async function handleLogout(): Promise<void> {
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
