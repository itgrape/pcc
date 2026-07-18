import { useState } from "react"
import { useParams, Link } from "react-router-dom"
import { useQueryClient } from "@tanstack/react-query"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import { Card, CardContent } from "@/components/ui/card"
import {
  Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription,
  DialogFooter, DialogTrigger,
} from "@/components/ui/dialog"
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from "@/components/ui/select"
import { Badge } from "@/components/ui/badge"
import { Plus, Pencil, Trash2, History, RotateCcw } from "lucide-react"
import {
  useListConfigs,
  useCreateConfig,
  useUpdateConfig,
  useDeleteConfig,
  useListConfigHistory,
  useRollbackConfig,
  getListConfigsQueryKey,
} from "@/lib/orval/configs/configs"
import { useGetNamespace } from "@/lib/orval/namespaces/namespaces"

interface ConfigItem {
  id: string
  namespace_id: string
  config_name: string
  format: number
  content?: string
  description: string
  version: number
  created_by?: string
  updated_by?: string
  created_at: string
  updated_at: string
}

interface ConfigHistoryItem {
  id: string
  version: number
  operator: string
  change_reason: string
  content?: string
  created_at: string
}

interface NamespaceDetail {
  namespace: string
}

const formatLabels: Record<number, string> = { 1: "JSON", 2: "YAML", 3: "Text", 4: "Properties" }
const formatColors: Record<number, string> = {
  1: "bg-blue-50 text-blue-700 dark:bg-blue-950 dark:text-blue-300",
  2: "bg-amber-50 text-amber-700 dark:bg-amber-950 dark:text-amber-300",
  3: "bg-slate-50 text-slate-700 dark:bg-slate-950 dark:text-slate-300",
  4: "bg-purple-50 text-purple-700 dark:bg-purple-950 dark:text-purple-300",
}

