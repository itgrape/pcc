import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { FolderKanban, FileCode, Key, Activity } from "lucide-react"
import { useListNamespaces } from "@/lib/orval/namespaces/namespaces"
import { useListAccessKeys } from "@/lib/orval/access-keys/access-keys"

function StatCard({ icon: Icon, label, value, sub }: { icon: any; label: string; value: string | number; sub?: string }) {
  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between pb-2">
        <CardTitle className="text-sm font-medium text-muted-foreground">{label}</CardTitle>
        <Icon className="h-4 w-4 text-muted-foreground" />
      </CardHeader>
      <CardContent>
        <div className="text-2xl font-bold">{value}</div>
        {sub && <p className="text-xs text-muted-foreground mt-1">{sub}</p>}
      </CardContent>
    </Card>
  )
}

export default function DashboardPage() {
  const { data: nsRes } = useListNamespaces({ page: 1, page_size: 1 })
  const { data: akRes } = useListAccessKeys({ page: 1, page_size: 1 })

  const nsTotal = ((nsRes?.data as { data?: { total: number } })?.data?.total) ?? 0
  const akTotal = ((akRes?.data as { data?: { total: number } })?.data?.total) ?? 0

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">仪表盘</h1>
        <p className="text-muted-foreground">配置中心概览</p>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <StatCard icon={FolderKanban} label="命名空间" value={nsTotal} sub="管理配置分组" />
        <StatCard icon={FileCode} label="配置项" value="-" sub="各命名空间下" />
        <StatCard icon={Key} label="访问密钥" value={akTotal} sub="API 认证" />
        <StatCard icon={Activity} label="系统状态" value="正常" sub="所有服务运行中" />
      </div>

      <div className="grid gap-4 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle className="text-base">快速开始</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2 text-sm text-muted-foreground">
            <p>1. 创建 <strong>命名空间</strong> 来组织你的配置</p>
            <p>2. 在命名空间下 <strong>添加配置项</strong></p>
            <p>3. 创建 <strong>访问密钥</strong> 并授权给命名空间</p>
            <p>4. 使用密钥通过 SDK 拉取配置</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-base">最近活动</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground">暂无最近活动记录</p>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
