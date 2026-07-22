import { useEffect, useRef } from "react"
import { Navigate } from "react-router-dom"

import { AuthLoadingScreen } from "./auth-loading-screen"
import { useAuthStore } from "./store"

export default function LoginPage() {
  const { login, isAuthenticated, isInitializing } = useAuthStore()
  const redirectStarted = useRef(false)

  useEffect(() => {
    if (!isInitializing && !isAuthenticated && !redirectStarted.current) {
      redirectStarted.current = true
      login()
    }
  }, [isAuthenticated, isInitializing, login])

  if (!isInitializing && isAuthenticated) {
    return <Navigate to="/" replace />
  }

  return <AuthLoadingScreen title="欢迎回来" description="即将跳转到统一身份认证中心" />
}
