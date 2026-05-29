import axios from 'axios'

const request = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api/v1',
  timeout: 10000,
})

request.interceptors.request.use(
  (config) => {
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

request.interceptors.response.use(
  (response) => {
    const res = response.data
    // 兼容 code: 0 和 code: 1000 都算成功
    if (res.code !== 0 && res.code !== 1000) {
      console.error('API Error:', res.message || res.msg)
      return Promise.reject(new Error(res.message || res.msg || 'Error'))
    }
    return res
  },
  (error) => {
    // 提取后端返回的错误信息
    const data = error?.response?.data
    const msg = data?.msg || data?.message || error.message || 'Request Error'
    console.error('Request Error:', msg)
    return Promise.reject(new Error(msg))
  }
)

export default request