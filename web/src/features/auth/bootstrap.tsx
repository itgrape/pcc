import { useEffect, useRef, type PropsWithChildren } from "react"

import { useAuthStore } from "./store"

export function AuthBootstrap({ children }: PropsWithChildren) {
  const initialize = useAuthStore((state) => state.initialize)
  const logout = useAuthStore((state) => state.logout)
  const initializationStarted = useRef(false)

  useEffect(() => {
    if (window.location.pathname !== "/callback" && !initializationStarted.current) {
      initializationStarted.current = true
      void initialize()
    }

    const handleUnauthorized = () => logout()
    window.addEventListener("pcc:unauthorized", handleUnauthorized)
    return () => window.removeEventListener("pcc:unauthorized", handleUnauthorized)
  }, [initialize, logout])

  return children
}
