import type { PropsWithChildren } from "react"
import { QueryClientProvider } from "@tanstack/react-query"

import { AuthBootstrap } from "@/features/auth/bootstrap"
import { queryClient } from "@/shared/api/query-client"

export function AppProviders({ children }: PropsWithChildren) {
  return (
    <QueryClientProvider client={queryClient}>
      <AuthBootstrap>{children}</AuthBootstrap>
    </QueryClientProvider>
  )
}
