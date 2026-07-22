import { create } from "zustand"

import { getUserInfo } from "@/lib/orval/user/user"
import { apiClient } from "@/shared/api/http"
import { queryClient } from "@/shared/api/query-client"
import { casdoor } from "./casdoor"

export interface AuthUser {
  name: string
  displayName?: string
  email?: string
  avatar?: string
  [key: string]: unknown
}

interface AuthState {
  token: string | null
  user: AuthUser | null
  isInitializing: boolean
  isAuthenticated: boolean
  initialize: () => Promise<void>
  login: () => void
  handleCallback: (code: string, state: string) => Promise<void>
  logout: () => void
}

let callbackPromise: Promise<void> | null = null

async function requestUser(): Promise<AuthUser> {
  const response = await getUserInfo()
  const user = (response.data as { data?: AuthUser }).data
  if (!user?.name) {
    throw new Error("User info response is empty")
  }
  return user
}

function clearToken() {
  localStorage.removeItem("casdoor_token")
}

export const useAuthStore = create<AuthState>()((set, get) => ({
  token: localStorage.getItem("casdoor_token"),
  user: null,
  isInitializing: true,
  isAuthenticated: false,

  initialize: async () => {
    const token = get().token
    if (!token) {
      set({ isInitializing: false })
      return
    }

    try {
      const user = await requestUser()
      set({ user, isAuthenticated: true, isInitializing: false })
    } catch {
      clearToken()
      set({ token: null, user: null, isAuthenticated: false, isInitializing: false })
    }
  },

  login: () => {
    const state = crypto.randomUUID()
    sessionStorage.setItem("casdoor_oauth_state", state)
    const signinUrl = new URL(casdoor.getSigninUrl())
    signinUrl.searchParams.set("state", state)
    window.location.assign(signinUrl.toString())
  },

  handleCallback: (code, state) => {
    if (callbackPromise) {
      return callbackPromise
    }

    callbackPromise = (async () => {
      try {
        const response = await apiClient.post<{ data?: string }>("/api/v1/auth/callback", {
          code,
          state,
        })
        const payload = response.data
        if (!payload.data) {
          throw new Error("No access token in response")
        }

        localStorage.setItem("casdoor_token", payload.data)
        set({ token: payload.data })

        const user = await requestUser()
        sessionStorage.removeItem("casdoor_oauth_state")
        set({ user, isAuthenticated: true, isInitializing: false })
      } catch (error) {
        clearToken()
        set({ token: null, user: null, isAuthenticated: false, isInitializing: false })
        throw error
      }
    })()

    return callbackPromise
  },

  logout: () => {
    clearToken()
    queryClient.clear()
    set({ token: null, user: null, isAuthenticated: false, isInitializing: false })
  },
}))
