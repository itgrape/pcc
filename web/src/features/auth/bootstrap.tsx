import { useEffect, useRef, type PropsWithChildren } from "react"

import { useAuthStore } from "./store"

export function AuthBootstrap({ children }: PropsWithChildren) {
  const initialize = useAuthStore((state) => state.initialize)
  const clearSession = useAuthStore((state) => state.clearSession)
  const initializationStarted = useRef(false)

  useEffect(() => {
    if (!initializationStarted.current) {
      initializationStarted.current = true
      void initialize()
    }

    const handleUnauthorized = () => clearSession()
    window.addEventListener("pcc:unauthorized", handleUnauthorized)
    return () => window.removeEventListener("pcc:unauthorized", handleUnauthorized)
  }, [initialize, clearSession])

  return children
}
