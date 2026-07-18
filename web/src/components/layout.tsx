import { useState } from "react"
import { Link, useLocation, Outlet } from "react-router-dom"
import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"
import { Avatar, AvatarFallback } from "@/components/ui/avatar"
import {
  LayoutDashboard,
  FolderKanban,
  Key,
  LogOut,
  ChevronLeft,
  ChevronRight,
} from "lucide-react"
import { useAuthStore } from "@/stores/auth"
import { Separator } from "@/components/ui/separator"

const navItems = [
  { icon: LayoutDashboard, label: "仪表盘", path: "/" },
  { icon: FolderKanban, label: "命名空间", path: "/namespaces" },
  { icon: Key, label: "密钥管理", path: "/access-keys" },
]

export default function AppLayout() {
  const [collapsed, setCollapsed] = useState(false)
  const location = useLocation()
  const { user, logout } = useAuthStore()

  return (
    <div className="flex h-screen overflow-hidden bg-muted/30">
      <aside
        className={cn(
          "flex flex-col border-r bg-sidebar text-sidebar-foreground transition-all duration-300",
          collapsed ? "w-16" : "w-60"
        )}
      >
        <div className="flex h-14 items-center gap-2 border-b border-sidebar-ring px-4">
          <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary text-primary-foreground text-sm font-bold">
            P
          </div>
          {!collapsed && <span className="font-semibold tracking-tight">PCC</span>}
        </div>

        <nav className="flex-1 space-y-1 p-3">
          {navItems.map((item) => {
            const active = item.path === "/" ? location.pathname === "/" : location.pathname.startsWith(item.path)
            return (
              <Link
                key={item.path}
                to={item.path}
                className={cn(
                  "flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors",
                  active ? "bg-sidebar-accent text-sidebar-accent-foreground" : "text-sidebar-foreground/70 hover:bg-sidebar-accent hover:text-sidebar-accent-foreground"
                )}
              >
                <item.icon className="h-4 w-4 shrink-0" />
                {!collapsed && <span>{item.label}</span>}
              </Link>
            )
          })}
        </nav>

        <Separator className="bg-sidebar-ring" />

        <div className="p-3">
          <div className={cn("flex items-center gap-3 rounded-lg px-3 py-2", collapsed && "justify-center")}>
            <Avatar className="h-8 w-8">
              <AvatarFallback className="bg-sidebar-accent text-sidebar-accent-foreground text-xs">
                {user?.username?.charAt(0).toUpperCase() || "U"}
              </AvatarFallback>
            </Avatar>
            {!collapsed && <div className="flex-1 min-w-0"><p className="text-sm font-medium truncate">{user?.username || "用户"}</p></div>}
          </div>

          <Button
            variant="ghost"
            size="sm"
            className={cn("mt-1 w-full justify-start text-sidebar-foreground/70 hover:text-sidebar-accent-foreground", collapsed && "justify-center px-2")}
            onClick={logout}
          >
            <LogOut className="h-4 w-4 shrink-0" />
            {!collapsed && <span>退出登录</span>}
          </Button>
        </div>

        <button
          onClick={() => setCollapsed(!collapsed)}
          className="absolute -right-3 top-20 flex h-6 w-6 items-center justify-center rounded-full border bg-background shadow-sm hover:bg-accent"
        >
          {collapsed ? <ChevronRight className="h-3 w-3" /> : <ChevronLeft className="h-3 w-3" />}
        </button>
      </aside>

      <main className="flex-1 overflow-auto">
        <div className="mx-auto max-w-7xl p-6">
          <Outlet />
        </div>
      </main>
    </div>
  )
}
