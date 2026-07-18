import { useEffect } from "react"
import { useAuthStore } from "@/stores/auth"
import { Loader2 } from "lucide-react"

export default function LoginPage() {
  const { login, isAuthenticated } = useAuthStore()

  useEffect(() => {
    if (!isAuthenticated) {
      login()
    }
  }, [isAuthenticated, login])

  return (
    <div className="flex items-center justify-center h-screen">
      <div className="text-center">
        <Loader2 className="w-8 h-8 animate-spin mx-auto mb-4 text-muted-foreground" />
        <p className="text-muted-foreground">正在跳转到登录页面...</p>
      </div>
    </div>
  )
}
