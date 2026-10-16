import { lazy, Suspense } from "react"
import { BrowserRouter, Route, Routes } from "react-router-dom"

import LoginPage from "@/features/auth/login-page"
import ProtectedRoute from "@/features/auth/protected-route"
import AppLayout from "@/layouts/app-layout"

const DashboardPage = lazy(() => import("@/pages/dashboard"))
const NamespacesPage = lazy(() => import("@/pages/namespaces"))
const ConfigsPage = lazy(() => import("@/pages/configs"))
const AccessKeysPage = lazy(() => import("@/pages/access-keys"))

function RouteFallback() {
  return (
    <div className="flex min-h-[50vh] items-center justify-center">
      <div className="flex items-center gap-3 rounded-2xl border bg-card/80 px-5 py-4 text-sm text-muted-foreground shadow-sm backdrop-blur">
        <span className="h-4 w-4 animate-spin rounded-full border-2 border-primary border-t-transparent" />
        正在加载工作空间…
      </div>
    </div>
  )
}

export function AppRouter() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route element={<ProtectedRoute />}>
          <Route element={<AppLayout />}>
            <Route path="/" element={<Suspense fallback={<RouteFallback />}><DashboardPage /></Suspense>} />
            <Route path="/namespaces" element={<Suspense fallback={<RouteFallback />}><NamespacesPage /></Suspense>} />
            <Route path="/namespaces/:namespaceId/configs" element={<Suspense fallback={<RouteFallback />}><ConfigsPage /></Suspense>} />
            <Route path="/access-keys" element={<Suspense fallback={<RouteFallback />}><AccessKeysPage /></Suspense>} />
          </Route>
        </Route>
      </Routes>
    </BrowserRouter>
  )
}
