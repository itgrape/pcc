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
  isAuthenticated: !!localStorage.getItem("casdoor_token"),

  login: () => {
    sdk.signin_redirect()
  },

  handleCallback: async () => {
    const params = new URLSearchParams(window.location.search)
    const code = params.get("code")
    const verifier = sessionStorage.getItem("pkce_code_verifier") || ""

    // 走 /casdoor-proxy 避免跨域问题
    const resp = await fetch("/casdoor-proxy/api/login/oauth/access_token", {
      method: "POST",
      headers: { "Content-Type": "application/x-www-form-urlencoded" },
      body: new URLSearchParams({
        grant_type: "authorization_code",
        code: code || "",
        client_id: import.meta.env.VITE_CASDOOR_CLIENT_ID || "",
        redirect_uri: `${window.location.origin}/callback`,
        code_verifier: verifier,
      }),
    })

    if (!resp.ok) {
      throw new Error(`Token exchange failed: ${resp.status}`)
    }

    const data = await resp.json()
    const token = data.access_token
    if (!token) {
      throw new Error("No access_token in response")
    }

    localStorage.setItem("casdoor_token", token)
    sessionStorage.clear()
    set({ token, isAuthenticated: true })

    // 获取用户信息
    try {
      const res = await getUserInfo()
      const user = (res?.data as unknown as { data?: UserInfo })?.data ?? null
      set({ user, isInitializing: false })
    } catch {
      set({ isInitializing: false })
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
      localStorage.removeItem("casdoor_token")
      set({ token: null, isAuthenticated: false, isInitializing: false })
    }
  },
}))

// 模块级别启动初始化
useAuthStore.getState().fetchUser()
