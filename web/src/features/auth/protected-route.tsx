import { Navigate, Outlet } from "react-router-dom"

import { AuthLoadingScreen } from "./auth-loading-screen"
import { useAuthStore } from "./store"

export default function ProtectedRoute() {
  const { isAuthenticated, isInitializing } = useAuthStore()

  if (isInitializing) {
    return <AuthLoadingScreen title="正在加载工作空间" description="正在恢复登录状态与用户信息" />
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />
  }

  return <Outlet />
}
