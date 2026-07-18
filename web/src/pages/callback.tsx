import { useEffect } from "react"
import { useNavigate } from "react-router-dom"
import { useAuthStore } from "@/stores/auth"
import { Loader2 } from "lucide-react"

export default function CallbackPage() {
  const navigate = useNavigate()

  useEffect(() => {
    const code = new URLSearchParams(window.location.search).get("code")
    if (!code) {
      navigate("/login", { replace: true })
      return
    }

    // 先清除旧 token，避免干扰
    localStorage.removeItem("casdoor_token")

    const verifier = sessionStorage.getItem("pkce_code_verifier") || ""

    fetch("/casdoor-proxy/api/login/oauth/access_token", {
      method: "POST",
      headers: { "Content-Type": "application/x-www-form-urlencoded" },
      body: new URLSearchParams({
        grant_type: "authorization_code",
        code,
        client_id: import.meta.env.VITE_CASDOOR_CLIENT_ID || "",
        redirect_uri: `${window.location.origin}/callback`,
        code_verifier: verifier,
      }),
    })
      .then((r) => r.json())
      .then((data) => {
        if (!data.access_token) throw new Error("no access_token")

        const t = data.access_token
        localStorage.setItem("casdoor_token", t)
        sessionStorage.clear()
        useAuthStore.setState({ token: t, user: null, isAuthenticated: true, isInitializing: false })
        navigate("/", { replace: true })
      })
      .catch(() => {
        navigate("/login", { replace: true })
      })
  }, [navigate])

  return (
    <div className="flex items-center justify-center h-screen">
      <div className="text-center">
        <Loader2 className="w-8 h-8 animate-spin mx-auto mb-4 text-muted-foreground" />
        <p className="text-muted-foreground">正在完成登录...</p>
      </div>
    </div>
  )
}
