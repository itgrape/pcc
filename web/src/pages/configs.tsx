import { lazy, Suspense, useState } from "react"
import { useParams, Link } from "react-router-dom"
import { useQueryClient } from "@tanstack/react-query"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Card, CardContent } from "@/components/ui/card"
import {
  Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription,
  DialogFooter, DialogTrigger,
} from "@/components/ui/dialog"
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from "@/components/ui/select"
import { Badge } from "@/components/ui/badge"
import { Eye, Plus, Pencil, Trash2, History, RotateCcw } from "lucide-react"
import {
  getConfig,
  getConfigHistory,
  useListConfigs,
  useCreateConfig,
  useUpdateConfig,
  useDeleteConfig,
  useListConfigHistory,
  useRollbackConfig,
  getListConfigsQueryKey,
} from "@/lib/orval/configs/configs"
import { useGetNamespace } from "@/lib/orval/namespaces/namespaces"
import { apiId, unwrapData } from "@/shared/api/response"

const ConfigEditor = lazy(() =>
  import("@/features/configs/config-editor").then((module) => ({
    default: module.ConfigEditor,
  })),
)

function EditorFallback() {
  return (
    <div className="flex h-[360px] items-center justify-center rounded-2xl border bg-muted/30 text-sm text-muted-foreground">
      <span className="mr-2 h-4 w-4 animate-spin rounded-full border-2 border-primary border-t-transparent" />
      正在加载配置编辑器…
    </div>
  )
}

