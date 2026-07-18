import { useEffect } from "react"
import { useNavigate } from "react-router-dom"
import { useAuthStore } from "@/stores/auth"
import { Loader2 } from "lucide-react"

export default function CallbackPage() {
  const navigate = useNavigate()
  const { handleCallback } = useAuthStore()

  useEffect(() => {
    const code = new URLSearchParams(window.location.search).get("code")
    const state = new URLSearchParams(window.location.search).get("state")
    if (!code) {
      navigate("/login", { replace: true })
      return
    }

    const storedState = sessionStorage.getItem("pkce_state")
    if (storedState && state && state !== storedState) {
      console.error("State mismatch, possible CSRF attack")
      navigate("/login", { replace: true })
      return
    }

    // 清除旧 token，避免干扰
    localStorage.removeItem("casdoor_token")

    handleCallback()
      .then(() => {
        navigate("/", { replace: true })
      })
      .catch((err) => {
        console.error("Login callback failed:", err)
        navigate("/login", { replace: true })
      })
  }, [navigate, handleCallback])

  return (
    <div className="flex items-center justify-center h-screen">
      <div className="text-center">
        <Loader2 className="w-8 h-8 animate-spin mx-auto mb-4 text-muted-foreground" />
        <p className="text-muted-foreground">正在完成登录...</p>
      </div>
    </div>
  )
}
