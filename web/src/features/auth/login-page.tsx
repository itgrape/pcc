import { useState, type FormEvent } from "react"
import { LockKeyhole, ShieldCheck } from "lucide-react"
import { Navigate, useNavigate } from "react-router-dom"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { useAuthStore } from "./store"

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
    } catch {
      setError("用户名或密码错误，请重试。")
    } finally {
      setPending(false)
    }
  }

  return (
    <main className="surface-grid flex h-screen items-center justify-center overflow-y-auto bg-background px-5 py-10">
      <div className="my-auto w-full max-w-md shrink-0 rounded-3xl border bg-card/95 p-8 shadow-xl sm:p-10">
        <div className="mb-8 flex h-12 w-12 items-center justify-center rounded-2xl bg-primary text-primary-foreground shadow-md">
          <LockKeyhole className="h-6 w-6" />
        </div>
        <p className="text-xs font-semibold uppercase tracking-[0.2em] text-primary">PCC Console</p>
        <h1 className="mt-2 text-3xl font-semibold tracking-tight">管理员登录</h1>
        <p className="mt-2 text-sm text-muted-foreground">登录后管理配置、命名空间与访问密钥。</p>
        <form onSubmit={submit} className="mt-8 space-y-5">
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
