import { CloudCog, Loader2, ShieldCheck } from "lucide-react"

interface AuthLoadingScreenProps {
  title: string
  description: string
}

export function AuthLoadingScreen({ title, description }: AuthLoadingScreenProps) {
  return (
    <div className="surface-grid flex min-h-screen items-center justify-center bg-background p-6">
      <div className="relative w-full max-w-md overflow-hidden rounded-[28px] border bg-card/85 p-8 text-center shadow-2xl backdrop-blur-xl">
        <div className="absolute inset-x-0 top-0 h-1 bg-gradient-to-r from-indigo-500 via-violet-500 to-fuchsia-500" />
        <div className="mx-auto flex h-16 w-16 items-center justify-center rounded-2xl bg-gradient-to-br from-indigo-500 via-violet-500 to-fuchsia-500 text-white shadow-xl shadow-violet-500/20">
          <CloudCog className="h-8 w-8" />
        </div>
        <p className="mt-5 text-xs font-semibold uppercase tracking-[0.2em] text-primary">PCC Console</p>
        <h1 className="mt-2 text-xl font-semibold tracking-tight">{title}</h1>
        <p className="mt-2 text-sm text-muted-foreground">{description}</p>
        <div className="mt-7 flex items-center justify-center gap-2 rounded-xl bg-muted/60 px-4 py-3 text-xs text-muted-foreground">
          <Loader2 className="h-4 w-4 animate-spin text-primary" />
          正在建立安全连接
        </div>
        <div className="mt-5 flex items-center justify-center gap-1.5 text-[11px] text-muted-foreground">
          <ShieldCheck className="h-3.5 w-3.5 text-emerald-500" />
          Casdoor 安全身份认证
        </div>
      </div>
    </div>
  )
}
