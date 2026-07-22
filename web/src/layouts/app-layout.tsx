import { useState } from "react"
import {
  ChevronLeft,
  ChevronRight,
  CloudCog,
  FolderKanban,
  Key,
  LayoutDashboard,
  LogOut,
} from "lucide-react"
import { Link, Outlet, useLocation } from "react-router-dom"

import { Avatar, AvatarFallback } from "@/components/ui/avatar"
import { Button } from "@/components/ui/button"
import { Separator } from "@/components/ui/separator"
import { useAuthStore } from "@/features/auth/store"
import { ThemeToggle } from "@/features/theme/theme-toggle"
import { cn } from "@/lib/utils"

const navItems = [
  { icon: LayoutDashboard, label: "仪表盘", path: "/" },
  { icon: FolderKanban, label: "命名空间", path: "/namespaces" },
  { icon: Key, label: "密钥管理", path: "/access-keys" },
]

export default function AppLayout() {
  const [collapsed, setCollapsed] = useState(false)
  const location = useLocation()
  const { user, logout } = useAuthStore()
  const currentPage =
    navItems.find((item) =>
      item.path === "/" ? location.pathname === "/" : location.pathname.startsWith(item.path),
    )?.label ?? "配置中心"

  return (
    <div className="surface-grid flex h-screen overflow-hidden bg-background">
      <aside
        className={cn(
          "relative hidden flex-col border-r border-white/5 bg-sidebar text-sidebar-foreground shadow-2xl transition-all duration-300 md:flex",
          collapsed ? "w-[76px]" : "w-[264px]",
        )}
      >
        <div className="flex h-[72px] items-center gap-3 border-b border-sidebar-ring/70 px-5">
          <div className="relative flex h-10 w-10 shrink-0 items-center justify-center rounded-2xl bg-gradient-to-br from-indigo-400 via-violet-500 to-fuchsia-500 text-white shadow-lg shadow-violet-950/40">
            <CloudCog className="h-5 w-5" />
            <span className="absolute -right-0.5 -top-0.5 h-2.5 w-2.5 rounded-full border-2 border-sidebar bg-emerald-400" />
          </div>
          {!collapsed && (
            <div className="min-w-0">
              <p className="font-semibold tracking-tight text-white">PCC Console</p>
              <p className="text-[11px] text-sidebar-foreground/45">Configuration Platform</p>
            </div>
          )}
        </div>

        <div className={cn("px-4 pt-6 pb-2 text-[10px] font-semibold uppercase tracking-[0.18em] text-sidebar-foreground/35", collapsed && "text-center")}>
          {collapsed ? "•••" : "Workspace"}
        </div>
        <nav className="flex-1 space-y-1.5 px-3">
          {navItems.map((item) => {
            const active =
              item.path === "/"
                ? location.pathname === "/"
                : location.pathname.startsWith(item.path)
            return (
              <Link
                key={item.path}
                to={item.path}
                className={cn(
                  "group flex items-center gap-3 rounded-xl px-3.5 py-3 text-sm font-medium transition-all",
                  active
                    ? "bg-white/10 text-white shadow-sm ring-1 ring-white/8"
                    : "text-sidebar-foreground/58 hover:bg-white/[0.06] hover:text-white",
                  collapsed && "justify-center px-0",
                )}
              >
                <item.icon className={cn("h-[18px] w-[18px] shrink-0 transition-colors", active ? "text-violet-300" : "text-sidebar-foreground/45 group-hover:text-violet-300")} />
                {!collapsed && <span>{item.label}</span>}
                {!collapsed && active && <span className="ml-auto h-1.5 w-1.5 rounded-full bg-violet-400 shadow-[0_0_10px] shadow-violet-400" />}
              </Link>
            )
          })}
        </nav>

        <div className="mx-3 mb-3 rounded-2xl border border-white/[0.06] bg-white/[0.035] p-3">
          {!collapsed && (
            <div className="mb-3 flex items-center gap-2 text-xs text-sidebar-foreground/55">
              <span className="relative flex h-2 w-2">
                <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-emerald-400 opacity-60" />
                <span className="relative inline-flex h-2 w-2 rounded-full bg-emerald-400" />
              </span>
              服务运行正常
            </div>
          )}
          <Separator className="mb-3 bg-sidebar-ring/70" />

          <div
            className={cn(
              "flex items-center gap-3",
              collapsed && "justify-center",
            )}
          >
            <Avatar className="h-9 w-9 border border-white/10">
              <AvatarFallback className="bg-gradient-to-br from-violet-500 to-indigo-500 text-xs font-semibold text-white">
                {(user?.displayName || user?.name)?.charAt(0).toUpperCase() || "U"}
              </AvatarFallback>
            </Avatar>
            {!collapsed && (
              <div className="flex-1 min-w-0">
                <p className="truncate text-sm font-medium text-white">
                  {user?.displayName || user?.name || "用户"}
                </p>
                <p className="truncate text-[11px] text-sidebar-foreground/40">{user?.email || "Administrator"}</p>
              </div>
            )}
            {!collapsed && (
              <Button
                variant="ghost"
                size="icon"
                aria-label="退出登录"
                className="h-8 w-8 text-sidebar-foreground/40 hover:bg-white/10 hover:text-white"
                onClick={logout}
              >
                <LogOut className="h-4 w-4" />
              </Button>
            )}
          </div>

          {collapsed && (
            <Button
              variant="ghost"
              size="icon"
              aria-label="退出登录"
              className="mx-auto mt-2 flex h-8 w-8 text-sidebar-foreground/40 hover:bg-white/10 hover:text-white"
              onClick={logout}
            >
              <LogOut className="h-4 w-4" />
            </Button>
          )}
        </div>

        <button
          type="button"
          aria-label={collapsed ? "展开侧边栏" : "收起侧边栏"}
          onClick={() => setCollapsed((value) => !value)}
          className="absolute -right-3.5 top-[86px] z-10 flex h-7 w-7 items-center justify-center rounded-full border bg-background text-muted-foreground shadow-md transition hover:text-foreground"
        >
          {collapsed ? (
            <ChevronRight className="h-3 w-3" />
          ) : (
            <ChevronLeft className="h-3 w-3" />
          )}
        </button>
      </aside>

      <div className="flex min-w-0 flex-1 flex-col">
        <header className="flex h-[72px] shrink-0 items-center justify-between border-b bg-background/75 px-5 backdrop-blur-xl md:px-8">
          <div>
            <p className="text-xs font-medium uppercase tracking-[0.16em] text-muted-foreground">PCC / Workspace</p>
            <h2 className="mt-0.5 text-lg font-semibold tracking-tight">{currentPage}</h2>
          </div>
          <div className="flex items-center gap-3">
            <div className="hidden items-center gap-2 rounded-xl border bg-card/70 px-3 py-2 text-xs text-muted-foreground shadow-sm sm:flex">
              <span className="h-2 w-2 rounded-full bg-emerald-500" />
              Production
            </div>
            <ThemeToggle />
          </div>
        </header>

        <main className="min-h-0 flex-1 overflow-auto pb-20 md:pb-0">
          <div className="mx-auto max-w-[1500px] p-5 md:p-8 lg:p-10">
            <Outlet />
          </div>
        </main>
      </div>

      <nav className="fixed inset-x-4 bottom-4 z-40 flex items-center justify-around rounded-2xl border bg-background/90 p-2 shadow-2xl backdrop-blur-xl md:hidden">
        {navItems.map((item) => {
          const active =
            item.path === "/"
              ? location.pathname === "/"
              : location.pathname.startsWith(item.path)
          return (
            <Link
              key={item.path}
              to={item.path}
              className={cn(
                "flex min-w-20 flex-col items-center gap-1 rounded-xl px-3 py-2 text-[11px] font-medium text-muted-foreground",
                active && "bg-primary/10 text-primary",
              )}
            >
              <item.icon className="h-4 w-4" />
              {item.label}
            </Link>
          )
        })}
      </nav>
    </div>
  )
}
