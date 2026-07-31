import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { loginAdmin, type AdminAccount } from '../api/admin'

const tokenStorageKey = 'automatictools.admin.token'
const accountStorageKey = 'automatictools.admin.account'

function readStoredAccount(): AdminAccount | null {
  try {
    const value = localStorage.getItem(accountStorageKey)
    return value ? (JSON.parse(value) as AdminAccount) : null
  } catch {
    return null
  }
}

function isExpired(token: string) {
  try {
    const payload = token.split('.')[1]
    if (!payload) return true

    const normalized = payload.replace(/-/g, '+').replace(/_/g, '/')
    const padded = normalized.padEnd(Math.ceil(normalized.length / 4) * 4, '=')
    const claims = JSON.parse(atob(padded)) as { exp?: number }
    return typeof claims.exp !== 'number' || claims.exp * 1000 <= Date.now()
  } catch {
    return true
  }
}

export const useAdminAuthStore = defineStore('adminAuth', () => {
  const storedToken = localStorage.getItem(tokenStorageKey) || ''
  const token = ref(storedToken && !isExpired(storedToken) ? storedToken : '')
  const admin = ref<AdminAccount | null>(token.value ? readStoredAccount() : null)

  if (!token.value) {
    localStorage.removeItem(tokenStorageKey)
    localStorage.removeItem(accountStorageKey)
  }

  const isAuthenticated = computed(() => Boolean(token.value && admin.value))

  async function login(username: string, password: string) {
    const response = await loginAdmin(username.trim(), password)
    token.value = response.token
    admin.value = response.admin
    localStorage.setItem(tokenStorageKey, response.token)
    localStorage.setItem(accountStorageKey, JSON.stringify(response.admin))
  }

  function logout() {
    token.value = ''
    admin.value = null
    localStorage.removeItem(tokenStorageKey)
    localStorage.removeItem(accountStorageKey)
  }

  function updateAccount(account: AdminAccount) {
    admin.value = account
    localStorage.setItem(accountStorageKey, JSON.stringify(account))
  }

  return {
    token,
    admin,
    isAuthenticated,
    login,
    logout,
    updateAccount,
  }
})
