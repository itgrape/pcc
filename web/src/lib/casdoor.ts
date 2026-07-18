// Casdoor OAuth 配置
// 生产环境建议通过 VITE_ 环境变量注入
export const casdoorConfig = {
  serverUrl: import.meta.env.VITE_CASDOOR_URL || "https://door.casbin.com",
  clientId: import.meta.env.VITE_CASDOOR_CLIENT_ID || "",
  appName: import.meta.env.VITE_CASDOOR_APP_NAME || "app-pcc",
  organizationName: import.meta.env.VITE_CASDOOR_ORG || "built-in",
  redirectPath: "/callback",
}
