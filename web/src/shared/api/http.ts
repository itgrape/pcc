import axios, {
  type AxiosError,
  type AxiosRequestConfig,
  type AxiosResponse,
} from "axios"

export const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || "",
})

apiClient.interceptors.request.use((config) => {
  const token = localStorage.getItem("casdoor_token")
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

apiClient.interceptors.response.use(
  (response) => response,
  (error: AxiosError) => {
    if (error.response?.status === 401) {
      localStorage.removeItem("casdoor_token")
      window.dispatchEvent(new Event("pcc:unauthorized"))
    }
    return Promise.reject(error)
  },
)

export function request<T>(
  config: AxiosRequestConfig,
  options?: AxiosRequestConfig,
): Promise<AxiosResponse<T>> {
  return apiClient.request<T>({
    ...config,
    ...options,
    headers: {
      ...config.headers,
      ...options?.headers,
    },
  })
}
