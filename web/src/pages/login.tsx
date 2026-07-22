import { useEffect, useRef } from "react"
import { Navigate } from "react-router-dom"
import { useAuthStore } from "@/stores/auth"
import { Loader2 } from "lucide-react"

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

  return (
    <div className="flex items-center justify-center h-screen">
      <div className="text-center">
        <Loader2 className="w-8 h-8 animate-spin mx-auto mb-4 text-muted-foreground" />
        <p className="text-muted-foreground">正在跳转到登录页面...</p>
      </div>
    </div>
  )
}
