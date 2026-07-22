import { useState } from "react"
import { useQueryClient } from "@tanstack/react-query"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Card, CardContent } from "@/components/ui/card"
import {
  Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription,
  DialogFooter,
} from "@/components/ui/dialog"
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from "@/components/ui/select"
import { Badge } from "@/components/ui/badge"
import { Search, Pencil, Trash2, Shield, Copy, Check, KeyRound, Plus } from "lucide-react"
import {
  useListAccessKeys,
  useCreateAccessKey,
  useUpdateAccessKey,
  useDeleteAccessKey,
  useListKeyNamespaces,
  useRevokeKeyNamespace,
  grantKeyNamespace,
  listKeyNamespaces,
  revokeKeyNamespace,
  updateKeyNamespacePermission,
  getListAccessKeysQueryKey,
  getListKeyNamespacesQueryKey,
} from "@/lib/orval/access-keys/access-keys"
import { useListNamespaces } from "@/lib/orval/namespaces/namespaces"
import { apiId, unwrapData } from "@/shared/api/response"

interface AccessKeyItem {
  id: string
  access_key: string
  status: number
  description: string
  created_at: string
  updated_at: string
}

interface AccessKeyCreated {
  id: string
  access_key: string
  secret: string
  description: string
  status: number
}

interface AKNamespace {
  access_key_id: string
  namespace_id: string
  permission: number
  created_at: string
}

interface NamespaceItem {
  id: string
  namespace: string
  description: string
}

