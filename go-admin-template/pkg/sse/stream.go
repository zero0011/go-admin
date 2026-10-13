package sse

import (
	"time"

	"github.com/gin-gonic/gin"
)

// heartbeatInterval 心跳间隔，防代理超时断连
const heartbeatInterval = 20 * time.Second

// Stream 建立 SSE 长连接并阻塞直到客户端断开
// onConnect 在连接注册后、进入事件循环前调用（可用于推送初始未读数）
func Stream(c *gin.Context, userID int, onConnect func(*Client)) {
	client := DefaultHub.Register(userID)
	defer DefaultHub.Unregister(userID, client)

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no") // 禁用 nginx 等代理缓冲

	if onConnect != nil {
		onConnect(client)
	}

	// 立即写出响应头，让 EventSource onopen 尽快触发
	c.Writer.Flush()

	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case ev := <-client.ch:
			c.SSEvent(ev.Name, ev.Data)
			c.Writer.Flush()
		case <-ticker.C:
			// 注释行心跳
			_, _ = c.Writer.WriteString(": ping\n\n")
			c.Writer.Flush()
		case <-c.Request.Context().Done():
			return
		}
	}
}
