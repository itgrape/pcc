import { create } from "zustand"

import { getUserInfo } from "@/lib/orval/user/user"
import { apiClient } from "@/shared/api/http"
import { queryClient } from "@/shared/api/query-client"

export interface AuthUser {
  name: string
  displayName?: string
  email?: string
  avatar?: string
  [key: string]: unknown
}

interface AuthState {
  user: AuthUser | null
  isInitializing: boolean
  isAuthenticated: boolean
  initialize: () => Promise<void>
  login: (username: string, password: string) => Promise<void>
  logout: () => Promise<void>
  clearSession: () => void
}

async function requestUser(): Promise<AuthUser> {
  const response = await getUserInfo()
  const user = (response.data as { data?: AuthUser }).data
  if (!user?.name) throw new Error("User info response is empty")
  return user
}

export const useAuthStore = create<AuthState>()((set) => ({
  user: null,
  isInitializing: true,
  isAuthenticated: false,

  initialize: async () => {
    try {
      const user = await requestUser()
      set({ user, isAuthenticated: true, isInitializing: false })
    } catch {
      set({ user: null, isAuthenticated: false, isInitializing: false })
    }
  },

  login: async (username, password) => {
    await apiClient.post("/api/v1/auth/login", { username, password })
    const user = await requestUser()
    set({ user, isAuthenticated: true, isInitializing: false })
  },

  logout: async () => {
    try {
      await apiClient.post("/api/v1/auth/logout")
    } finally {
      queryClient.clear()
      set({ user: null, isAuthenticated: false, isInitializing: false })
    }
  },

  clearSession: () => {
    queryClient.clear()
    set({ user: null, isAuthenticated: false, isInitializing: false })
  },
}))
