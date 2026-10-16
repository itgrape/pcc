import { useState, type FormEvent } from "react"
import axios from "axios"
import { LockKeyhole, ShieldCheck } from "lucide-react"
import { Navigate, useNavigate } from "react-router-dom"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { SessionUnavailableError, useAuthStore } from "./store"

export default function LoginPage() {
  const navigate = useNavigate()
  const { login, isAuthenticated, isInitializing } = useAuthStore()
  const [username, setUsername] = useState("")
  const [password, setPassword] = useState("")
  const [error, setError] = useState("")
  const [pending, setPending] = useState(false)

  if (!isInitializing && isAuthenticated) return <Navigate to="/" replace />

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (pending) return
    setError("")
    setPending(true)
    try {
      await login(username, password)
      setPassword("")
      navigate("/", { replace: true })
    } catch (cause) {
      if (cause instanceof SessionUnavailableError) {
        setError("登录已成功，但会话未建立。请检查浏览器 Cookie 设置；使用 HTTP 时需启用 HTTP 测试模式。")
      } else if (axios.isAxiosError(cause) && cause.response?.status === 429) {
        setError("尝试次数过多，请稍后再试。")
      } else if (axios.isAxiosError(cause) && cause.response?.data?.message === "HTTPS is required for login") {
        setError("当前服务要求 HTTPS 登录。请改用 HTTPS，或联系管理员启用 HTTP 测试模式。")
      } else {
        setError("用户名或密码错误，请重试。")
      }
    } finally {
      setPending(false)
    }
  }

  return (
    <main aria-label="登录" className="surface-grid flex h-screen items-center justify-center overflow-y-auto bg-background px-5 py-10">
      <div className="my-auto w-full max-w-md shrink-0 rounded-3xl border bg-card/95 p-8 shadow-xl sm:p-10">
        <div className="mb-8 flex h-12 w-12 items-center justify-center rounded-2xl bg-primary text-primary-foreground shadow-md">
          <LockKeyhole className="h-6 w-6" />
        </div>
        <form onSubmit={submit} className="space-y-5">
          <div className="space-y-2">
            <label htmlFor="username" className="text-sm font-medium">用户名</label>
            <Input id="username" name="username" autoComplete="username" required maxLength={256} value={username} onChange={(event) => setUsername(event.target.value)} autoFocus />
          </div>
          <div className="space-y-2">
            <label htmlFor="password" className="text-sm font-medium">密码</label>
            <Input id="password" name="password" type="password" autoComplete="current-password" required value={password} onChange={(event) => setPassword(event.target.value)} />
          </div>
          {error && <p role="alert" className="text-sm text-destructive">{error}</p>}
          <Button type="submit" className="w-full" size="lg" disabled={pending || isInitializing}>
            {pending ? "正在验证…" : "登录"}
          </Button>
        </form>
        <p className="mt-7 flex items-center gap-2 text-xs text-muted-foreground"><ShieldCheck className="h-4 w-4" />仅允许管理员账号访问</p>
      </div>
    </main>
  )
}
