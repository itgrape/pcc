import axios from "axios"

axios.defaults.baseURL = import.meta.env.VITE_API_BASE_URL || ""

axios.interceptors.request.use((config) => {
  const token = localStorage.getItem("casdoor_token")
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

axios.interceptors.response.use(
  (res) => res,
  (err) => {
    if (err.response?.status === 401) {
      localStorage.removeItem("casdoor_token")
    }
    return Promise.reject(err)
  }
)