function ConfigContent({ content }: { content?: string }) {
  return (
    <pre className="max-h-[520px] overflow-auto whitespace-pre-wrap rounded-2xl border bg-muted/35 p-5 font-mono text-sm leading-6">
      {content || "（空配置）"}
    </pre>
  )
}

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
  const nsId = apiId(namespaceId ?? "0")
  const queryClient = useQueryClient()

  const [open, setOpen] = useState(false)
  const [editOpen, setEditOpen] = useState(false)
  const [editConfig, setEditConfig] = useState<ConfigItem | null>(null)
  const [viewOpen, setViewOpen] = useState(false)
  const [viewConfig, setViewConfig] = useState<ConfigItem | null>(null)
  const [historyOpen, setHistoryOpen] = useState(false)
  const [historyConfig, setHistoryConfig] = useState<ConfigItem | null>(null)
  const [historyDetailOpen, setHistoryDetailOpen] = useState(false)
  const [historyDetail, setHistoryDetail] = useState<ConfigHistoryItem | null>(null)
  const [rollbackOpen, setRollbackOpen] = useState(false)
  const [rollbackTarget, setRollbackTarget] = useState<ConfigHistoryItem | null>(null)
  const [form, setForm] = useState({ config_name: "", format: "1", content: "", description: "" })
  const [editForm, setEditForm] = useState({ content: "", description: "", change_reason: "" })
  const [rollbackReason, setRollbackReason] = useState("")
  const [formValid, setFormValid] = useState(true)
  const [editFormValid, setEditFormValid] = useState(true)
  const [actionError, setActionError] = useState("")

  const { data: nsRes } = useGetNamespace(nsId)
  const { data: configsRes, isLoading } = useListConfigs(nsId)
  const { data: historyRes, isLoading: historyLoading } = useListConfigHistory(
    apiId(historyConfig?.id ?? "0"),
    { query: { enabled: Boolean(historyConfig) } },
  )

  const invalidate = () => queryClient.invalidateQueries({ queryKey: getListConfigsQueryKey(nsId) })

  const { mutate: createMutate } = useCreateConfig()
  const { mutate: updateMutate } = useUpdateConfig()
  const { mutate: deleteMutate } = useDeleteConfig()
  const { mutate: rollbackMutate } = useRollbackConfig()

  const configs = (configsRes?.data as unknown as { data?: ConfigItem[] })?.data ?? []
  const namespaceName = (nsRes?.data as unknown as { data?: NamespaceDetail })?.data?.namespace ?? ""
  const historyList = (historyRes?.data as unknown as { data?: ConfigHistoryItem[] })?.data ?? []

  const loadConfigDetail = async (config: ConfigItem) => {
    const response = await getConfig(apiId(config.id))
    return unwrapData<ConfigItem>(response)
  }

  const handleViewConfig = async (config: ConfigItem) => {
    setActionError("")
    try {
      setViewConfig(await loadConfigDetail(config))
      setViewOpen(true)
    } catch {
      setActionError("配置详情加载失败，请稍后重试")
    }
  }

  const handleEditConfig = async (config: ConfigItem) => {
    setActionError("")
    try {
      const detail = await loadConfigDetail(config)
      setEditConfig(detail)
      setEditForm({ content: detail.content || "", description: detail.description, change_reason: "" })
      setEditOpen(true)
    } catch {
      setActionError("配置详情加载失败，无法编辑")
    }
  }

  const handleViewHistory = async (history: ConfigHistoryItem) => {
    setActionError("")
    try {
      const response = await getConfigHistory(apiId(history.id))
      setHistoryDetail(unwrapData<ConfigHistoryItem>(response))
      setHistoryDetailOpen(true)
    } catch {
      setActionError("历史版本加载失败，请稍后重试")
    }
  }

  return (
    <div className="space-y-7">
      <div className="flex flex-col justify-between gap-5 sm:flex-row sm:items-end">
        <div>
          <div className="mb-3 flex items-center gap-2 text-xs font-medium text-muted-foreground">
            <Link to="/namespaces" className="rounded-lg px-2 py-1 transition hover:bg-muted hover:text-foreground">命名空间</Link>
            <span>/</span>
            <span className="rounded-lg bg-primary/8 px-2 py-1 text-primary">{namespaceName}</span>
          </div>
          <h1 className="text-3xl font-semibold tracking-tight">配置管理</h1>
          <p className="mt-2 text-sm text-muted-foreground">编辑、校验并追踪 {namespaceName} 下的所有配置</p>
        </div>
        <Dialog open={open} onOpenChange={setOpen}>
          <DialogTrigger asChild>
            <Button className="rounded-xl shadow-lg shadow-primary/20">
              <Plus className="h-4 w-4" />新建配置
            </Button>
          </DialogTrigger>
          <DialogContent className="max-w-5xl">
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
                <Suspense fallback={<EditorFallback />}>
                  <ConfigEditor
                    value={form.content}
                    onChange={(content) => setForm({ ...form, content })}
                    format={Number(form.format)}
                    onValidityChange={setFormValid}
                  />
                </Suspense>
              </div>
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => setOpen(false)}>取消</Button>
              <Button onClick={() => createMutate(
                { namespaceId: nsId, data: { config_name: form.config_name, format: parseInt(form.format), content: form.content, description: form.description || undefined } },
                { onSuccess: () => { invalidate(); setOpen(false); setForm({ config_name: "", format: "1", content: "", description: "" }) } },
              )} disabled={!form.config_name || !form.content || !formValid}>创建</Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>

      {actionError && (
        <div className="rounded-xl border border-destructive/20 bg-destructive/5 px-4 py-3 text-sm text-destructive">
          {actionError}
        </div>
      )}

      <Card className="overflow-hidden border-border/60 bg-card/80 shadow-sm backdrop-blur">
        <CardContent className="p-0">
          <div className="overflow-x-auto">
          <table className="w-full min-w-[900px]">
            <thead>
              <tr className="border-b bg-muted/35 text-xs uppercase tracking-wide text-muted-foreground">
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
                  <tr key={cfg.id} className="group border-b transition last:border-0 hover:bg-muted/35">
                    <td className="px-6 py-4">
                      <div className="flex items-center gap-3">
                        <span className="flex h-9 w-9 items-center justify-center rounded-xl bg-primary/8 font-mono text-xs font-semibold text-primary">
                          {"{ }"}
                        </span>
                        <span className="font-medium">{cfg.config_name}</span>
                      </div>
                    </td>
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
                      <Button variant="ghost" size="icon" aria-label="查看配置" onClick={() => void handleViewConfig(cfg)}>
                        <Eye className="h-4 w-4" />
                      </Button>
                      <Button variant="ghost" size="icon" aria-label="编辑配置" onClick={() => void handleEditConfig(cfg)}>
                        <Pencil className="h-4 w-4" />
                      </Button>
                      <Button variant="ghost" size="icon" aria-label="查看历史" onClick={() => { setHistoryConfig(cfg); setHistoryOpen(true) }}>
                        <History className="h-4 w-4" />
                      </Button>
                      <Button variant="ghost" size="icon" onClick={() => {
                        if (confirm("确定删除？")) deleteMutate({ configId: apiId(cfg.id) }, { onSuccess: () => invalidate() })
                      }}>
                        <Trash2 className="h-4 w-4 text-destructive" />
                      </Button>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
          </div>
        </CardContent>
      </Card>

      <Dialog open={viewOpen} onOpenChange={setViewOpen}>
        <DialogContent className="max-w-4xl">
          <DialogHeader>
            <DialogTitle>{viewConfig?.config_name}</DialogTitle>
            <DialogDescription>
              {formatLabels[viewConfig?.format ?? 0] || "未知格式"} · v{viewConfig?.version}
            </DialogDescription>
          </DialogHeader>
          <ConfigContent content={viewConfig?.content} />
        </DialogContent>
      </Dialog>

      <Dialog open={editOpen} onOpenChange={setEditOpen}>
        <DialogContent className="max-w-5xl">
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
              <Suspense fallback={<EditorFallback />}>
                <ConfigEditor
                  value={editForm.content}
                  onChange={(content) => setEditForm({ ...editForm, content })}
                  format={editConfig?.format ?? 3}
                  onValidityChange={setEditFormValid}
                />
              </Suspense>
            </div>
            <div className="space-y-2">
              <Label>变更原因</Label>
              <Input value={editForm.change_reason} onChange={(e) => setEditForm({ ...editForm, change_reason: e.target.value })} placeholder="可选" />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setEditOpen(false)}>取消</Button>
            <Button onClick={() => updateMutate(
              { configId: apiId(editConfig!.id), data: editForm },
              { onSuccess: () => { invalidate(); setEditOpen(false) } },
            )} disabled={!editFormValid}>保存</Button>
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
            {historyLoading ? (
              <p className="py-8 text-center text-muted-foreground">正在加载历史版本...</p>
            ) : historyList.length > 0 ? (
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
                  <div className="flex gap-2">
                    <Button variant="ghost" size="sm" onClick={() => void handleViewHistory(h)}>
                      <Eye className="mr-1 h-3.5 w-3.5" />查看
                    </Button>
                    <Button variant="outline" size="sm" onClick={() => {
                      setRollbackTarget(h)
                      setRollbackOpen(true)
                    }}>
                      <RotateCcw className="h-3 w-3 mr-1" />回滚到此版本
                    </Button>
                  </div>
                </div>
              ))
            ) : (
              <p className="text-center text-muted-foreground py-8">暂无历史版本</p>
            )}
          </div>
        </DialogContent>
      </Dialog>

      <Dialog open={historyDetailOpen} onOpenChange={setHistoryDetailOpen}>
        <DialogContent className="max-w-4xl">
          <DialogHeader>
            <DialogTitle>历史版本 v{historyDetail?.version}</DialogTitle>
            <DialogDescription>
              {historyDetail?.operator || "未知操作人"} · {historyDetail?.created_at}
            </DialogDescription>
          </DialogHeader>
          <ConfigContent content={historyDetail?.content} />
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
              { configId: apiId(historyConfig!.id), historyId: apiId(rollbackTarget!.id), data: { change_reason: rollbackReason || undefined } },
              { onSuccess: () => { invalidate(); setRollbackOpen(false); setRollbackReason("") } },
            )}>确认回滚</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
