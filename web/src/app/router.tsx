import { BrowserRouter, Route, Routes } from "react-router-dom"

import CallbackPage from "@/features/auth/callback-page"
import LoginPage from "@/features/auth/login-page"
import ProtectedRoute from "@/features/auth/protected-route"
import AppLayout from "@/layouts/app-layout"
import AccessKeysPage from "@/pages/access-keys"
import ConfigsPage from "@/pages/configs"
import DashboardPage from "@/pages/dashboard"
import NamespacesPage from "@/pages/namespaces"

export function AppRouter() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/callback" element={<CallbackPage />} />
        <Route element={<ProtectedRoute />}>
          <Route element={<AppLayout />}>
            <Route path="/" element={<DashboardPage />} />
            <Route path="/namespaces" element={<NamespacesPage />} />
            <Route path="/namespaces/:namespaceId/configs" element={<ConfigsPage />} />
            <Route path="/access-keys" element={<AccessKeysPage />} />
          </Route>
        </Route>
      </Routes>
    </BrowserRouter>
  )
}
