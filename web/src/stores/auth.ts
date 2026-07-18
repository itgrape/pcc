import { create } from "zustand"
import sdk from "@/lib/sdk"
import { getUserInfo } from "@/lib/orval/user/user"

interface UserInfo {
  username: string
}

interface AuthState {
  token: string | null
  user: UserInfo | null
  isInitializing: boolean
  isAuthenticated: boolean
  login: () => void
  handleCallback: () => Promise<void>
  logout: () => void
  fetchUser: () => Promise<void>
}

export const useAuthStore = create<AuthState>()((set, get) => ({
  token: localStorage.getItem("casdoor_token"),
  user: null,
  isInitializing: true,
  isAuthenticated: false,

  login: () => {
    sdk.signin_redirect()
  },

  handleCallback: async () => {
    try {
      const tokenResponse = await sdk.exchangeForAccessToken()
      const token = tokenResponse.access_token
      if (token) {
        localStorage.setItem("casdoor_token", token)
        set({ token, isAuthenticated: true })
      } else {
        get().logout()
      }
    } catch (error) {
      console.error("Callback failed:", error)
      get().logout()
    }
  },

  logout: () => {
    localStorage.removeItem("casdoor_token")
    set({ token: null, user: null, isAuthenticated: false })
  },

  fetchUser: async () => {
    const token = get().token
    if (!token) {
      set({ isInitializing: false })
      return
    }

    try {
      const res = await getUserInfo()
      const user = (res?.data as unknown as { data?: UserInfo })?.data ?? null
      set({ user, isAuthenticated: true, isInitializing: false })
    } catch {
      set({ isInitializing: false })
    }
  },
}))

// 模块级别启动初始化
useAuthStore.getState().fetchUser()
