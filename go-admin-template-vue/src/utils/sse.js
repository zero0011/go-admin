import useUserStore from '@/store/user'

// EventSource 封装：登录后连接、指数退避重连、登出关闭
// 注意：EventSource 无法自定义 Header，鉴权走 ?Authorization=<jwt>
// （后端 JwtMiddleware 支持从 Query 读取 token）

let es = null
let retryCount = 0
let reconnectTimer = null
let handlers = {}

const BASE_DELAY = 1000
const MAX_RETRY_DELAY = 30000

function buildUrl() {
    const uStore = useUserStore()
    const base = import.meta.env.VITE_APP_BASE_API
    return `${base}/message/stream?Authorization=${encodeURIComponent(uStore.token)}`
}

function cleanup() {
    if (es) {
        es.close()
        es = null
    }
}

function parseEvent(e) {
    try {
        return JSON.parse(e.data)
    } catch {
        return null
    }
}

function scheduleReconnect() {
    if (reconnectTimer) {
        return
    }
    const delay = Math.min(BASE_DELAY * Math.pow(2, retryCount), MAX_RETRY_DELAY)
    retryCount++
    reconnectTimer = setTimeout(() => {
        reconnectTimer = null
        connectSSE(handlers)
    }, delay)
}

// connectSSE 建立 SSE 连接（已连接或未登录时跳过）
export function connectSSE(h) {
    handlers = h || handlers
    const uStore = useUserStore()
    if (es || !uStore.token) {
        return
    }
    cleanup()
    es = new EventSource(buildUrl())

    es.onopen = () => {
        retryCount = 0
        if (handlers.onOpen) handlers.onOpen()
    }

    es.addEventListener('message', e => {
        if (handlers.onMessage) handlers.onMessage(parseEvent(e))
    })

    es.addEventListener('unread', e => {
        if (handlers.onUnread) handlers.onUnread(parseEvent(e))
    })

    es.onerror = () => {
        // 关闭后按指数退避手动重连，避免 401 等场景下浏览器无限快速重试
        cleanup()
        if (handlers.onError) handlers.onError()
        scheduleReconnect()
    }
}

// disconnectSSE 断开连接并取消重连
export function disconnectSSE() {
    cleanup()
    if (reconnectTimer) {
        clearTimeout(reconnectTimer)
        reconnectTimer = null
    }
    retryCount = 0
    handlers = {}
}
