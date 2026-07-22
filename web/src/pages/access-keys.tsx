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
  useGrantKeyNamespace,
  useRevokeKeyNamespace,
  getListAccessKeysQueryKey,
  getListKeyNamespacesQueryKey,
} from "@/lib/orval/access-keys/access-keys"
import { useListNamespaces } from "@/lib/orval/namespaces/namespaces"

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
  const [grantForm, setGrantForm] = useState({ namespace_id: "", permission: "2" })
  const [createdKey, setCreatedKey] = useState<AccessKeyCreated | null>(null)

  const params = { page, page_size: 20, keyword: search || undefined }
  const { data, isLoading } = useListAccessKeys(params)
  const { data: nsRes } = useListNamespaces({ page: 1, page_size: 100 })
  const { data: keyNsRes } = useListKeyNamespaces(Number(detailKey?.id))

  const invalidate = () => queryClient.invalidateQueries({ queryKey: getListAccessKeysQueryKey() })
  const invalidateKeyNamespaces = (keyId: number) =>
    queryClient.invalidateQueries({ queryKey: getListKeyNamespacesQueryKey(keyId) })

  const { mutate: createMutate } = useCreateAccessKey()
  const { mutate: updateMutate } = useUpdateAccessKey()
  const { mutate: deleteMutate } = useDeleteAccessKey()
  const { mutate: grantMutate } = useGrantKeyNamespace()
  const { mutate: revokeMutate } = useRevokeKeyNamespace()

  const akData = data?.data as unknown as { data?: { list: unknown[]; total: number } } | undefined
  const list = (akData?.data?.list ?? []) as AccessKeyItem[]
  const total = akData?.data?.total ?? 0
  const allNamespaces = ((nsRes?.data as unknown as { data?: { list: NamespaceItem[] } })?.data?.list ?? []) as NamespaceItem[]
  const keyNss = ((keyNsRes?.data as unknown as { data?: AKNamespace[] })?.data ?? []) as AKNamespace[]

  const copyToClipboard = (text: string, id: string) => {
    navigator.clipboard.writeText(text)
    setCopied(id)
    setTimeout(() => setCopied(""), 2000)
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
                      <Button variant="outline" size="icon" onClick={() => copyToClipboard(createdKey.access_key, "ak")}>
                        {copied === "ak" ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
                      </Button>
                    </div>
                  </div>
                  <div className="space-y-1">
                    <Label>Secret</Label>
                    <div className="flex gap-2">
                      <Input readOnly value={createdKey.secret} className="font-mono text-xs" />
                      <Button variant="outline" size="icon" onClick={() => copyToClipboard(createdKey.secret, "secret")}>
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
                    { onSuccess: (res) => { setCreatedKey(res?.data as unknown as AccessKeyCreated ?? null); invalidate() } },
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
                      <Button variant="ghost" size="icon" onClick={() => { setGrantKey(ak); setGrantOpen(true) }}>
                        <Shield className="h-4 w-4" />
                      </Button>
                      <Button variant="ghost" size="icon" onClick={() => { setEditKey(ak); setEditForm({ description: ak.description, status: String(ak.status) }); setEditOpen(true) }}>
                        <Pencil className="h-4 w-4" />
                      </Button>
                      <Button variant="ghost" size="icon" onClick={() => {
                        if (confirm("确定删除？")) deleteMutate({ keyId: Number(ak.id) }, { onSuccess: () => invalidate() })
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
              { keyId: Number(editKey!.id), data: { description: editForm.description || undefined, status: parseInt(editForm.status) } },
              { onSuccess: () => { invalidate(); setEditOpen(false) } },
            )}>保存</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={grantOpen} onOpenChange={setGrantOpen}>
        <DialogContent>
          <DialogHeader><DialogTitle>授权命名空间</DialogTitle><DialogDescription>为 {grantKey?.access_key} 添加命名空间授权</DialogDescription></DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label>命名空间</Label>
              <Select value={grantForm.namespace_id} onValueChange={(v) => setGrantForm({ ...grantForm, namespace_id: v })}>
                <SelectTrigger><SelectValue placeholder="选择命名空间" /></SelectTrigger>
                <SelectContent>
                  {allNamespaces.map((ns: NamespaceItem) => (
                    <SelectItem key={ns.id} value={ns.id}>{ns.namespace}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <Label>权限</Label>
              <Select value={grantForm.permission} onValueChange={(v) => setGrantForm({ ...grantForm, permission: v })}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="1">只读</SelectItem>
                  <SelectItem value="2">读写</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setGrantOpen(false)}>取消</Button>
            <Button onClick={() => grantMutate(
              { keyId: Number(grantKey!.id), data: { namespace_id: parseInt(grantForm.namespace_id), permission: parseInt(grantForm.permission) } },
              { onSuccess: () => { invalidate(); invalidateKeyNamespaces(Number(grantKey!.id)); setGrantOpen(false) } },
            )} disabled={!grantForm.namespace_id}>授权</Button>
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
                        <span className="text-sm">命名空间 #{akn.namespace_id}</span>
                        <Badge variant="secondary">{akn.permission === 2 ? "读写" : "只读"}</Badge>
                      </div>
                      <Button variant="ghost" size="sm" className="text-destructive"
                        onClick={() => revokeMutate(
                          { keyId: Number(detailKey!.id), namespaceId: Number(akn.namespace_id) },
                          { onSuccess: () => { invalidate(); invalidateKeyNamespaces(Number(detailKey!.id)) } },
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
