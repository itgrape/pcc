import type { LucideIcon } from "lucide-react"
import {
  Activity,
  ArrowUpRight,
  CircleCheck,
  CloudCog,
  FileCode2,
  FolderKanban,
  KeyRound,
  ShieldCheck,
  Zap,
} from "lucide-react"
import { Link } from "react-router-dom"
import { useQuery } from "@tanstack/react-query"

import { Card, CardContent } from "@/components/ui/card"
import { apiClient } from "@/shared/api/http"

interface DashboardStats {
  namespaces: number
  configs: number
  access_keys: number
}

function StatCard({
  icon: Icon,
  label,
  value,
  sub,
  accent,
}: {
  icon: LucideIcon
  label: string
  value: string | number
  sub: string
  accent: string
}) {
  return (
    <Card className="group relative overflow-hidden border-border/60 bg-card/80 shadow-sm backdrop-blur transition duration-300 hover:-translate-y-0.5 hover:shadow-xl">
      <div className={`absolute inset-x-0 top-0 h-1 ${accent}`} />
      <CardContent className="p-5">
        <div className="flex items-start justify-between">
          <div>
            <p className="text-sm font-medium text-muted-foreground">{label}</p>
            <p className="mt-3 text-3xl font-semibold tracking-tight">{value}</p>
          </div>
          <div className="rounded-2xl bg-primary/8 p-3 text-primary transition group-hover:scale-110">
            <Icon className="h-5 w-5" />
          </div>
        </div>
        <p className="mt-3 text-xs text-muted-foreground">{sub}</p>
      </CardContent>
    </Card>
  )
}

export default function DashboardPage() {
  const { data: stats } = useQuery({
    queryKey: ["dashboard-stats"],
    queryFn: async () => {
      const response = await apiClient.get<{ data: DashboardStats }>(
        "/api/v1/admin/system/dashboard-stats",
      )
      return response.data.data
    },
  })

  return (
    <div className="space-y-8">
      <section className="relative overflow-hidden rounded-3xl border border-border/70 bg-card/85 p-7 shadow-sm backdrop-blur md:p-9">
        <div className="absolute inset-y-0 left-0 w-1 bg-primary" />
        <div className="absolute -right-20 -top-24 h-72 w-72 rounded-full bg-primary/[0.045]" />
        <div className="relative max-w-3xl">
          <div className="mb-5 flex h-12 w-12 items-center justify-center rounded-2xl border bg-primary/8 text-primary shadow-sm">
            <CloudCog className="h-6 w-6" />
          </div>
          <p className="mb-2 text-xs font-semibold uppercase tracking-[0.18em] text-primary">
            Configuration workspace
          </p>
          <h1 className="text-3xl font-semibold tracking-tight text-foreground md:text-4xl">
            让每一次配置变更都清晰、可靠
          </h1>
          <p className="mt-3 max-w-2xl text-sm leading-6 text-muted-foreground md:text-base">
            集中管理多环境配置、访问凭证和版本历史，让团队更安全地交付服务。
          </p>
          <div className="mt-6 flex flex-wrap gap-3">
            <Link
              to="/namespaces"
              className="inline-flex items-center gap-2 rounded-xl bg-primary px-4 py-2.5 text-sm font-semibold text-primary-foreground shadow-sm transition hover:-translate-y-0.5 hover:bg-primary/90 hover:shadow-md"
            >
              管理命名空间 <ArrowUpRight className="h-4 w-4" />
            </Link>
            <Link
              to="/access-keys"
              className="inline-flex items-center gap-2 rounded-xl border bg-background/70 px-4 py-2.5 text-sm font-medium text-foreground shadow-sm transition hover:bg-accent"
            >
              <ShieldCheck className="h-4 w-4" /> 查看访问密钥
            </Link>
          </div>
        </div>
        <div className="absolute bottom-8 right-8 hidden rounded-2xl border bg-background/65 p-4 shadow-sm backdrop-blur lg:block">
          <div className="flex items-center gap-3">
            <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-emerald-500/10">
              <CircleCheck className="h-5 w-5 text-emerald-500" />
            </div>
            <div>
              <p className="text-sm font-semibold">系统运行正常</p>
              <p className="text-xs text-muted-foreground">All systems operational</p>
            </div>
          </div>
        </div>
      </section>

      <div className="flex items-end justify-between">
        <div>
          <h2 className="text-xl font-semibold tracking-tight">资源概览</h2>
          <p className="mt-1 text-sm text-muted-foreground">工作空间内的实时资源状态</p>
        </div>
        <div className="hidden items-center gap-2 text-xs text-muted-foreground sm:flex">
          <Activity className="h-3.5 w-3.5 text-emerald-500" />
          数据已同步
        </div>
      </div>

      <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <StatCard icon={FolderKanban} label="命名空间" value={stats?.namespaces ?? "—"} sub="隔离不同业务和环境" accent="bg-indigo-500" />
        <StatCard icon={FileCode2} label="配置项" value={stats?.configs ?? "—"} sub="统一托管结构化配置" accent="bg-violet-500" />
        <StatCard icon={KeyRound} label="访问密钥" value={stats?.access_keys ?? "—"} sub="精细控制 API 权限" accent="bg-amber-500" />
        <StatCard icon={Zap} label="系统状态" value="健康" sub="服务与数据链路正常" accent="bg-emerald-500" />
      </div>

      <div className="grid gap-5 lg:grid-cols-[1.35fr_1fr]">
        <Card className="border-border/60 bg-card/80 shadow-sm">
          <CardContent className="p-6">
            <div className="mb-5 flex items-center justify-between">
              <div>
                <h3 className="font-semibold">推荐工作流</h3>
                <p className="mt-1 text-sm text-muted-foreground">四步完成安全的配置交付</p>
              </div>
              <Zap className="h-5 w-5 text-amber-500" />
            </div>
            <div className="grid gap-3 sm:grid-cols-2">
              {[
                ["01", "创建命名空间", "按业务或环境隔离配置"],
                ["02", "添加配置项", "使用编辑器维护配置内容"],
                ["03", "创建访问密钥", "为调用方生成独立凭证"],
                ["04", "授权并发布", "按命名空间分配读写权限"],
              ].map(([step, title, description]) => (
                <div key={step} className="rounded-2xl border bg-muted/25 p-4 transition hover:bg-muted/50">
                  <div className="flex items-center gap-3">
                    <span className="flex h-8 w-8 items-center justify-center rounded-xl bg-primary/10 text-xs font-bold text-primary">{step}</span>
                    <div>
                      <p className="text-sm font-medium">{title}</p>
                      <p className="mt-0.5 text-xs text-muted-foreground">{description}</p>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

        <Card className="border-border/60 bg-card/80 shadow-sm">
          <CardContent className="p-6">
            <h3 className="font-semibold">平台能力</h3>
            <p className="mt-1 text-sm text-muted-foreground">为生产环境设计的核心保障</p>
            <div className="mt-5 space-y-3">
              {[
                ["版本历史", "完整记录每一次配置变更"],
                ["权限隔离", "命名空间级别授权控制"],
                ["安全凭证", "密钥仅在创建时展示一次"],
              ].map(([title, description]) => (
                <div key={title} className="flex gap-3 rounded-xl p-2">
                  <span className="mt-0.5 flex h-7 w-7 shrink-0 items-center justify-center rounded-lg bg-emerald-500/10 text-emerald-600 dark:text-emerald-400">
                    <CircleCheck className="h-4 w-4" />
                  </span>
                  <div>
                    <p className="text-sm font-medium">{title}</p>
                    <p className="text-xs text-muted-foreground">{description}</p>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