export default function AccessKeysPage() {
  const [page, setPage] = useState(1)
  const [keyword, setKeyword] = useState("")
  const [search, setSearch] = useState("")
  const [copied, setCopied] = useState("")
  const queryClient = useQueryClient()

  const [createOpen, setCreateOpen] = useState(false)
  const [editOpen, setEditOpen] = useState(false)
  const [editKey, setEditKey] = useState<AccessKeyItem | null>(null)
  const [grantOpen, setGrantOpen] = useState(false)
  const [grantKey, setGrantKey] = useState<AccessKeyItem | null>(null)
  const [detailOpen, setDetailOpen] = useState(false)
  const [detailKey, setDetailKey] = useState<AccessKeyItem | null>(null)

  const [createForm, setCreateForm] = useState({ description: "" })
  const [editForm, setEditForm] = useState({ description: "", status: "1" })
  const [grantForm, setGrantForm] = useState({ namespace_ids: [] as string[], permission: "2" })
  const [existingGrants, setExistingGrants] = useState<AKNamespace[]>([])
  const [grantLoading, setGrantLoading] = useState(false)
  const [grantSaving, setGrantSaving] = useState(false)
  const [grantError, setGrantError] = useState("")
  const [createdKey, setCreatedKey] = useState<AccessKeyCreated | null>(null)

  const params = { page, page_size: 20, keyword: search || undefined }
  const { data, isLoading } = useListAccessKeys(params)
  const { data: nsRes } = useListNamespaces({ page: 1, page_size: 100 })
  const { data: keyNsRes } = useListKeyNamespaces(
    apiId(detailKey?.id ?? "0"),
    { query: { enabled: Boolean(detailKey) } },
  )

  const invalidate = () => queryClient.invalidateQueries({ queryKey: getListAccessKeysQueryKey() })
  const invalidateKeyNamespaces = (keyId: number) =>
    queryClient.invalidateQueries({ queryKey: getListKeyNamespacesQueryKey(keyId) })

  const { mutate: createMutate } = useCreateAccessKey()
  const { mutate: updateMutate } = useUpdateAccessKey()
  const { mutate: deleteMutate } = useDeleteAccessKey()
  const { mutate: revokeMutate } = useRevokeKeyNamespace()

  const akData = data?.data as unknown as { data?: { list: unknown[]; total: number } } | undefined
  const list = (akData?.data?.list ?? []) as AccessKeyItem[]
  const total = akData?.data?.total ?? 0
  const allNamespaces = ((nsRes?.data as unknown as { data?: { list: NamespaceItem[] } })?.data?.list ?? []) as NamespaceItem[]
  const keyNss = ((keyNsRes?.data as unknown as { data?: AKNamespace[] })?.data ?? []) as AKNamespace[]

  const copyToClipboard = async (text: string, id: string) => {
    if (!text) return
    try {
      await navigator.clipboard.writeText(text)
    } catch {
      const textarea = document.createElement("textarea")
      textarea.value = text
      textarea.style.position = "fixed"
      textarea.style.opacity = "0"
      document.body.appendChild(textarea)
      textarea.select()
      document.execCommand("copy")
      textarea.remove()
    }
    setCopied(id)
    window.setTimeout(() => setCopied(""), 2000)
  }

  const openGrantDialog = async (key: AccessKeyItem) => {
    setGrantKey(key)
    setGrantOpen(true)
    setGrantLoading(true)
    setGrantError("")
    try {
      const response = await listKeyNamespaces(apiId(key.id))
      const grants = unwrapData<AKNamespace[]>(response)
      setExistingGrants(grants)
      setGrantForm({
        namespace_ids: grants.map((grant) => grant.namespace_id),
        permission: grants[0] ? String(grants[0].permission) : "2",
      })
    } catch {
      setGrantError("命名空间授权加载失败")
    } finally {
      setGrantLoading(false)
    }
  }

  const toggleGrantNamespace = (namespaceId: string) => {
    setGrantForm((current) => ({
      ...current,
      namespace_ids: current.namespace_ids.includes(namespaceId)
        ? current.namespace_ids.filter((id) => id !== namespaceId)
        : [...current.namespace_ids, namespaceId],
    }))
  }

  const saveNamespaceGrants = async () => {
    if (!grantKey) return
    setGrantSaving(true)
    setGrantError("")
    const selected = new Set(grantForm.namespace_ids)
    const existing = new Map(existingGrants.map((grant) => [grant.namespace_id, grant]))
    const permission = Number(grantForm.permission)

    try {
      const operations: Promise<unknown>[] = []
      for (const namespaceId of selected) {
        const current = existing.get(namespaceId)
        if (!current) {
          operations.push(grantKeyNamespace(apiId(grantKey.id), {
            namespace_id: apiId(namespaceId),
            permission,
          }))
        } else if (current.permission !== permission) {
          operations.push(
            updateKeyNamespacePermission(
              apiId(grantKey.id),
              apiId(namespaceId),
              { permission },
            ),
          )
        }
      }
      for (const grant of existingGrants) {
        if (!selected.has(grant.namespace_id)) {
          operations.push(
            revokeKeyNamespace(apiId(grantKey.id), apiId(grant.namespace_id)),
          )
        }
      }
      await Promise.all(operations)
      invalidate()
      invalidateKeyNamespaces(apiId(grantKey.id))
      setGrantOpen(false)
    } catch {
      setGrantError("授权保存失败，请检查命名空间是否已被删除")
    } finally {
      setGrantSaving(false)
    }
  }

  return (
    <div className="space-y-7">
      <div className="flex flex-col justify-between gap-5 sm:flex-row sm:items-end">
        <div>
          <div className="mb-3 inline-flex items-center gap-2 rounded-full border bg-card/70 px-3 py-1 text-xs font-medium text-muted-foreground shadow-sm">
            <KeyRound className="h-3.5 w-3.5 text-amber-500" />
            Credential Security
          </div>
          <h1 className="text-3xl font-semibold tracking-tight">访问密钥</h1>
          <p className="mt-2 text-sm text-muted-foreground">为应用创建独立凭证，并精确控制命名空间权限</p>
        </div>
        <Dialog open={createOpen} onOpenChange={(v) => { setCreateOpen(v); if (!v) setCreatedKey(null) }}>
          <DialogContent className="rounded-2xl">
            {createdKey ? (
              <>
                <DialogHeader>
                  <DialogTitle>密钥创建成功</DialogTitle>
                  <DialogDescription>请立即复制并安全保存密钥，关闭后不再显示</DialogDescription>
                </DialogHeader>
                <div className="space-y-3">
                  <div className="space-y-1">
                    <Label>Access Key</Label>
                    <div className="flex gap-2">
                      <Input readOnly value={createdKey.access_key} className="font-mono text-xs" />
                      <Button variant="outline" size="icon" onClick={() => void copyToClipboard(createdKey.access_key, "ak")}>
                        {copied === "ak" ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
                      </Button>
                    </div>
                  </div>
                  <div className="space-y-1">
                    <Label>Secret</Label>
                    <div className="flex gap-2">
                      <Input readOnly value={createdKey.secret} className="font-mono text-xs" />
                      <Button variant="outline" size="icon" onClick={() => void copyToClipboard(createdKey.secret, "secret")}>
                        {copied === "secret" ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
                      </Button>
                    </div>
                  </div>
                </div>
                <DialogFooter>
                  <Button onClick={() => { setCreateOpen(false); setCreatedKey(null) }}>我已安全保存</Button>
                </DialogFooter>
              </>
            ) : (
              <>
                <DialogHeader>
                  <DialogTitle>创建密钥</DialogTitle>
                  <DialogDescription>创建新的 API 访问密钥</DialogDescription>
                </DialogHeader>
                <div className="space-y-2">
                  <Label>描述</Label>
                  <Input value={createForm.description} onChange={(e) => setCreateForm({ description: e.target.value })} placeholder="例如：生产环境密钥" />
                </div>
                <DialogFooter>
                  <Button variant="outline" onClick={() => setCreateOpen(false)}>取消</Button>
                  <Button onClick={() => createMutate(
                    { data: { description: createForm.description || undefined } },
                    { onSuccess: (res) => {
                      setCreatedKey(unwrapData<AccessKeyCreated>(res))
                      setCreateForm({ description: "" })
                      invalidate()
                    } },
                  )}>创建</Button>
                </DialogFooter>
              </>
            )}
          </DialogContent>
        </Dialog>
        <Button className="rounded-xl shadow-lg shadow-primary/20" onClick={() => setCreateOpen(true)}>
          <Plus className="h-4 w-4" />
          创建访问密钥
        </Button>
      </div>

      <div className="flex flex-col gap-3 rounded-2xl border bg-card/65 p-3 shadow-sm backdrop-blur sm:flex-row sm:items-center">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input placeholder="按密钥或描述搜索..." className="h-10 border-0 bg-transparent pl-9 shadow-none focus-visible:ring-0" value={keyword} onChange={(e) => setKeyword(e.target.value)}
            onKeyDown={(e) => { if (e.key === "Enter") { setSearch(keyword); setPage(1) } }} />
        </div>
        <Button variant="secondary" className="rounded-xl" onClick={() => { setSearch(keyword); setPage(1) }}>搜索</Button>
      </div>

      <Card className="overflow-hidden border-border/60 bg-card/80 shadow-sm backdrop-blur">
        <CardContent className="p-0">
          <div className="overflow-x-auto">
          <table className="w-full min-w-[880px]">
            <thead>
              <tr className="border-b bg-muted/35 text-xs uppercase tracking-wide text-muted-foreground">
                <th className="text-left font-medium px-6 py-3">ID</th>
                <th className="text-left font-medium px-6 py-3">密钥</th>
                <th className="text-left font-medium px-6 py-3">状态</th>
                <th className="text-left font-medium px-6 py-3">描述</th>
                <th className="text-left font-medium px-6 py-3">创建时间</th>
                <th className="text-right font-medium px-6 py-3">操作</th>
              </tr>
            </thead>
            <tbody>
              {isLoading ? (
                <tr><td colSpan={6} className="px-6 py-12 text-center text-muted-foreground">加载中...</td></tr>
              ) : list.length === 0 ? (
                <tr><td colSpan={6} className="px-6 py-12 text-center text-muted-foreground">暂无数据</td></tr>
              ) : (
                list.map((ak: AccessKeyItem) => (
                  <tr key={ak.id} className="group border-b transition last:border-0 hover:bg-muted/35">
                    <td className="px-6 py-4 font-mono text-xs text-muted-foreground">#{ak.id}</td>
                    <td className="px-6 py-4">
                      <button onClick={() => { setDetailKey(ak); setDetailOpen(true) }} className="flex items-center gap-3 font-mono text-sm hover:text-primary">
                        <span className="flex h-9 w-9 items-center justify-center rounded-xl bg-amber-500/10 text-amber-600 dark:text-amber-400">
                          <KeyRound className="h-4 w-4" />
                        </span>
                        {ak.access_key.substring(0, 16)}...
                      </button>
                    </td>
                    <td className="px-6 py-3">
                      <Badge variant={ak.status === 1 ? "success" : "secondary"}>
                        {ak.status === 1 ? "启用" : "禁用"}
                      </Badge>
                    </td>
                    <td className="px-6 py-3 text-sm text-muted-foreground">{ak.description || "-"}</td>
                    <td className="px-6 py-3 text-sm text-muted-foreground">{ak.created_at}</td>
                    <td className="px-6 py-3 text-right">
                      <Button variant="ghost" size="icon" onClick={() => void openGrantDialog(ak)}>
                        <Shield className="h-4 w-4" />
                      </Button>
                      <Button variant="ghost" size="icon" onClick={() => { setEditKey(ak); setEditForm({ description: ak.description, status: String(ak.status) }); setEditOpen(true) }}>
                        <Pencil className="h-4 w-4" />
                      </Button>
                      <Button variant="ghost" size="icon" onClick={() => {
                        if (confirm("确定删除？")) deleteMutate({ keyId: apiId(ak.id) }, { onSuccess: () => invalidate() })
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

      <div className="flex items-center justify-between text-sm text-muted-foreground">
        <span>共 <strong className="font-medium text-foreground">{total}</strong> 个访问密钥</span>
        <div className="flex gap-2">
          <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => setPage(page - 1)}>上一页</Button>
          <Button variant="outline" size="sm" disabled={page * 20 >= total} onClick={() => setPage(page + 1)}>下一页</Button>
        </div>
      </div>

      <Dialog open={editOpen} onOpenChange={setEditOpen}>
        <DialogContent>
          <DialogHeader><DialogTitle>编辑密钥</DialogTitle><DialogDescription>{editKey?.access_key}</DialogDescription></DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2"><Label>描述</Label><Input value={editForm.description} onChange={(e) => setEditForm({ ...editForm, description: e.target.value })} /></div>
            <div className="space-y-2">
              <Label>状态</Label>
              <Select value={editForm.status} onValueChange={(v) => setEditForm({ ...editForm, status: v })}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="1">启用</SelectItem>
                  <SelectItem value="0">禁用</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setEditOpen(false)}>取消</Button>
            <Button onClick={() => updateMutate(
              { keyId: apiId(editKey!.id), data: { description: editForm.description || undefined, status: parseInt(editForm.status) } },
              { onSuccess: () => { invalidate(); setEditOpen(false) } },
            )}>保存</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={grantOpen} onOpenChange={(open) => {
        setGrantOpen(open)
        if (!open) setGrantKey(null)
      }}>
        <DialogContent className="max-w-xl">
          <DialogHeader>
            <DialogTitle>管理命名空间授权</DialogTitle>
            <DialogDescription>可同时为 {grantKey?.access_key} 绑定多个命名空间</DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label>统一权限</Label>
              <Select value={grantForm.permission} onValueChange={(v) => setGrantForm({ ...grantForm, permission: v })}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="1">只读</SelectItem>
                  <SelectItem value="2">读写</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <Label>命名空间（可多选）</Label>
                <span className="text-xs text-muted-foreground">已选 {grantForm.namespace_ids.length} 个</span>
              </div>
              <div className="max-h-72 space-y-2 overflow-y-auto rounded-xl border p-2">
                {grantLoading ? (
                  <p className="py-8 text-center text-sm text-muted-foreground">正在加载授权...</p>
                ) : allNamespaces.length === 0 ? (
                  <p className="py-8 text-center text-sm text-muted-foreground">暂无可授权的命名空间</p>
                ) : allNamespaces.map((namespace: NamespaceItem) => {
                  const selected = grantForm.namespace_ids.includes(namespace.id)
                  return (
                    <button
                      key={namespace.id}
                      type="button"
                      onClick={() => toggleGrantNamespace(namespace.id)}
                      className={`flex w-full items-center gap-3 rounded-lg border px-3 py-3 text-left transition ${
                        selected ? "border-primary/40 bg-primary/8" : "hover:bg-muted/60"
                      }`}
                    >
                      <span className={`flex h-5 w-5 items-center justify-center rounded border ${
                        selected ? "border-primary bg-primary text-primary-foreground" : "border-input"
                      }`}>
                        {selected && <Check className="h-3.5 w-3.5" />}
                      </span>
                      <span className="min-w-0">
                        <span className="block text-sm font-medium">{namespace.namespace}</span>
                        <span className="block truncate text-xs text-muted-foreground">{namespace.description || "暂无描述"}</span>
                      </span>
                    </button>
                  )
                })}
              </div>
            </div>
            {grantError && (
              <p className="rounded-lg bg-destructive/8 px-3 py-2 text-sm text-destructive">{grantError}</p>
            )}
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setGrantOpen(false)}>取消</Button>
            <Button onClick={() => void saveNamespaceGrants()} disabled={grantLoading || grantSaving}>
              {grantSaving ? "保存中..." : "保存授权"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={detailOpen} onOpenChange={(v) => { setDetailOpen(v); if (!v) setDetailKey(null) }}>
        <DialogContent className="max-w-lg">
          <DialogHeader>
            <DialogTitle>密钥详情</DialogTitle>
            <DialogDescription>{detailKey?.access_key}</DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div className="rounded-lg border p-3">
              <p className="text-xs text-muted-foreground mb-1">密钥</p>
              <p className="font-mono text-sm">{detailKey?.access_key}</p>
            </div>
            <div>
              <p className="text-sm font-medium mb-2">已授权的命名空间</p>
              {keyNss.length === 0 ? (
                <p className="text-sm text-muted-foreground">暂无授权</p>
              ) : (
                <div className="space-y-2">
                  {keyNss.map((akn: AKNamespace) => (
                    <div key={`${akn.access_key_id}-${akn.namespace_id}`} className="flex items-center justify-between rounded-lg border p-3">
                      <div className="flex items-center gap-2">
                        <Shield className="h-4 w-4 text-muted-foreground" />
                        <span className="text-sm">
                          {allNamespaces.find((namespace) => namespace.id === akn.namespace_id)?.namespace ?? `命名空间 #${akn.namespace_id}`}
                        </span>
                        <Badge variant="secondary">{akn.permission === 2 ? "读写" : "只读"}</Badge>
                      </div>
                      <Button variant="ghost" size="sm" className="text-destructive"
                        onClick={() => revokeMutate(
                          { keyId: apiId(detailKey!.id), namespaceId: apiId(akn.namespace_id) },
                          { onSuccess: () => { invalidate(); invalidateKeyNamespaces(apiId(detailKey!.id)) } },
                        )}>
                        解除
                      </Button>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  )
}
