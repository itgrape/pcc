import { BrowserRouter, Routes, Route } from "react-router-dom"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import LoginPage from "@/pages/login"
import CallbackPage from "@/pages/callback"
import DashboardPage from "@/pages/dashboard"
import NamespacesPage from "@/pages/namespaces"
import ConfigsPage from "@/pages/configs"
import AccessKeysPage from "@/pages/access-keys"
import ProtectedRoute from "@/components/protected-route"
import AppLayout from "@/components/layout"

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      refetchOnWindowFocus: false,
      staleTime: 30000,
    },
  },
})

function App() {
  return (
    <QueryClientProvider client={queryClient}>
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
    </QueryClientProvider>
  )
}

export default App
