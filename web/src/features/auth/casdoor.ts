import Sdk from "casdoor-js-sdk"

const config = {
  serverUrl: import.meta.env.VITE_CASDOOR_URL || "https://door.casbin.com",
  clientId: import.meta.env.VITE_CASDOOR_CLIENT_ID || "",
  appName: import.meta.env.VITE_CASDOOR_APP_NAME || "app-pcc",
  organizationName: import.meta.env.VITE_CASDOOR_ORG || "built-in",
  redirectPath: "/callback",
}

export const casdoor = new Sdk(config)
