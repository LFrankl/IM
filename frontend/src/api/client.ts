import axios from 'axios'

const client = axios.create({
  // 开发环境通过 Vite proxy 转发，生产环境同源部署，无需指定 baseURL
  baseURL: import.meta.env.VITE_API_BASE_URL || '',
  timeout: 10000,
})

// 请求拦截：自动附加 token
client.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// 响应拦截：统一处理 401
client.interceptors.response.use(
  (res) => res,
  (err) => {
    if (err.response?.status === 401) {
      localStorage.removeItem('token')
      window.location.href = '/login'
    }
    return Promise.reject(err)
  },
)

export default client

// 统一响应类型
export interface ApiResponse<T = unknown> {
  code: number
  message: string
  data: T
}
