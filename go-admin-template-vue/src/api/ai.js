// 使用独立的较长超时，因为 LLM 调用可能需要 30s+
import axios from 'axios'
const aiRequest = axios.create({
    baseURL: import.meta.env.VITE_APP_BASE_API,
    timeout: 60000
})
import useUserStore from '@/store/user'
aiRequest.interceptors.request.use(config => {
    const uStore = useUserStore()
    if (uStore.token) {
        config.headers['Authorization'] = uStore.token
    }
    return config
})
aiRequest.interceptors.response.use(
    response => {
        const { data } = response
        if (data.code !== 0) {
            return Promise.reject(data.message)
        }
        return data
    },
    error => Promise.reject(error)
)

export function text2sql(data) {
    return aiRequest({
        url: '/ai/text2sql',
        method: 'post',
        data
    })
}
