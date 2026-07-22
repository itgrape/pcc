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
          "relative hidden flex-col border-r bg-card/95 text-card-foreground shadow-sm backdrop-blur transition-all duration-300 md:flex",
          collapsed ? "w-[72px]" : "w-[248px]",
        )}
      >
        <div className="flex h-[72px] items-center gap-3 px-4">
          <div className="relative flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-primary text-primary-foreground shadow-sm">
            <CloudCog className="h-5 w-5" />
          </div>
          {!collapsed && (
            <div className="min-w-0">
              <p className="font-semibold tracking-tight">PCC Console</p>
              <p className="text-[11px] text-muted-foreground">Configuration Platform</p>
            </div>
          )}
        </div>

        <div className={cn("px-4 pb-2 pt-5 text-[10px] font-semibold uppercase tracking-[0.18em] text-muted-foreground/70", collapsed && "text-center")}>
          {collapsed ? "•••" : "Workspace"}
        </div>
        <nav className="flex-1 space-y-1 px-3">
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
                  "group flex items-center gap-3 rounded-xl px-3 py-2.5 text-sm font-medium transition-colors",
                  active
                    ? "bg-primary/10 text-primary"
                    : "text-muted-foreground hover:bg-muted hover:text-foreground",
                  collapsed && "justify-center px-0",
                )}
              >
                <item.icon className={cn("h-[18px] w-[18px] shrink-0", active ? "text-primary" : "text-muted-foreground group-hover:text-foreground")} />
                {!collapsed && <span>{item.label}</span>}
              </Link>
            )
          })}
        </nav>

        <div className="m-3 rounded-2xl bg-muted/65 p-2">
          <div
            className={cn(
              "flex items-center gap-2.5",
              collapsed && "justify-center",
            )}
          >
            <Avatar className="h-9 w-9 border bg-background">
              <AvatarFallback className="bg-primary/10 text-xs font-semibold text-primary">
                {(user?.displayName || user?.name)?.charAt(0).toUpperCase() || "U"}
              </AvatarFallback>
            </Avatar>
            {!collapsed && (
              <div className="flex-1 min-w-0">
                <p className="truncate text-sm font-medium">
                  {user?.displayName || user?.name || "用户"}
                </p>
                <p className="truncate text-[11px] text-muted-foreground">{user?.email || "Administrator"}</p>
              </div>
            )}
            {!collapsed && (
              <Button
                variant="ghost"
                size="icon"
                aria-label="退出登录"
                className="h-8 w-8 text-muted-foreground hover:bg-background hover:text-foreground"
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
              className="mx-auto mt-2 flex h-8 w-8 text-muted-foreground hover:bg-background hover:text-foreground"
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
          className="absolute -right-3 top-[86px] z-10 flex h-6 w-6 items-center justify-center rounded-full border bg-card text-muted-foreground shadow-sm transition hover:bg-accent hover:text-foreground"
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