export default function ConfigsPage() {
  const { namespaceId } = useParams<{ namespaceId: string }>()
  const nsId = Number(namespaceId)
  const queryClient = useQueryClient()

  const [open, setOpen] = useState(false)
  const [editOpen, setEditOpen] = useState(false)
  const [editConfig, setEditConfig] = useState<ConfigItem | null>(null)
  const [historyOpen, setHistoryOpen] = useState(false)
  const [historyConfig, setHistoryConfig] = useState<ConfigItem | null>(null)
  const [rollbackOpen, setRollbackOpen] = useState(false)
  const [rollbackTarget, setRollbackTarget] = useState<ConfigHistoryItem | null>(null)
  const [form, setForm] = useState({ config_name: "", format: "1", content: "", description: "" })
  const [editForm, setEditForm] = useState({ content: "", description: "", change_reason: "" })
  const [rollbackReason, setRollbackReason] = useState("")

  const { data: nsRes } = useGetNamespace(nsId)
  const { data: configsRes, isLoading } = useListConfigs(nsId)
  const { data: historyRes } = useListConfigHistory(Number(historyConfig?.id))

  const invalidate = () => queryClient.invalidateQueries({ queryKey: getListConfigsQueryKey(nsId) })

  const { mutate: createMutate } = useCreateConfig()
  const { mutate: updateMutate } = useUpdateConfig()
  const { mutate: deleteMutate } = useDeleteConfig()
  const { mutate: rollbackMutate } = useRollbackConfig()

  const configs = (configsRes?.data as unknown as { data?: ConfigItem[] })?.data ?? []
  const namespaceName = (nsRes?.data as unknown as { data?: NamespaceDetail })?.data?.namespace ?? ""
  const historyList = (historyRes?.data as unknown as { data?: ConfigHistoryItem[] })?.data ?? []

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <div className="flex items-center gap-2 text-sm text-muted-foreground mb-1">
            <Link to="/namespaces" className="hover:text-foreground">命名空间</Link>
            <span>/</span>
            <span className="text-foreground font-medium">{namespaceName}</span>
          </div>
          <h1 className="text-2xl font-bold tracking-tight">配置列表</h1>
          <p className="text-muted-foreground">管理 {namespaceName} 下的配置项</p>
        </div>
        <Dialog open={open} onOpenChange={setOpen}>
          <DialogTrigger asChild>
            <Button>
              <Plus className="h-4 w-4" />新建配置
            </Button>
          </DialogTrigger>
          <DialogContent className="max-w-xl">
            <DialogHeader>
              <DialogTitle>新建配置</DialogTitle>
              <DialogDescription>在 {namespaceName} 下创建新配置</DialogDescription>
            </DialogHeader>
            <div className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label>配置名称</Label>
                  <Input value={form.config_name} onChange={(e) => setForm({ ...form, config_name: e.target.value })} placeholder="my-config" />
                </div>
                <div className="space-y-2">
                  <Label>格式</Label>
                  <Select value={form.format} onValueChange={(v) => setForm({ ...form, format: v })}>
                    <SelectTrigger><SelectValue /></SelectTrigger>
                    <SelectContent>
                      <SelectItem value="1">JSON</SelectItem>
                      <SelectItem value="2">YAML</SelectItem>
                      <SelectItem value="3">Text</SelectItem>
                      <SelectItem value="4">Properties</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
              </div>
              <div className="space-y-2">
                <Label>描述</Label>
                <Input value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} placeholder="可选" />
              </div>
              <div className="space-y-2">
                <Label>配置内容</Label>
                <Textarea className="min-h-[200px] font-mono text-sm" value={form.content} onChange={(e) => setForm({ ...form, content: e.target.value })} placeholder="输入配置内容..." />
              </div>
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => setOpen(false)}>取消</Button>
              <Button onClick={() => createMutate(
                { namespaceId: nsId, data: { config_name: form.config_name, format: parseInt(form.format), content: form.content, description: form.description || undefined } },
                { onSuccess: () => { invalidate(); setOpen(false); setForm({ config_name: "", format: "1", content: "", description: "" }) } },
              )} disabled={!form.config_name || !form.content}>创建</Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>

      <Card>
        <CardContent className="p-0">
          <table className="w-full">
            <thead>
              <tr className="border-b text-sm text-muted-foreground">
                <th className="text-left font-medium px-6 py-3">名称</th>
                <th className="text-left font-medium px-6 py-3">格式</th>
                <th className="text-left font-medium px-6 py-3">版本</th>
                <th className="text-left font-medium px-6 py-3">描述</th>
                <th className="text-left font-medium px-6 py-3">更新人</th>
                <th className="text-left font-medium px-6 py-3">更新时间</th>
                <th className="text-right font-medium px-6 py-3">操作</th>
              </tr>
            </thead>
            <tbody>
              {isLoading ? (
                <tr><td colSpan={7} className="px-6 py-12 text-center text-muted-foreground">加载中...</td></tr>
              ) : configs.length === 0 ? (
                <tr><td colSpan={7} className="px-6 py-12 text-center text-muted-foreground">暂无配置</td></tr>
              ) : (
                configs.map((cfg: ConfigItem) => (
                  <tr key={cfg.id} className="border-b last:border-0 hover:bg-muted/50">
                    <td className="px-6 py-3 font-medium">{cfg.config_name}</td>
                    <td className="px-6 py-3">
                      <span className={`inline-flex items-center rounded-md px-2 py-0.5 text-xs font-medium ${formatColors[cfg.format] || ""}`}>
                        {formatLabels[cfg.format] || "未知"}
                      </span>
                    </td>
                    <td className="px-6 py-3 text-sm text-muted-foreground">v{cfg.version}</td>
                    <td className="px-6 py-3 text-sm text-muted-foreground max-w-[200px] truncate">{cfg.description || "-"}</td>
                    <td className="px-6 py-3 text-sm text-muted-foreground">{cfg.updated_by || cfg.created_by || "-"}</td>
                    <td className="px-6 py-3 text-sm text-muted-foreground">{cfg.updated_at}</td>
                    <td className="px-6 py-3 text-right">
                      <Button variant="ghost" size="icon" onClick={() => { setEditConfig(cfg); setEditForm({ content: cfg.content || "", description: cfg.description, change_reason: "" }); setEditOpen(true) }}>
                        <Pencil className="h-4 w-4" />
                      </Button>
                      <Button variant="ghost" size="icon" onClick={() => { setHistoryConfig(cfg); setHistoryOpen(true) }}>
                        <History className="h-4 w-4" />
                      </Button>
                      <Button variant="ghost" size="icon" onClick={() => {
                        if (confirm("确定删除？")) deleteMutate({ configId: Number(cfg.id) }, { onSuccess: () => invalidate() })
                      }}>
                        <Trash2 className="h-4 w-4 text-destructive" />
                      </Button>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </CardContent>
      </Card>

      <Dialog open={editOpen} onOpenChange={setEditOpen}>
        <DialogContent className="max-w-xl">
          <DialogHeader>
            <DialogTitle>编辑配置</DialogTitle>
            <DialogDescription>{editConfig?.config_name} (v{editConfig?.version})</DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label>描述</Label>
              <Input value={editForm.description} onChange={(e) => setEditForm({ ...editForm, description: e.target.value })} />
            </div>
            <div className="space-y-2">
              <Label>配置内容</Label>
              <Textarea className="min-h-[200px] font-mono text-sm" value={editForm.content} onChange={(e) => setEditForm({ ...editForm, content: e.target.value })} />
            </div>
            <div className="space-y-2">
              <Label>变更原因</Label>
              <Input value={editForm.change_reason} onChange={(e) => setEditForm({ ...editForm, change_reason: e.target.value })} placeholder="可选" />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setEditOpen(false)}>取消</Button>
            <Button onClick={() => updateMutate(
              { configId: Number(editConfig!.id), data: editForm },
              { onSuccess: () => { invalidate(); setEditOpen(false) } },
            )}>保存</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={historyOpen} onOpenChange={(v) => { setHistoryOpen(v); if (!v) setHistoryConfig(null) }}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>历史版本</DialogTitle>
            <DialogDescription>{historyConfig?.config_name}</DialogDescription>
          </DialogHeader>
          <div className="space-y-3 max-h-[400px] overflow-y-auto">
            {historyList.length > 0 ? (
              historyList.map((h: ConfigHistoryItem) => (
                <div key={h.id} className="flex items-start justify-between rounded-lg border p-4">
                  <div className="space-y-1">
                    <div className="flex items-center gap-2">
                      <Badge variant="secondary">v{h.version}</Badge>
                      <span className="text-sm text-muted-foreground">{h.operator}</span>
                      <span className="text-xs text-muted-foreground">{h.created_at}</span>
                    </div>
                    {h.change_reason && <p className="text-sm">{h.change_reason}</p>}
                  </div>
                  <Button variant="outline" size="sm" onClick={() => {
                    setRollbackTarget(h)
                    setRollbackOpen(true)
                  }}>
                    <RotateCcw className="h-3 w-3 mr-1" />回滚到此版本
                  </Button>
                </div>
              ))
            ) : (
              <p className="text-center text-muted-foreground py-8">暂无历史版本</p>
            )}
          </div>
        </DialogContent>
      </Dialog>

      <Dialog open={rollbackOpen} onOpenChange={setRollbackOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>回滚配置</DialogTitle>
            <DialogDescription>将 {historyConfig?.config_name} 回滚到 v{rollbackTarget?.version}</DialogDescription>
          </DialogHeader>
          <div className="space-y-2">
            <Label>回滚原因</Label>
            <Input value={rollbackReason} onChange={(e) => setRollbackReason(e.target.value)} placeholder="可选" />
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setRollbackOpen(false)}>取消</Button>
            <Button onClick={() => rollbackMutate(
              { configId: Number(historyConfig!.id), historyId: Number(rollbackTarget!.id), data: { change_reason: rollbackReason || undefined } },
              { onSuccess: () => { invalidate(); setRollbackOpen(false); setRollbackReason("") } },
            )}>确认回滚</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
