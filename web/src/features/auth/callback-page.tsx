import { useEffect, useRef } from "react"
import { useNavigate } from "react-router-dom"

import { AuthLoadingScreen } from "./auth-loading-screen"
import { useAuthStore } from "./store"

export default function CallbackPage() {
  const navigate = useNavigate()
  const { handleCallback, logout } = useAuthStore()
  const started = useRef(false)

  useEffect(() => {
    if (started.current) {
      return
    }
    started.current = true

    const params = new URLSearchParams(window.location.search)
    const code = params.get("code")
    const state = params.get("state")
    const storedState = sessionStorage.getItem("casdoor_oauth_state")
    if (!code || !state || !storedState) {
      logout()
      navigate("/login", { replace: true })
      return
    }

    if (state !== storedState) {
      console.error("State mismatch, possible CSRF attack")
      logout()
      navigate("/login", { replace: true })
      return
    }

    void handleCallback(code, state)
      .then(() => navigate("/", { replace: true }))
      .catch((error) => {
        console.error("Login callback failed:", error)
        navigate("/login", { replace: true })
      })
  }, [navigate, handleCallback, logout])

  return <AuthLoadingScreen title="正在完成登录" description="正在验证身份并加载工作空间" />
}
