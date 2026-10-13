import { defineStore } from 'pinia'
import { getUnreadCount, markRead, markAllRead } from '@/api/message'
import { connectSSE, disconnectSSE } from '@/utils/sse'
import useUserStore from '@/store/user'

// SSE 断开期间的轮询降级间隔（doc.md NFR：EventSource 不可用时降级为轮询未读数）
const POLL_INTERVAL = 30000
let pollTimer = null

// 站内信 store：未读数徽标 + SSE 连接状态
// 新消息到达仅更新徽标（不做弹窗/声音/桌面通知）
export default defineStore('message', {
    state: () => ({
        unreadCount: 0,
        connected: false
    }),
    actions: {
        // 拉取最新未读数
        async fetchUnreadCount() {
            const res = await getUnreadCount()
            this.unreadCount = res.data.count
        },
        // 单条已读（幂等）
        async markRead(id) {
            const res = await markRead({ id })
            if (res.data && res.data.count > 0) {
                this.unreadCount = Math.max(0, this.unreadCount - res.data.count)
            }
            return res
        },
        // 一键已读（幂等）
        async markAllRead() {
            const res = await markAllRead()
            if (res.data && res.data.count > 0) {
                this.unreadCount = Math.max(0, this.unreadCount - res.data.count)
            }
            return res
        },
        // SSE 事件：新消息摘要
        onMessageEvent() {
            this.unreadCount++
        },
        // SSE 事件：服务端下发的最新未读数
        onUnreadEvent(payload) {
            if (payload && typeof payload.count === 'number') {
                this.unreadCount = payload.count
            }
        },
        // 建立 SSE 连接（登录后调用），并兜底开启轮询降级
        connect() {
            this.startPolling()
            connectSSE({
                onOpen: () => {
                    this.connected = true
                    // 重连补偿：连接成功后拉取最新未读数，保证离线期间消息不丢
                    this.fetchUnreadCount().catch(() => {})
                },
                onMessage: payload => this.onMessageEvent(payload),
                onUnread: payload => this.onUnreadEvent(payload),
                onError: () => {
                    this.connected = false
                }
            })
        },
        // 轮询降级：仅当 SSE 未连接且已登录时拉取，避免与实时推送重复请求
        startPolling() {
            if (pollTimer) {
                return
            }
            pollTimer = setInterval(() => {
                if (this.connected || (useUserStore().token) === '') {
                    return
                }
                this.fetchUnreadCount().catch(() => {})
            }, POLL_INTERVAL)
        },
        stopPolling() {
            if (pollTimer) {
                clearInterval(pollTimer)
                pollTimer = null
            }
        },
        // 断开 SSE 连接（登出/离开布局时调用）
        disconnect() {
            disconnectSSE()
            this.connected = false
            this.stopPolling()
        }
    }
})
