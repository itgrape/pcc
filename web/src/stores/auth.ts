import { create } from "zustand"
import sdk from "@/lib/sdk"
import { getUserInfo } from "@/lib/orval/user/user"

interface UserInfo {
  name: string
  displayName?: string
  email?: string
  avatar?: string
  [key: string]: unknown
}

interface AuthState {
  token: string | null
  user: UserInfo | null
  isInitializing: boolean
  isAuthenticated: boolean
  login: () => void
  handleCallback: (code: string, state: string) => Promise<void>
  logout: () => void
  fetchUser: () => Promise<void>
}

let callbackPromise: Promise<void> | null = null

async function requestUser(): Promise<UserInfo> {
  const res = await getUserInfo()
  const user = (res.data as { data?: UserInfo }).data
  if (!user?.name) {
    throw new Error("User info response is empty")
  }
  return user
}

export const useAuthStore = create<AuthState>()((set, get) => ({
  token: localStorage.getItem("casdoor_token"),
  user: null,
  isInitializing: true,
  isAuthenticated: false,

  login: () => {
    const state = crypto.randomUUID()
    sessionStorage.setItem("casdoor_oauth_state", state)
    const signinUrl = new URL(sdk.getSigninUrl())
    signinUrl.searchParams.set("state", state)
    window.location.assign(signinUrl.toString())
  },

  handleCallback: (code, state) => {
    if (callbackPromise) {
      return callbackPromise
    }

    callbackPromise = (async () => {
      try {
        const resp = await fetch("/api/v1/auth/callback", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ code, state }),
        })
        if (!resp.ok) {
          throw new Error(`Token exchange failed: ${resp.status}`)
        }

        const data = (await resp.json()) as { data?: string }
        if (!data.data) {
          throw new Error("No access token in response")
        }

        localStorage.setItem("casdoor_token", data.data)
        set({ token: data.data })

        const user = await requestUser()
        sessionStorage.removeItem("casdoor_oauth_state")
        set({ user, isAuthenticated: true, isInitializing: false })
      } catch (error) {
        localStorage.removeItem("casdoor_token")
        set({ token: null, user: null, isAuthenticated: false, isInitializing: false })
        throw error
      }
    })()

    return callbackPromise
  },

  logout: () => {
    localStorage.removeItem("casdoor_token")
    set({ token: null, user: null, isAuthenticated: false, isInitializing: false })
  },

  fetchUser: async () => {
    const token = get().token
    if (!token) {
      set({ isInitializing: false })
      return
    }

    try {
      const user = await requestUser()
      set({ user, isAuthenticated: true, isInitializing: false })
    } catch {
      localStorage.removeItem("casdoor_token")
      set({ token: null, user: null, isAuthenticated: false, isInitializing: false })
    }
  },
}))

// Callback owns initialization while it exchanges the one-time authorization
// code. Running a stale-token user request in parallel could erase the new token.
if (window.location.pathname !== "/callback") {
  useAuthStore.getState().fetchUser()
}

window.addEventListener("pcc:unauthorized", () => {
  useAuthStore.getState().logout()
})
