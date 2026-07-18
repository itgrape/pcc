import { useState } from "react"
import { useQueryClient } from "@tanstack/react-query"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Card, CardContent } from "@/components/ui/card"
import {
  Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription,
  DialogFooter, DialogTrigger,
} from "@/components/ui/dialog"
import { Plus, Search, Pencil, Trash2, ExternalLink } from "lucide-react"
import { Link } from "react-router-dom"
import {
  useListNamespaces,
  useCreateNamespace,
  useDeleteNamespace,
  useUpdateNamespace,
  getListNamespacesQueryKey,
} from "@/lib/orval/namespaces/namespaces"

interface Namespace {
  id: string
  namespace: string
  description: string
  created_at: string
  updated_at: string
}

interface PageData {
  list: Namespace[]
  total: number
  page: number
  page_size: number
}

export default function NamespacesPage() {
  const [page, setPage] = useState(1)
  const [keyword, setKeyword] = useState("")
  const [search, setSearch] = useState("")
  const queryClient = useQueryClient()

  const [open, setOpen] = useState(false)
  const [editOpen, setEditOpen] = useState(false)
  const [editNs, setEditNs] = useState<Namespace | null>(null)
  const [form, setForm] = useState({ namespace: "", description: "" })

  const params = { page, page_size: 20, keyword: search || undefined }
  const { data, isLoading } = useListNamespaces(params)

  const { mutate: createMutate, isPending: creating } = useCreateNamespace()
  const { mutate: updateMutate } = useUpdateNamespace()
  const { mutate: deleteMutate } = useDeleteNamespace()

  const invalidate = () => queryClient.invalidateQueries({ queryKey: getListNamespacesQueryKey() })

  const nsData = data?.data as unknown as { data?: PageData } | undefined
  const list = nsData?.data?.list ?? []
  const total = nsData?.data?.total ?? 0

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">命名空间</h1>
          <p className="text-muted-foreground">管理配置分组和权限边界</p>
        </div>
        <Dialog open={open} onOpenChange={setOpen}>
          <DialogTrigger asChild>
            <Button>
              <Plus className="h-4 w-4" />
              新建命名空间
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>新建命名空间</DialogTitle>
              <DialogDescription>创建一个新的配置命名空间</DialogDescription>
            </DialogHeader>
            <div className="space-y-4">
              <div className="space-y-2">
                <Label>命名空间名称</Label>
                <Input
                  placeholder="例如：production, staging"
                  value={form.namespace}
                  onChange={(e) => setForm({ ...form, namespace: e.target.value })}
                />
              </div>
              <div className="space-y-2">
                <Label>描述</Label>
                <Input
                  placeholder="可选描述"
                  value={form.description}
                  onChange={(e) => setForm({ ...form, description: e.target.value })}
                />
              </div>
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => setOpen(false)}>取消</Button>
              <Button
                onClick={() =>
                  createMutate(
                    { data: { namespace: form.namespace, description: form.description || undefined } },
                    { onSuccess: () => { invalidate(); setOpen(false); setForm({ namespace: "", description: "" }) } },
                  )
                }
                disabled={!form.namespace || creating}
              >
                创建
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>

      <div className="flex items-center gap-2">
        <div className="relative flex-1 max-w-sm">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="搜索命名空间..."
            className="pl-9"
            value={keyword}
            onChange={(e) => setKeyword(e.target.value)}
            onKeyDown={(e) => { if (e.key === "Enter") { setSearch(keyword); setPage(1) } }}
          />
        </div>
        <Button variant="secondary" size="sm" onClick={() => { setSearch(keyword); setPage(1) }}>搜索</Button>
      </div>

      <Card>
        <CardContent className="p-0">
          <table className="w-full">
            <thead>
              <tr className="border-b text-sm text-muted-foreground">
                <th className="text-left font-medium px-6 py-3">ID</th>
                <th className="text-left font-medium px-6 py-3">名称</th>
                <th className="text-left font-medium px-6 py-3">描述</th>
                <th className="text-left font-medium px-6 py-3">创建时间</th>
                <th className="text-right font-medium px-6 py-3">操作</th>
              </tr>
            </thead>
            <tbody>
              {isLoading ? (
                <tr><td colSpan={5} className="px-6 py-12 text-center text-muted-foreground">加载中...</td></tr>
              ) : list.length === 0 ? (
                <tr><td colSpan={5} className="px-6 py-12 text-center text-muted-foreground">暂无数据</td></tr>
              ) : (
                list.map((ns: Namespace) => (
                  <tr key={ns.id} className="border-b last:border-0 hover:bg-muted/50">
                    <td className="px-6 py-3 text-sm text-muted-foreground">{ns.id}</td>
                    <td className="px-6 py-3">
                      <Link to={`/namespaces/${ns.id}/configs`} className="font-medium hover:text-primary flex items-center gap-1">
                        {ns.namespace}
                        <ExternalLink className="h-3 w-3" />
                      </Link>
                    </td>
                    <td className="px-6 py-3 text-sm text-muted-foreground">{ns.description || "-"}</td>
                    <td className="px-6 py-3 text-sm text-muted-foreground">{ns.created_at}</td>
                    <td className="px-6 py-3 text-right">
                      <Button variant="ghost" size="icon" onClick={() => { setEditNs(ns); setEditOpen(true) }}>
                        <Pencil className="h-4 w-4" />
                      </Button>
                      <Button variant="ghost" size="icon" onClick={() => {
                        if (confirm("确定删除？")) deleteMutate(
                          { namespaceId: Number(ns.id) },
                          { onSuccess: () => invalidate() },
                        )
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

      <div className="flex items-center justify-between text-sm text-muted-foreground">
        <span>共 {total} 条</span>
        <div className="flex gap-2">
          <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => setPage(page - 1)}>上一页</Button>
          <Button variant="outline" size="sm" disabled={page * 20 >= total} onClick={() => setPage(page + 1)}>下一页</Button>
        </div>
      </div>

      <Dialog open={editOpen} onOpenChange={setEditOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>编辑命名空间</DialogTitle>
            <DialogDescription>{editNs?.namespace}</DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label>描述</Label>
              <Input
                value={editNs?.description ?? ""}
                onChange={(e) => setEditNs(editNs ? { ...editNs, description: e.target.value } : null)}
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setEditOpen(false)}>取消</Button>
            <Button onClick={() =>
              editNs && updateMutate(
                { namespaceId: Number(editNs.id), data: { description: editNs.description } },
                { onSuccess: () => { invalidate(); setEditOpen(false) } },
              )
            }>保存</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
