import type { PropsWithChildren } from "react"
import { QueryClientProvider } from "@tanstack/react-query"

import { AuthBootstrap } from "@/features/auth/bootstrap"
import { ThemeProvider } from "@/features/theme/theme-provider"
import { queryClient } from "@/shared/api/query-client"

export function AppProviders({ children }: PropsWithChildren) {
  return (
    <ThemeProvider>
      <QueryClientProvider client={queryClient}>
        <AuthBootstrap>{children}</AuthBootstrap>
      </QueryClientProvider>
    </ThemeProvider>
  )
}
