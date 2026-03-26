import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { ProfileResponse, AccountBrief, AccountRole } from '@/types'
import { loginByPassword, logout as apiLogout } from '@/api/auth'
import { getProfile } from '@/api/account'
import { getToken, setToken, clearTokens, getRefreshToken } from '@/utils/request'

export const useAuthStore = defineStore('auth', () => {
  // --- State ---
  const accessToken = ref<string | null>(getToken())
  const refreshTokenValue = ref<string | null>(getRefreshToken())
  const profile = ref<ProfileResponse | null>(null)
  const loading = ref(false)

  // --- Getters ---
  const isAuthenticated = computed(() => !!accessToken.value)
  const role = computed<AccountRole | null>(() => profile.value?.role ?? null)
  const isAdmin = computed(() => role.value === 'admin')
  const displayName = computed(() => profile.value?.displayName ?? '')

  // --- Actions ---

  /** 密码登录 */
  async function login(account: string, password: string): Promise<AccountBrief> {
    loading.value = true
    try {
      const { data } = await loginByPassword({ account, password })
      const result = data.data
      accessToken.value = result.accessToken
      refreshTokenValue.value = result.refreshToken
      setToken(result.accessToken, result.refreshToken)
      // 登录后立即拉取 profile
      await fetchProfile()
      return result.account
    } finally {
      loading.value = false
    }
  }

  /** 获取当前用户资料 */
  async function fetchProfile(): Promise<ProfileResponse | null> {
    if (!accessToken.value) return null
    try {
      const { data } = await getProfile()
      profile.value = data.data
      return data.data
    } catch {
      profile.value = null
      return null
    }
  }

  /** 登出 */
  async function logout(): Promise<void> {
    try {
      if (accessToken.value) {
        await apiLogout()
      }
    } finally {
      accessToken.value = null
      refreshTokenValue.value = null
      profile.value = null
      clearTokens()
    }
  }

  /** 更新 token（供 token 刷新场景使用） */
  function updateTokens(newAccessToken: string, newRefreshToken: string): void {
    accessToken.value = newAccessToken
    refreshTokenValue.value = newRefreshToken
    setToken(newAccessToken, newRefreshToken)
  }

  /** 重置状态 */
  function reset(): void {
    accessToken.value = null
    refreshTokenValue.value = null
    profile.value = null
    clearTokens()
  }

  return {
    // state
    accessToken,
    refreshTokenValue,
    profile,
    loading,
    // getters
    isAuthenticated,
    role,
    isAdmin,
    displayName,
    // actions
    login,
    fetchProfile,
    logout,
    updateTokens,
    reset,
  }
})
